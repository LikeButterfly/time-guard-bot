package bot

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"time-guard-bot/internal/helpers"
	"time-guard-bot/internal/models"
	"time-guard-bot/internal/storage/redis"
)

// FIXME передавать сюда context?
// Starts a task timer
func (b *Bot) startTaskTimer(chatID int64, taskID string, duration time.Duration) {
	timer := time.AfterFunc(duration, func() {
		b.handleTaskTimeout(chatID, taskID)
	})

	b.timersMx.Lock()
	timerKey := fmt.Sprintf("%d:%s", chatID, taskID)
	b.timers[timerKey] = timer
	b.timersMx.Unlock()
}

// Handles a task timeout
func (b *Bot) handleTaskTimeout(chatID int64, taskID string) {
	// Remove timer from map
	b.timersMx.Lock()
	timerKey := fmt.Sprintf("%d:%s", chatID, taskID)
	delete(b.timers, timerKey)
	b.timersMx.Unlock()

	// Create context
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Get task
	_, err := b.storage.GetTask(ctx, chatID, taskID)
	if err != nil {
		log.Printf("Failed to get task on timeout: %v", err)
		return
	}

	// Get active task to find the original message ID
	activeTask, err := b.storage.GetActiveTask(ctx, chatID, taskID)
	if err != nil {
		log.Printf("Failed to get active task on timeout: %v", err)
	} else if activeTask.BotResponseID > 0 {
		// Удаляем inline keyboard
		emptyMarkup := tgbotapi.InlineKeyboardMarkup{
			InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{},
		}

		editMsg := tgbotapi.NewEditMessageReplyMarkup(chatID, activeTask.BotResponseID, emptyMarkup)
		if _, err := b.api.Send(editMsg); err != nil {
			log.Printf("Failed to remove keyboard from original message: %v", err)
		}
	}

	if err := b.storage.EndTask(ctx, chatID, taskID); err != nil {
		log.Printf("Failed to end task on timeout: %v", err)
		return
	}

	text := "Time has expired! How's it going?"
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyToMessageID = activeTask.MessageID

	if _, err := b.api.Send(msg); err != nil {
		log.Printf("Failed to send task timeout message: %v", err)
	}
}

// Handles the command like /{time} {name}
func (b *Bot) handleTimeCommand(ctx context.Context, message *tgbotapi.Message, duration int, taskName string) error {
	task, err := b.storage.GetTaskByName(ctx, message.Chat.ID, taskName)
	if err != nil {
		if errors.Is(err, redis.ErrNotFound) {
			return b.sendErrorMessage(
				message.Chat.ID,
				message.MessageID,
				fmt.Sprintf("Task *%s* not found", taskName),
			)
		}

		return fmt.Errorf("failed to get task: %w", err)
	}

	// Check if task is locked
	if task.IsLocked {
		errMsg := "Task is locked"
		if task.LockReason != "" {
			errMsg += fmt.Sprintf(". Reason: %s", task.LockReason)
		}

		return b.sendErrorMessage(message.Chat.ID, message.MessageID, errMsg)
	}

	// Проверяем, что задача не в работе
	if task.OwnerID != 0 {
		remaining := task.TimeRemaining()
		remainingMin := remaining / 60
		remainingSec := remaining % 60

		if task.OwnerID == message.From.ID {
			errMsg := fmt.Sprintf("You're already working on this task. %d:%02d remaining", remainingMin, remainingSec)
			return b.sendErrorMessage(message.Chat.ID, message.MessageID, errMsg)
		}

		errMsg := fmt.Sprintf("Another user is currently working on the task. %d:%02d remaining", remainingMin, remainingSec)

		return b.sendErrorMessage(message.Chat.ID, message.MessageID, errMsg)
	}

	// Check if user has too many active tasks
	count, err := b.storage.GetCountUserActiveTasks(ctx, message.Chat.ID, message.From.ID)
	if err != nil {
		return err
	}

	if count >= helpers.MaxTasksPerUser {
		return b.sendErrorMessage(
			message.Chat.ID,
			message.MessageID,
			fmt.Sprintf("You've reached the maximum number of active tasks (%d)", 4),
		)
	}

	// Do timer
	startTime := time.Now()
	endTime := startTime.Add(time.Duration(duration) * time.Minute)

	// Update task
	task.OwnerID = message.From.ID
	task.StartTime = startTime
	task.EndTime = endTime
	task.Duration = duration
	task.MessageID = message.MessageID

	var responseText string
	if duration == 1 {
		responseText = fmt.Sprintf("Timer started for %d minute", duration)
	} else {
		responseText = fmt.Sprintf("Timer started for %d minutes", duration)
	}

	replyMsg := tgbotapi.NewMessage(message.Chat.ID, responseText)
	replyMsg.ReplyToMessageID = message.MessageID

	sentMsg, err := b.api.Send(replyMsg)
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	task.BotResponseID = sentMsg.MessageID

	// Create active task
	activeTask := &models.ActiveTask{
		TaskID:        task.ID,
		UserID:        int64(message.From.ID),
		ChatID:        message.Chat.ID,
		StartTime:     startTime,
		EndTime:       endTime,
		Duration:      duration,
		MessageID:     message.MessageID,
		BotResponseID: sentMsg.MessageID,
	}

	// Start task in storage
	err = b.storage.StartTask(ctx, activeTask)
	if err != nil {
		return fmt.Errorf("failed to start task: %w", err)
	}

	// Add hourglass button
	hourglassButton := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⌛", fmt.Sprintf("check_time:%s:%d", task.ID, message.From.ID)),
		),
	)
	editMsg := tgbotapi.NewEditMessageReplyMarkup(message.Chat.ID, sentMsg.MessageID, hourglassButton)

	sentMsg, err = b.api.Send(editMsg)
	if err != nil {
		log.Printf("Failed to add inline keyboard: %v", err)
	}

	// Start timer
	b.startTaskTimer(message.Chat.ID, task.ID, time.Duration(duration)*time.Minute)

	return nil
}

// Handles the /cancel [name] command
// Cancels a running timer for a task. If name is provided, cancels that specific task
// If no name is provided, cancels the last task the user started
func (b *Bot) HandleCancelCommand(ctx context.Context, message *tgbotapi.Message, args []string) error {
	var taskName string
	if len(args) > 0 {
		taskName = args[0]
	}

	// Get user's active tasks
	activeTasks, err := b.storage.GetUserActiveTasks(ctx, message.Chat.ID, int64(message.From.ID))
	if err != nil {
		return fmt.Errorf("failed to get user's active tasks: %w", err)
	}

	if len(activeTasks) == 0 {
		return b.sendErrorMessage(
			message.Chat.ID,
			message.MessageID,
			"You don't have any active tasks",
		)
	}

	var taskToCancel *models.ActiveTask

	// If task name is provided, find that specific task
	if taskName != "" {
		// Find the task by name
		task, err := b.storage.GetTaskByName(ctx, message.Chat.ID, taskName)
		if err != nil {
			if errors.Is(err, redis.ErrNotFound) {
				return b.sendErrorMessage(
					message.Chat.ID,
					message.MessageID,
					fmt.Sprintf("Task *%s* not found", taskName),
				)
			}

			return fmt.Errorf("failed to get task: %w", err)
		}

		// Check if this task is active for the user
		for _, activeTask := range activeTasks {
			if activeTask.TaskID == task.ID {
				taskToCancel = activeTask
				break
			}
		}

		if taskToCancel == nil {
			return b.sendErrorMessage(
				message.Chat.ID,
				message.MessageID,
				fmt.Sprintf("You don't have an active timer for task *%s*", taskName),
			)
		}
	} else {
		// No task name provided, use the last started task
		// Sort tasks by start time, descending
		var lastTask *models.ActiveTask

		var latestStartTime time.Time

		for _, task := range activeTasks {
			if lastTask == nil || task.StartTime.After(latestStartTime) {
				lastTask = task
				latestStartTime = task.StartTime
			}
		}

		taskToCancel = lastTask
	}

	// Cancel the timer
	b.timersMx.Lock()
	timerKey := fmt.Sprintf("%d:%s", message.Chat.ID, taskToCancel.TaskID)

	timer, exists := b.timers[timerKey]
	if exists {
		timer.Stop()
		delete(b.timers, timerKey)
	}
	b.timersMx.Unlock()

	// Get task
	task, err := b.storage.GetTask(ctx, message.Chat.ID, taskToCancel.TaskID)
	if err != nil {
		return fmt.Errorf("failed to get task: %w", err)
	}

	if taskToCancel.BotResponseID > 0 {
		// Удаляем inline keyboard
		emptyMarkup := tgbotapi.InlineKeyboardMarkup{
			InlineKeyboard: [][]tgbotapi.InlineKeyboardButton{},
		}

		editMsg := tgbotapi.NewEditMessageReplyMarkup(message.Chat.ID, taskToCancel.BotResponseID, emptyMarkup)
		if _, err := b.api.Send(editMsg); err != nil {
			log.Printf("Failed to remove keyboard from original message: %v", err)
		}
	}

	if err := b.storage.EndTask(ctx, message.Chat.ID, taskToCancel.TaskID); err != nil {
		return fmt.Errorf("failed to end task: %w", err)
	}

	text := fmt.Sprintf("Timer for task *%s* has been cancelled", task.Name)
	replyMsg := tgbotapi.NewMessage(message.Chat.ID, text)
	replyMsg.ReplyToMessageID = task.MessageID
	replyMsg.ParseMode = tgbotapi.ModeMarkdown

	if _, err := b.api.Send(replyMsg); err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	return nil
}
