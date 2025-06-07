package models

import (
	"time"
)

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

type TicketFlight struct {
	TicketNo       string `gorm:"column:ticket_no;primaryKey" json:"ticket_no"`
	FlightID       uint   `gorm:"column:flight_id" json:"flight_id"`
	FareConditions string `gorm:"column:fare_conditions" json:"fare_conditions"`
}
