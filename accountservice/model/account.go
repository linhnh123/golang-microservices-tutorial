package model

import (
	"github.com/linhnh123/golang-microservices-tutorial/common/model"
)

type Quote struct {
	Text     string `json:"quote"`
	ServedBy string `json:"ipAddress"`
	Language string `json:"language"`
}

type Account struct {
	Id            string               `json:"id"`
	Name          string               `json:"name"`
	ServedBy      string               `json:"servedBy"`
	Quote         Quote                `json:"quote"`
	ImageUrl      string               `json:"imageUrl"`
	ImageData     model.AccountImage   `json:"imageData"`
	AccountEvents []model.AccountEvent `json:"accountEvents"`
}
