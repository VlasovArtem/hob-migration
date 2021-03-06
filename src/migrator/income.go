package migrator

import (
	"fmt"
	"github.com/VlasovArtem/hob-migration/src/client"
	"github.com/VlasovArtem/hob-migration/src/model"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"strconv"
	"strings"
)

type IncomeMigrator struct {
	*BaseMigrator[[]model.IncomeDto]
	client   *client.HobClient
	houseMap map[string]model.HouseDto
	groupMap map[string]model.GroupDto
}

func NewIncomeMigrator(
	requestMigrator RequestMigrator,
	hobClient *client.HobClient,
	houseMap map[string]model.HouseDto,
	groupMap map[string]model.GroupDto,
) *IncomeMigrator {
	log.Info().Msg("Starting Income Migrator")

	path, ok := requestMigrator.TypeToPathMap["incomes"]
	if !ok {
		log.Info().Msg("income path not found")
		return nil
	}
	migrator := &IncomeMigrator{
		client:   hobClient,
		houseMap: houseMap,
		groupMap: groupMap,
	}
	filePath := path
	migrator.BaseMigrator = &BaseMigrator[[]model.IncomeDto]{
		mappers: map[string]Mapper[[]model.IncomeDto]{
			"csv": &CSVMigrator[model.CreateIncomeRequest, []model.IncomeDto]{
				filePath: filePath,
				header:   []string{"House Identifier", "Groups", "Name", "Description", "Date", "Sum"},
				parser:   migrator.parseCSVLine(),
				mapper:   migrator.mapIncomes,
			},
		},
		filePath: filePath,
		rollback: migrator.rollback,
	}

	return migrator
}

func (i *IncomeMigrator) mapIncomes(requests []model.CreateIncomeRequest) (responses []model.IncomeDto, err error) {
	request := model.CreateIncomeBatchRequest{Incomes: requests}

	if response, err := i.client.CreateIncomeBatch(request); err != nil {
		log.Error().Err(err).Msg("failed to create income batch")
		return nil, err
	} else {
		log.Info().Msg(fmt.Sprintf("%d incomes created", len(response)))
		return response, nil
	}
}

func (i *IncomeMigrator) parseCSVLine() func(line []string, lineNumber int) (model.CreateIncomeRequest, error) {
	return func(line []string, lineNumber int) (model.CreateIncomeRequest, error) {
		sum, err := strconv.ParseFloat(line[5], 2)

		if err != nil {
			log.Error().Msgf("sum not valid float %s at the csv line %d", line[3], lineNumber)
			return model.CreateIncomeRequest{}, err
		}

		groupIds, err := func() ([]string, error) {
			groups := line[1]
			if len(groups) == 0 {
				return nil, nil
			}
			var groupIds []string
			for _, group := range strings.Split(groups, ",") {
				trimGroup := strings.Trim(group, " ")
				if dto, ok := i.groupMap[trimGroup]; ok {
					groupIds = append(groupIds, dto.Id.String())
				} else {
					err = errors.Errorf("group with name %s not found at the csv line %d", trimGroup, lineNumber)
					return nil, err
				}
			}
			return groupIds, nil
		}()

		if err != nil {
			return model.CreateIncomeRequest{}, err
		}

		houseId, err := func() (*string, error) {
			if len(groupIds) == 0 {
				if dto, ok := i.houseMap[line[0]]; ok {
					id := dto.Id.String()
					return &id, nil
				} else {
					err = errors.Errorf("house with name %s not found at the csv line %d", line[0], lineNumber)
					return nil, err
				}
			}
			return nil, nil
		}()

		if err != nil {
			return model.CreateIncomeRequest{}, err
		}

		request := model.CreateIncomeRequest{
			Name:        line[2],
			Description: strings.Replace(line[3], ";", ",", -1),
			Date:        line[4],
			Sum:         float32(sum),
			HouseId:     houseId,
			GroupIds:    groupIds,
		}

		return request, nil
	}
}

func (i *IncomeMigrator) rollback(data []model.IncomeDto) {
	log.Info().Msg("Rolling back incomes")
	if len(data) == 0 {
		log.Info().Msg("No incomes to rollback")
	}

	for _, income := range data {
		if err := i.client.DeleteIncomeById(income.Id); err != nil {
			log.Error().Err(err).Msgf("Failed to delete income with id %s and name %s", income.Id, income.Name)
		} else {
			log.Info().Msgf("Income with id %s and name %s deleted", income.Id, income.Name)
		}
	}
}

func (i *IncomeMigrator) Migrate(rollbackOperation []func()) ([]model.IncomeDto, []func()) {
	if i != nil {
		return i.BaseMigrator.Migrate(rollbackOperation)
	}
	return nil, rollbackOperation
}
