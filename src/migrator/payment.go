package migrator

import (
	"fmt"
	"github.com/VlasovArtem/hob-migration/src/client"
	"github.com/VlasovArtem/hob-migration/src/config"
	"github.com/VlasovArtem/hob-migration/src/model"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"strconv"
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
	}

	return migrator
}

func (p *PaymentMigrator) mapPayments(requests []model.CreatePaymentRequest) (responses []model.PaymentDto, err error) {
	request := model.CreatePaymentBatchRequest{Payments: requests}

	if response, err := p.client.CreatePaymentBatch(request); err != nil {
		log.Fatal().Err(err).Msg("error while creating payments")
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
			log.Fatal().Msgf("sum not valid float %s at the csv line %d", line[3], lineNumber)
		}

		houseId := func() string {
			if dto, ok := p.houseMap[line[0]]; ok {
				return dto.Id.String()
			}
			log.Fatal().Msgf("house identifier is missing at the csv line %d", lineNumber)
			return uuid.Nil.String()
		}()

		request := model.CreatePaymentRequest{
			Name:        line[1],
			Description: line[2],
			HouseId:     houseId,
			UserId:      p.config.UserId,
			Date:        line[3],
			ProviderId:  nil,
			Sum:         float32(sum),
		}

		return request, nil
	}
}
