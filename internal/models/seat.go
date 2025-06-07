package models

type Seat struct {
	AircraftCode   string `gorm:"column:aircraft_code;primaryKey" json:"aircraft_code"`
	SeatNo         string `gorm:"column:seat_no;primaryKey" json:"seat_no"`
	FareConditions string `gorm:"column:fare_conditions" json:"fare_conditions"`
}
