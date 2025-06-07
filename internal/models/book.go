package models

type Book struct {
	GUID           string `gorm:"column:guid;primaryKey" json:"guid"`
	FlightID       uint   `gorm:"column:flight_id" json:"flight_id"`
	FareConditions string `gorm:"column:fare_conditions" json:"fare_conditions"`
	TicketNo       string `gorm:"column:ticket_no" json:"ticket_no"`
	Passanger      string `gorm:"column:passanger" json:"passanger"`
}

type BookingRequest struct {
	Passanger      string `json:"passanger"`
	FareConditions string `json:"fare_conditions"`
	FlightIDs      []uint `json:"flight_ids"`
}
