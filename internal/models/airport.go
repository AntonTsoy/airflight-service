package models

type Airport struct {
	AirportCode string `json:"airport_code"`
	AirportName string `json:"airport_name"`
	City        string `json:"city"`
	Timezone    string `json:"timezone"`
}
