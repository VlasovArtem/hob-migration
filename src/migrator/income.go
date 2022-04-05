package migrator

import (
	"fmt"
	"github.com/VlasovArtem/hob-migration/src/client"
	"github.com/VlasovArtem/hob-migration/src/model"
	"github.com/google/uuid"
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
	}

	return migrator
}

func (i *IncomeMigrator) mapIncomes(requests []model.CreateIncomeRequest) (responses []model.IncomeDto, err error) {
	request := model.CreateIncomeBatchRequest{Incomes: requests}

	if response, err := i.client.CreateIncomeBatch(request); err != nil {
		log.Fatal().Err(err).Msg("failed to create income batch")
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
			log.Fatal().Msgf("sum not valid float %s at the csv line %d", line[3], lineNumber)
		}

		groupIds := func() []string {
			groups := line[1]
			if groups == "" {
				return nil
			}
			var groupIds []string
			for _, group := range strings.Split(groups, ",") {
				trimGroup := strings.Trim(group, " ")
				if dto, ok := i.groupMap[trimGroup]; ok {
					groupIds = append(groupIds, dto.Id.String())
				} else {
					log.Fatal().Msgf("group with name %s not found at the csv line %d", trimGroup, lineNumber)
				}
			}
			return groupIds
		}()

		houseId := func() string {
			if len(groupIds) == 0 {
				if dto, ok := i.houseMap[line[0]]; ok {
					return dto.Id.String()
				} else {
					log.Fatal().Msgf("group name or house identifier is missing at the csv line %d", lineNumber)
				}
			}
			return uuid.Nil.String()
		}()

		request := model.CreateIncomeRequest{
			Name:        line[2],
			Description: line[3],
			Date:        line[4],
			Sum:         float32(sum),
			HouseId:     houseId,
			GroupIds:    groupIds,
		}

		return request, nil
	}
}
