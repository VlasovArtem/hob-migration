package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/VlasovArtem/hob-migration/src/config"
	"github.com/VlasovArtem/hob-migration/src/model"
	"github.com/rs/zerolog/log"
	"io/ioutil"
	"net/http"
)

type HobClient struct {
	config *config.CMDConfig
}

func NewHobClient(config *config.CMDConfig) *HobClient {
	return &HobClient{
		config: config,
	}
}

func (h *HobClient) HealthCheck() error {
	get, err := http.Get(h.config.HobURL + "/api/v1/health")

	if err != nil {
		return err
	}

	if get.StatusCode != 200 {
		return errors.New("hob server is not available")
	}

	return nil
}

func (h *HobClient) CreateHouse(request model.CreateHouseRequest) (model.HouseDto, error) {
	requestBytes, err := json.Marshal(request)

	if err != nil {
		return model.HouseDto{}, err
	}

	return ReadBody[model.HouseDto](http.Post(h.config.HobURL+"/api/v1/houses", "application/json", bytes.NewReader(requestBytes)))
}

func (h *HobClient) CreateGroupBatch(request model.CreateGroupBatchRequest) ([]model.GroupDto, error) {
	requestBytes, err := json.Marshal(request)

	if err != nil {
		return []model.GroupDto{}, err
	}

	return ReadBody[[]model.GroupDto](http.Post(h.config.HobURL+"/api/v1/groups/batch", "application/json", bytes.NewReader(requestBytes)))
}

func (h *HobClient) CreateIncomeBatch(request model.CreateIncomeBatchRequest) ([]model.IncomeDto, error) {
	requestBytes, err := json.Marshal(request)

	if err != nil {
		return []model.IncomeDto{}, err
	}

	return ReadBody[[]model.IncomeDto](http.Post(h.config.HobURL+"/api/v1/incomes/batch", "application/json", bytes.NewReader(requestBytes)))
}

func (h *HobClient) CreatePaymentBatch(request model.CreatePaymentBatchRequest) ([]model.PaymentDto, error) {
	requestBytes, err := json.Marshal(request)

	if err != nil {
		return []model.PaymentDto{}, err
	}

	return ReadBody[[]model.PaymentDto](http.Post(h.config.HobURL+"/api/v1/payments/batch", "application/json", bytes.NewReader(requestBytes)))
}

func (h *HobClient) UserExists(id string) bool {
	response, err := http.Get(h.config.HobURL + "/api/v1/users/" + id)

	if err != nil {
		log.Error().Err(err)
		return false
	}
	if response.StatusCode != 200 {
		return false
	}
	return true
}

func ReadBody[T any](response *http.Response, err error) (T, error) {
	t := *new(T)
	if err != nil {
		return t, err
	}

	body := response.Body

	defer body.Close()

	allBytes, err := ioutil.ReadAll(body)

	if err != nil {
		return t, err
	}

	if response.StatusCode != 200 && response.StatusCode != 201 {
		text := string(allBytes)
		log.Fatal().Msg(text)
		return t, errors.New(text)
	}

	err = json.Unmarshal(allBytes, &t)

	if err != nil {
		return t, err
	}

	return t, nil
}
