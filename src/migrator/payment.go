package migrator

import (
	"fmt"
	"github.com/VlasovArtem/hob-migration/src/client"
	"github.com/VlasovArtem/hob-migration/src/config"
	"github.com/VlasovArtem/hob-migration/src/model"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"strconv"
	"strings"
)

type PaymentMigrator struct {
	*BaseMigrator[[]model.PaymentDto]
	client   *client.HobClient
	houseMap map[string]model.HouseDto
	groupMap map[string]model.GroupDto
	config   *config.CMDConfig
}

func NewPaymentMigrator(
	requestMigrator RequestMigrator,
	config *config.CMDConfig,
	hobClient *client.HobClient,
	houseMap map[string]model.HouseDto,
) *PaymentMigrator {
	log.Info().Msg("Starting Payment Migrator")

	path, ok := requestMigrator.TypeToPathMap["payments"]
	if !ok {
		log.Info().Msg("payments path not found")
		return nil
	}
	migrator := &PaymentMigrator{
		client:   hobClient,
		houseMap: houseMap,
		config:   config,
	}
	filePath := path
	migrator.BaseMigrator = &BaseMigrator[[]model.PaymentDto]{
		mappers: map[string]Mapper[[]model.PaymentDto]{
			"csv": &CSVMigrator[model.CreatePaymentRequest, []model.PaymentDto]{
				filePath: filePath,
				header:   []string{"House Identifier", "Name", "Description", "Date", "Sum"},
				parser:   migrator.parseCSVLine(),
				mapper:   migrator.mapPayments,
			},
		},
		filePath: filePath,
		rollback: migrator.rollback,
	}

	return migrator
}

func (p *PaymentMigrator) mapPayments(requests []model.CreatePaymentRequest) (responses []model.PaymentDto, err error) {
	request := model.CreatePaymentBatchRequest{Payments: requests}

	if response, err := p.client.CreatePaymentBatch(request); err != nil {
		log.Error().Err(err).Msg("error while creating payments")
		return nil, err
	} else {
		log.Info().Msg(fmt.Sprintf("%d payments created", len(response)))
		return response, nil
	}
}

func (p *PaymentMigrator) parseCSVLine() func(line []string, lineNumber int) (model.CreatePaymentRequest, error) {
	return func(line []string, lineNumber int) (model.CreatePaymentRequest, error) {
		sum, err := strconv.ParseFloat(line[4], 2)

		if err != nil {
			log.Error().Err(err).Msgf("sum not valid float %s at the csv line %d", line[3], lineNumber)
			return model.CreatePaymentRequest{}, err
		}

		houseId, err := func() (string, error) {
			if dto, ok := p.houseMap[line[0]]; ok {
				return dto.Id.String(), nil
			}
			return uuid.Nil.String(), fmt.Errorf("house identifier is missing at the csv line %d", lineNumber)
		}()

		if err != nil {
			return model.CreatePaymentRequest{}, err
		}

		request := model.CreatePaymentRequest{
			Name:        line[1],
			Description: strings.Replace(line[2], ";", ",", -1),
			HouseId:     houseId,
			UserId:      p.config.UserId,
			Date:        line[3],
			ProviderId:  nil,
			Sum:         float32(sum),
		}

		return request, nil
	}
}

func (p *PaymentMigrator) rollback(data []model.PaymentDto) {
	log.Info().Msg("Rolling back payments")
	if len(data) == 0 {
		log.Info().Msg("No payments to rollback")
	}

	for _, payment := range data {
		if err := p.client.DeletePaymentById(payment.Id); err != nil {
			log.Error().Err(err).Msgf("Failed to delete payment with id %s and name %s", payment.Id, payment.Name)
		} else {
			log.Info().Msgf("Payment with id %s and name %s deleted", payment.Id, payment.Name)
		}
	}
}

func (p *PaymentMigrator) Migrate(rollbackOperation []func()) ([]model.PaymentDto, []func()) {
	if p != nil {
		return p.BaseMigrator.Migrate(rollbackOperation)
	}
	return nil, rollbackOperation
}
