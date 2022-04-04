package migrator

import (
	"github.com/VlasovArtem/hob-migration/src/client"
	"github.com/VlasovArtem/hob-migration/src/model"
	"github.com/rs/zerolog/log"
)

type GroupMigrator struct {
	*BaseMigrator[map[string]model.GroupDto]
	client *client.HobClient
}

func NewGroupMigrator(requestMigrator RequestMigrator, hobClient *client.HobClient) *GroupMigrator {
	path, ok := requestMigrator.TypeToPathMap["groups"]
	if !ok {
		log.Info().Msg("groups path not found")
		return nil
	}
	migrator := &GroupMigrator{
		client: hobClient,
	}
	filePath := path
	migrator.BaseMigrator = &BaseMigrator[map[string]model.GroupDto]{
		mappers: map[string]Mapper[map[string]model.GroupDto]{
			"csv": &CSVMigrator[model.CreateGroupRequest, map[string]model.GroupDto]{
				filePath: filePath,
				header:   []string{"Name"},
				parser:   migrator.parseCSVLine(),
				mapper:   migrator.mapGroups,
			},
		},
		filePath: filePath,
	}

	return migrator
}

func (g *GroupMigrator) mapGroups(requests []model.CreateGroupRequest) (map[string]model.GroupDto, error) {
	var response = make(map[string]model.GroupDto)

	if batchResponse, err := g.client.CreateGroupBatch(model.CreateGroupBatchRequest{Groups: requests}); err != nil {
		return nil, err
	} else {
		for _, group := range batchResponse {
			response[group.Name] = group
		}
	}

	return response, nil
}

func (g *GroupMigrator) parseCSVLine() func(line []string) (model.CreateGroupRequest, error) {
	return func(line []string) (model.CreateGroupRequest, error) {
		request := model.CreateGroupRequest{
			Name: line[0],
		}

		return request, nil
	}
}
