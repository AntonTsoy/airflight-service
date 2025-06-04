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
	"time"

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

type Flight struct {
	FlightID           uint      `json:"flight_id" gorm:"column:flight_id;primaryKey"`
	FlightNo           string    `json:"flight_no" gorm:"column:flight_no"`
	ScheduledArrival   time.Time `json:"scheduled_arrival" gorm:"column:scheduled_arrival"`
	ScheduledDeparture time.Time `json:"scheduled_departure" gorm:"column:scheduled_departure"`
	ArrivalAirport     string    `json:"arrival_airport" gorm:"column:arrival_airport"`
	DepartureAirport   string    `json:"departure_airport" gorm:"column:departure_airport"`
}

type FlightSchedule struct {
	DayOfWeek     string `json:"day_of_week"`
	TimeOfArrival string `json:"time_of_arrival"`
	FlightNo      string `json:"flight_no"`
	OriginAirport string `json:"origin_airport"`
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
// @Param city query string false "Airport city"
// @Success 200 {array} Airport
// @Failure 500 {object} map[string]string
// @Router /airports [get]
func getAirports(w http.ResponseWriter, r *http.Request) {
	var airports []Airport
	result := db.Distinct()
	city := r.URL.Query().Get("city")
	if city != "" {
		result = result.Where("city LIKE ?", city)
	}
	result = result.Find(&airports)

	if result.Error != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(airports)
}

// @Summary Get inbound schedule for an airport
// @Description Retrieves the inbound flight schedule for a specified airport
// @Tags airports
// @Produce json
// @Param airport_code path string true "Airport code"
// @Success 200 {array} FlightSchedule
// @Failure 400 {string} ErrorResponse "Missing or invalid airport code"
// @Failure 500 {object} map[string]string
// @Router /airports/inbound-schedule/{airport_code} [get]
func getInboundScheduleAirport(w http.ResponseWriter, r *http.Request) {
	airportCode := chi.URLParam(r, "airport_code")
	if airportCode == "" {
		http.Error(w, "Missing airport code parameter", http.StatusBadRequest)
		return
	}

	var flights []Flight
	if err := db.Where("arrival_airport = ?", airportCode).Find(&flights).Error; err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	schedules := make([]FlightSchedule, 0, len(flights))
	for _, flight := range flights {
		schedule := FlightSchedule{
			DayOfWeek:     flight.ScheduledArrival.Weekday().String(),
			TimeOfArrival: flight.ScheduledArrival.Format("15:04"),
			FlightNo:      flight.FlightNo,
			OriginAirport: flight.DepartureAirport,
		}
		schedules = append(schedules, schedule)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(schedules)
}

// @Summary Get outbound schedule for an airport
// @Description Retrieves the outbound flight schedule for a specified airport
// @Tags airports
// @Produce json
// @Param airport_code path string true "Airport code"
// @Success 200 {array} FlightSchedule
// @Failure 400 {string} ErrorResponse "Missing or invalid airport code"
// @Failure 500 {object} map[string]string
// @Router /airports/outbound-schedule/{airport_code} [get]
func getOutboundScheduleAirport(w http.ResponseWriter, r *http.Request) {
	airportCode := chi.URLParam(r, "airport_code")
	if airportCode == "" {
		http.Error(w, "Missing airport code parameter", http.StatusBadRequest)
		return
	}

	var flights []Flight
	if err := db.Where("departure_airport = ?", airportCode).Find(&flights).Error; err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	schedules := make([]FlightSchedule, 0, len(flights))
	for _, flight := range flights {
		schedule := FlightSchedule{
			DayOfWeek:     flight.ScheduledDeparture.Weekday().String(),
			TimeOfArrival: flight.ScheduledDeparture.Format("15:04"),
			FlightNo:      flight.FlightNo,
			OriginAirport: flight.ArrivalAirport,
		}
		schedules = append(schedules, schedule)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(schedules)
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
	r.Use(enableCORS)
	r.Use(middleware.Logger)

	r.Get("/airports", getAirports)
	r.Get("/airports/inbound-schedule/{airport_code}", getInboundScheduleAirport)
	r.Get("/airports/outbound-schedule/{airport_code}", getOutboundScheduleAirport)
	r.Get("/cities", getCities)
	r.Get("/swagger/*", httpSwagger.Handler(httpSwagger.URL("/swagger/doc.json")))

	fmt.Printf("Listening on http://%s/swagger/\n", config.ListenAddr)
	http.ListenAndServe(config.ListenAddr, r)
}
