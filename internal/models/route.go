package models

import (
	"time"
)

type Route struct {
	FlightNo           string    `json:"flight_no"`
	DepartureAirport   string    `json:"departure_airport"`
	ArrivalAirport     string    `json:"arrival_airport"`
	ScheduledDeparture time.Time `json:"scheduled_departure"`
	ScheduledArrival   time.Time `json:"scheduled_arrival"`
}
