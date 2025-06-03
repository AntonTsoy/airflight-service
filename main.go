// @title Airflight Service API
// @version 1.0
// @host localhost:8000
// @BasePath /
package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	_ "github.com/AntonTsoy/airflight-service/docs" // Импортируем сгенерированную документацию
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger" // Добавляем http-swagger
)

type Airport struct {
	ID   uint   `json:"id" gorm:"primaryKey"`
	Code string `json:"code" gorm:"uniqueIndex"`
	Name string `json:"name"`
}

// addAirport godoc
// @Summary Create a new airport
// @Description Creates a new airport with provided code and name
// @Tags airports
// @Accept json
// @Produce json
// @Param airport body Airport true "Airport data"
// @Success 201 {object} Airport
// @Failure 400 {string} string "Invalid input"
// @Router /airports [post]
func addAirport(w http.ResponseWriter, r *http.Request) {
	var airport Airport
	if err := json.NewDecoder(r.Body).Decode(&airport); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(airport)
}

func main() {
	config, err := LoadConfig()
	if err != nil {
		panic(err)
	}

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Post("/airports", addAirport)

	r.Get("/swagger/*", httpSwagger.Handler(httpSwagger.URL("/swagger/doc.json")))

	fmt.Printf("Listening on http://%s/swagger/\n", config.ListenAddr)
	http.ListenAndServe(config.ListenAddr, r)
}
