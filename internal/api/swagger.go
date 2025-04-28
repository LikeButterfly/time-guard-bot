package api

import (
	"net/http"

	httpSwagger "github.com/swaggo/http-swagger"
	_ "github.com/swaggo/swag"

	_ "time-guard-bot/docs/swagger"
)

// Adds Swagger documentation routes to the provided ServeMux
func RegisterSwaggerRoutes(mux *http.ServeMux) {
	// Add route for Swagger UI
	mux.Handle("/swagger/", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
		httpSwagger.DeepLinking(true),
		httpSwagger.DocExpansion("list"),
	))
}
