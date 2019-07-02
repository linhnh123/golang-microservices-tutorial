package model

type AccountEvent struct {
	ID        string `json:"" gorm:"primary_key"`
	AccountID string `json:"-" gorm:"index"` // Don't serialize + index which is very important for performance.
	EventName string `json:"eventName"`
	Created   string `json:"created"`
}

type AccountImage struct {
	ID       string `json:"id" gorm:"primary_key"`
	URL      string `json:"url"`
	ServedBy string `json:"servedBy"`
}

type AccountData struct {
	ID     string         `json:"" gorm:"primary_key"`
	Name   string         `json:"name"`
	Events []AccountEvent `json:"events" gorm:"ForeignKey:AccountID"`
}
