package migrator

import (
	"fmt"
	"github.com/VlasovArtem/hob-migration/src/client"
	"github.com/VlasovArtem/hob-migration/src/config"
	"github.com/VlasovArtem/hob-migration/src/model"
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
	path, ok := requestMigrator.TypeToPathMap["payments"]
	if !ok {
		log.Info().Msg("houses path not found")
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
				header:   []string{"House Identifier", "Name", "Description", "Sum", "Date"},
				parser:   migrator.parseCSVLine(),
				mapper:   migrator.mapPayments,
			},
		},
		filePath: filePath,
	}

	return migrator
}

func (p *PaymentMigrator) mapPayments(requests []model.CreatePaymentRequest) (responses []model.PaymentDto, err error) {
	paymentBatch, err := p.client.CreatePaymentBatch(model.CreatePaymentBatchRequest{Payments: requests})
	if err != nil {
		log.Fatal().Err(err).Msg("error while creating payments")
		return nil, err
	} else {
		return paymentBatch, nil
	}
}

func (p *PaymentMigrator) parseCSVLine() func(line []string) (model.CreatePaymentRequest, error) {
	return func(line []string) (model.CreatePaymentRequest, error) {
		sum, err := strconv.ParseFloat(line[3], 2)

		if err != nil {
			log.Fatal().Msg(fmt.Sprintf("sum not valid float %s", line[3]))
		}

		houseId := func() string {
			if dto, ok := p.houseMap[line[0]]; ok {
				return dto.Id.String()
			}
			log.Fatal().Msg("house identifier is missing")
			return ""
		}()

		request := model.CreatePaymentRequest{
			Name:        line[1],
			Description: line[2],
			HouseId:     houseId,
			UserId:      p.config.UserId,
			Date:        line[4],
			Sum:         float32(sum),
		}

		return request, nil
	}
}
