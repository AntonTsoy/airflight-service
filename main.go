// @title Airflight Service API
// @version 1.0
// @host localhost:8000
// @BasePath /
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	_ "github.com/AntonTsoy/airflight-service/docs" // Импортируем сгенерированную документацию
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger" // Добавляем http-swagger
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Aircraft struct {
	AircraftCode string `json:"id" gorm:"primaryKey"`
	Model        string `json:"model"`
	Range        int    `json:"range"`
}

type Airport struct {
	AirportCode string `json:"airport_code"`
	AirportName string `json:"airport_name"`
	City        string `json:"city"`
	Timezone    string `json:"timezone"`
}

var db *gorm.DB

func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// @Summary Get all cities
// @Description Retrieve a list of all cities from the database
// @Tags cities
// @Accept json
// @Produce json
// @Success 200 {array} string
// @Failure 500 {object} map[string]string
// @Router /cities [get]
func getCities(w http.ResponseWriter, r *http.Request) {
	var cities []string
	result := db.Model(&Airport{}).Distinct("city").Pluck("city", &cities)
	if result.Error != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cities)
}

// @Summary Get all airports
// @Description Retrieve a list of all airports from the database
// @Tags airports
// @Accept json
// @Produce json
// @Success 200 {array} Airport
// @Failure 500 {object} map[string]string
// @Router /airports [get]
func getAirports(w http.ResponseWriter, r *http.Request) {
	var airports []Airport
	result := db.Distinct().Find(&airports)
	if result.Error != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(airports)
}

func main() {
	config, err := LoadConfig()
	if err != nil {
		panic(err)
	}

	db, err = gorm.Open(postgres.Open(config.DatabaseDSN), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect to database:", err)
	}

	r := chi.NewRouter()
	// Добавляем CORS middleware перед другими middleware
	r.Use(enableCORS)
	r.Use(middleware.Logger)

	r.Get("/airports", getAirports)
	r.Get("/cities", getCities)
	r.Get("/swagger/*", httpSwagger.Handler(httpSwagger.URL("/swagger/doc.json")))

	fmt.Printf("Listening on http://%s/swagger/\n", config.ListenAddr)
	http.ListenAndServe(config.ListenAddr, r)
}
