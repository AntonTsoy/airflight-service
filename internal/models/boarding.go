package models

type BoardingPass struct {
	TicketNo   string `gorm:"column:ticket_no;primaryKey" json:"ticket_no"`
	FlightID   uint   `gorm:"column:flight_id;primaryKey" json:"flight_id"`
	BoardingNo int    `gorm:"column:boarding_no" json:"boarding_no"`
	SeatNo     string `gorm:"column:seat_no" json:"seat_no"`
}
