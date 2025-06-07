// @title Airflight Service API
// @version 1.0
// @host localhost:8000
// @BasePath /
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	_ "github.com/AntonTsoy/airflight-service/docs"
	"github.com/AntonTsoy/airflight-service/internal/config"
)

type Aircraft struct {
	AircraftCode string `json:"id" gorm:"primaryKey"`
	Model        string `json:"model"`
	Range        uint   `json:"range"`
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

type Seat struct {
	AircraftCode   string `gorm:"column:aircraft_code;primaryKey" json:"aircraft_code"`
	SeatNo         string `gorm:"column:seat_no;primaryKey" json:"seat_no"`
	FareConditions string `gorm:"column:fare_conditions" json:"fare_conditions"`
}

type BoardingPass struct {
	TicketNo   string `gorm:"column:ticket_no;primaryKey" json:"ticket_no"`
	FlightID   uint   `gorm:"column:flight_id;primaryKey" json:"flight_id"`
	BoardingNo int    `gorm:"column:boarding_no" json:"boarding_no"`
	SeatNo     string `gorm:"column:seat_no" json:"seat_no"`
}

type Route struct {
	FlightNo           string    `json:"flight_no"`
	DepartureAirport   string    `json:"departure_airport"`
	ArrivalAirport     string    `json:"arrival_airport"`
	ScheduledDeparture time.Time `json:"scheduled_departure"`
	ScheduledArrival   time.Time `json:"scheduled_arrival"`
}

type BookingRequest struct {
	Passanger      string `json:"passanger"`
	FareConditions string `json:"fare_conditions"`
	FlightIDs      []uint `json:"flight_ids"`
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
	guid := chi.URLParam(r, "guid")
	if guid == "" {
		http.Error(w, "Missing guid parameter", http.StatusBadRequest)
		return
	}

	defer r.Body.Close()
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

// @Summary Check-in for a flight
// @Description Assigns a seat for a booked flight using a GUID
// @Tags bookings
// @Accept json
// @Produce json
// @Param guid path string true "Booking GUID"
// @Param flight_id path uint true "Flight ID"
// @Success 200 {object} BoardingPass "Boarding pass details"
// @Failure 400 {string} ErrorResponse "Invalid input"
// @Failure 404 {string} ErrorResponse "Booking or seat not found"
// @Failure 500 {string} ErrorResponse "Internal server error"
// @Router /bookings/{guid}/check-in/{flight_id} [put]
func checkIn(w http.ResponseWriter, r *http.Request) {
	guid := chi.URLParam(r, "guid")
	flight_id, err := strconv.ParseUint(chi.URLParam(r, "flight_id"), 10, 0)
	if guid == "" || err != nil {
		http.Error(w, "GUID and Fligth ID are required", http.StatusBadRequest)
		return
	}
	reqFligthId := uint(flight_id)

	var boardingPass BoardingPass
	err = db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("ticket_no IN (?) AND flight_id = ?",
			tx.Table("books").Select("ticket_no").Where("guid = ? AND flight_id = ?", guid, reqFligthId),
			reqFligthId).First(&boardingPass).Error; err == nil {
			return nil // посадочный талон уже существует, возвращаем его
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("failed to check existing boarding pass: %v", err)
		}

		var book Book
		if err := tx.Where("guid = ? AND flight_id = ?", guid, reqFligthId).First(&book).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("booking not found for GUID %s and flight ID %d", guid, reqFligthId)
			}
			return fmt.Errorf("failed to find booking: %v", err)
		}

		var seat Seat
		subQuery := tx.Table("boarding_passes").Select("seat_no").Where("flight_id = ?", reqFligthId)
		if err := tx.Table("flights f").
			Select("s.seat_no").
			Joins("JOIN seats s ON s.aircraft_code = f.aircraft_code").
			Where("f.flight_id = ? AND s.fare_conditions = ?", reqFligthId, book.FareConditions).
			Where("s.seat_no NOT IN (?)", subQuery).
			Limit(1).
			Find(&seat).Error; err != nil {
			return fmt.Errorf("failed to find available seat: %v", err)
		}

		if seat.SeatNo == "" {
			return fmt.Errorf("no available seats for fare condition %s on flight %d", book.FareConditions, reqFligthId)
		}

		var maxBoardingNo struct{ Max int }
		tx.Table("boarding_passes").
			Select("COALESCE(MAX(boarding_no), 0) as max").
			Where("flight_id = ?", reqFligthId).
			Scan(&maxBoardingNo)

		boardingPass = BoardingPass{
			TicketNo:   book.TicketNo,
			FlightID:   reqFligthId,
			BoardingNo: maxBoardingNo.Max + 1,
			SeatNo:     seat.SeatNo,
		}
		if err := tx.Create(&boardingPass).Error; err != nil {
			return fmt.Errorf("failed to create boarding pass: %v", err)
		}

		return nil
	})

	if err != nil {
		var status = 400
		if strings.Contains(err.Error(), "booking not found") || strings.Contains(err.Error(), "no available seats") {
			status = http.StatusNotFound
		} else {
			status = http.StatusInternalServerError
		}
		http.Error(w, err.Error(), status)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(boardingPass)
}

// @Summary Get routes between two points
// @Description Lists routes connecting two points (airport or city) with specified filters
// @Tags routes
// @Produce json
// @Param from query string true "Departure point (airport code or city)"
// @Param to query string true "Arrival point (airport code or city)"
// @Param departure_date query string true "Departure date (YYYY-MM-DD)"
// @Param booking_class query string true "Booking class (Economy, Comfort, Business)"
// @Param connections query int false "Number of connections (0, 1, 2, 3); default 0"
// @Success 200 {array} Route "List of routes"
// @Failure 400 {string} ErrorResponse "Invalid input"
// @Failure 500 {string} ErrorResponse "Internal server error"
// @Router /routes [get]
func getRoutes(w http.ResponseWriter, r *http.Request) {
	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")
	departureDateStr := r.URL.Query().Get("departure_date")
	bookingClass := r.URL.Query().Get("booking_class")
	connectionsStr := r.URL.Query().Get("connections")

	if from == "" || to == "" || departureDateStr == "" || bookingClass == "" {
		http.Error(w, "From, to, departure_date, and booking_class are required", http.StatusBadRequest)
		return
	}
	validClasses := map[string]bool{"Economy": true, "EconomySec": true, "Comfort": true, "Business": true}
	if !validClasses[bookingClass] {
		http.Error(w, "Invalid booking class!", http.StatusBadRequest)
		return
	}
	connections := 0
	if connectionsStr != "" {
		if c, err := strconv.Atoi(connectionsStr); err == nil && c >= 0 {
			connections = c
		} else {
			http.Error(w, "Connections must be 0, 1, 2, or greeter", http.StatusBadRequest)
			return
		}
	}

	departureDate, err := time.Parse("2006-01-02", departureDateStr)
	if err != nil {
		http.Error(w, "Invalid departure date format. Use YYYY-MM-DD", http.StatusBadRequest)
		return
	}
	nextDate := departureDate.Add(24 * time.Hour)

	var fromAirports, toAirports []Airport
	if err := db.Where("city = ? OR airport_code = ?", from, from).Find(&fromAirports).Error; err != nil {
		http.Error(w, "Failed to fetch 'from' airports", http.StatusInternalServerError)
		return
	}
	if err := db.Where("city = ? OR airport_code = ?", to, to).Find(&toAirports).Error; err != nil {
		http.Error(w, "Failed to fetch 'to' airports", http.StatusInternalServerError)
		return
	}
	if len(fromAirports) == 0 || len(toAirports) == 0 {
		http.Error(w, "No airports found for given points", http.StatusNotFound)
		return
	}

	var routes []Route
	fromCodes := make([]string, len(fromAirports))
	for i, airport := range fromAirports {
		fromCodes[i] = airport.AirportCode
	}
	toCodes := make([]string, len(toAirports))
	for i, airport := range toAirports {
		toCodes[i] = airport.AirportCode
	}
	if connections == 0 {
		var flights []Flight
		if err := db.Where("departure_airport IN ? AND arrival_airport IN ? AND scheduled_departure BETWEEN ? AND ?",
			fromCodes, toCodes, departureDate, nextDate).Find(&flights).Error; err != nil {
			http.Error(w, "Failed to fetch flights", http.StatusInternalServerError)
			return
		}
		for _, flight := range flights {
			routes = append(routes, Route{
				FlightNo:           flight.FlightNo,
				DepartureAirport:   flight.DepartureAirport,
				ArrivalAirport:     flight.ArrivalAirport,
				ScheduledDeparture: flight.ScheduledDeparture,
				ScheduledArrival:   flight.ScheduledArrival,
			})
		}
	}

	if connections >= 1 {
		var connectingFlights []Flight
		if err := db.Raw(`
            SELECT f1.flight_no, f1.departure_airport, f2.arrival_airport, f1.scheduled_departure, f2.scheduled_arrival
            FROM flights f1
            JOIN flights f2 ON f1.arrival_airport = f2.departure_airport
            WHERE f1.departure_airport IN ? AND f2.arrival_airport IN ?
            AND f1.scheduled_departure BETWEEN ? AND ?
            AND f2.scheduled_departure > f1.scheduled_arrival
            AND f2.scheduled_departure < f1.scheduled_arrival + INTERVAL '24 hours'`,
			fromCodes, toCodes, departureDate, nextDate).Scan(&connectingFlights).Error; err != nil {
			http.Error(w, "Failed to fetch flights", http.StatusInternalServerError)
			return
		}
		for _, flight := range connectingFlights {
			routes = append(routes, Route{
				FlightNo:           flight.FlightNo,
				DepartureAirport:   flight.DepartureAirport,
				ArrivalAirport:     flight.ArrivalAirport,
				ScheduledDeparture: flight.ScheduledDeparture,
				ScheduledArrival:   flight.ScheduledArrival,
			})
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(routes)
}

func main() {
	config, err := config.Load()
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
	r.Get("/routes", getRoutes)
	r.Put("/bookings/{guid}", bookRoute)
	r.Put("/bookings/{guid}/check-in/{flight_id}", checkIn)
	r.Get("/swagger/*", httpSwagger.Handler(httpSwagger.URL("/swagger/doc.json")))

	fmt.Printf("Listening on http://%s/swagger/\n", config.ListenAddr)
	http.ListenAndServe(config.ListenAddr, r)
}
