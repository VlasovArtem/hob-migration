package migrator

import (
	"fmt"
	"github.com/VlasovArtem/hob-migration/src/client"
	"github.com/VlasovArtem/hob-migration/src/model"
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
	path, ok := requestMigrator.TypeToPathMap["incomes"]
	if !ok {
		log.Info().Msg("houses path not found")
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
				header:   []string{"House Identifier", "Groups", "Name", "Description", "Date", "Sum", "House Name"},
				parser:   migrator.parseCSVLine(),
				mapper:   migrator.mapIncomes,
			},
		},
		filePath: filePath,
	}

	return migrator
}

func (i *IncomeMigrator) mapIncomes(requests []model.CreateIncomeRequest) (responses []model.IncomeDto, err error) {
	if batch, err := i.client.CreateIncomeBatch(model.CreateIncomeBatchRequest{
		Incomes: requests,
	}); err != nil {
		log.Fatal().Err(err).Msg("failed to create income batch")
		return nil, err
	} else {
		return batch, nil
	}
}

func (i *IncomeMigrator) parseCSVLine() func(line []string) (model.CreateIncomeRequest, error) {
	return func(line []string) (model.CreateIncomeRequest, error) {
		sum, err := strconv.ParseFloat(line[3], 2)

		if err != nil {
			log.Fatal().Msg(fmt.Sprintf("sum not valid float %s", line[3]))
		}

		groupIds := func() []string {
			groups := line[5]
			if groups == "" {
				return []string{}
			}
			var groupIds []string
			for _, group := range strings.Split(groups, ",") {
				trimGroup := strings.Trim(group, " ")
				if dto, ok := i.groupMap[trimGroup]; ok {
					groupIds = append(groupIds, dto.Id.String())
				} else {
					log.Fatal().Msg(fmt.Sprintf("group with name %s not found", trimGroup))
				}
			}
			return groupIds
		}()

		houseId := func() string {
			if len(groupIds) == 0 {
				if dto, ok := i.houseMap[line[4]]; ok {
					return dto.Id.String()
				} else {
					log.Fatal().Msg("group name or house identifier is missing")
				}
			}
			return ""
		}()

		request := model.CreateIncomeRequest{
			Name:        line[0],
			Description: line[1],
			Date:        line[2],
			Sum:         float32(sum),
			HouseId:     houseId,
			GroupIds:    groupIds,
		}

		return request, nil
	}
}
