package models

type Aircraft struct {
	AircraftCode string `json:"id" gorm:"primaryKey"`
	Model        string `json:"model"`
	Range        uint   `json:"range"`
}
