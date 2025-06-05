// @title Airflight Service API
// @version 1.0
// @host localhost:8000
// @BasePath /
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
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

type BookingRequest struct {
	Passanger      string `json:"passanger"`
	FareConditions string `json:"fare_conditions"`
	FlightIDs      []uint `json:"flight_ids"`
}

type Book struct {
	GUID           string `gorm:"column:guid;primaryKey" json:"guid"`
	FlightID       uint   `gorm:"column:flight_id" json:"flight_id"`
	FareConditions string `gorm:"column:fare_conditions" json:"fare_conditions"`
	TicketNo       string `gorm:"column:ticket_no" json:"ticket_no"`
	Passanger      string `gorm:"column:passanger" json:"passanger"`
}

type TicketFlight struct {
	TicketNo       string `gorm:"column:ticket_no;primaryKey" json:"ticket_no"`
	FlightID       uint   `gorm:"column:flight_id" json:"flight_id"`
	FareConditions string `gorm:"column:fare_conditions" json:"fare_conditions"`
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

func generateTicketNo() string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 13)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
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
// @Router /airports/{airport_code}/inbound-schedule [get]
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
// @Router /airports/{airport_code}/outbound-schedule [get]
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

// @Summary Book a route
// @Description Idempotent booking of flights with a GUID
// @Tags bookings
// @Accept json
// @Produce json
// @Param guid path string true "GUID"
// @Param booking body BookingRequest true "Booking data"
// @Success 200 {array} TicketFlight "Existing or new tickets"
// @Failure 400 {string}  map[string]string
// @Failure 500 {string}  map[string]string
// @Router /bookings/{guid} [put]
func bookRoute(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	guid := chi.URLParam(r, "guid")
	if guid == "" {
		http.Error(w, "Missing guid parameter", http.StatusBadRequest)
		return
	}
	var req BookingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Failed to decode input", http.StatusBadRequest)
		return
	}

	validClasses := map[string]bool{"Economy": true, "Comfort": true, "Business": true, "EconomySec": true}
	if !validClasses[req.FareConditions] {
		http.Error(w, "Invalid fare condition. Must be 'Economy', 'Comfort', 'Business', or 'EconomySec'", http.StatusBadRequest)
		return
	}

	var tickets []TicketFlight
	err := db.Transaction(func(tx *gorm.DB) error {
		var existingBooks []Book
		if err := tx.Where("guid = ?", guid).Find(&existingBooks).Error; err != nil {
			return err
		}

		if len(existingBooks) > 0 {
			var ticketNos []string
			for _, book := range existingBooks {
				ticketNos = append(ticketNos, book.TicketNo)
			}
			if err := tx.Where("ticket_no IN ?", ticketNos).Find(&tickets).Error; err != nil {
				return err
			}
			return nil
		}

		for _, flightID := range req.FlightIDs {
			ticketNo := generateTicketNo()
			book := Book{
				GUID:           guid,
				FlightID:       flightID,
				FareConditions: req.FareConditions,
				TicketNo:       ticketNo,
				Passanger:      req.Passanger,
			}
			ticketFlight := TicketFlight{
				TicketNo:       ticketNo,
				FlightID:       flightID,
				FareConditions: req.FareConditions,
			}

			if err := tx.Create(&book).Error; err != nil {
				return err
			}
			if err := tx.Create(&ticketFlight).Error; err != nil {
				return err
			}
			tickets = append(tickets, ticketFlight)
		}

		return nil
	})

	if err != nil {
		http.Error(w, "Failed to process booking in DB", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(tickets)
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
	r.Get("/airports/{airport_code}/inbound-schedule", getInboundScheduleAirport)
	r.Get("/airports/{airport_code}/outbound-schedule", getOutboundScheduleAirport)
	r.Get("/cities", getCities)
	r.Put("/bookings/{guid}", bookRoute)
	r.Get("/swagger/*", httpSwagger.Handler(httpSwagger.URL("/swagger/doc.json")))

	fmt.Printf("Listening on http://%s/swagger/\n", config.ListenAddr)
	http.ListenAndServe(config.ListenAddr, r)
}
