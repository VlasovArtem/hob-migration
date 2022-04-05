package model

import (
	"github.com/google/uuid"
	"time"
)

type CreateHouseRequest struct {
	Name        string
	CountryCode string
	City        string
	StreetLine1 string
	StreetLine2 string
	UserId      string
	GroupIds    []uuid.UUID
}

type CreateHouseBatchRequest struct {
	Houses []CreateHouseRequest
}

type HouseDto struct {
	Id          uuid.UUID
	Name        string
	CountryCode string
	City        string
	StreetLine1 string
	StreetLine2 string
	UserId      uuid.UUID
}

type GroupDto struct {
	Id      uuid.UUID
	Name    string
	OwnerId uuid.UUID
}

type CreateGroupRequest struct {
	Name    string
	OwnerId string
}

type CreateGroupBatchRequest struct {
	Groups []CreateGroupRequest
}

type CreateIncomeRequest struct {
	Name        string
	Description string
	Date        string
	Sum         float32
	HouseId     string
	GroupIds    []string
}

type CreateIncomeBatchRequest struct {
	Incomes []CreateIncomeRequest
}

type IncomeDto struct {
	Id          uuid.UUID
	Name        string
	Description string
	Date        time.Time
	Sum         float32
	HouseId     uuid.UUID
}

type CreatePaymentRequest struct {
	Name        string
	Description string
	HouseId     string
	UserId      string
	ProviderId  string
	Date        string
	Sum         float32
}

type CreatePaymentBatchRequest struct {
	Payments []CreatePaymentRequest
}

type PaymentDto struct {
	Id          uuid.UUID
	Name        string
	Description string
	HouseId     uuid.UUID
	UserId      uuid.UUID
	ProviderId  uuid.UUID
	Date        time.Time
	Sum         float32
}
