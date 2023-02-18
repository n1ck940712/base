package gamemanager_loltower

import (
	"time"

	constants_loltower "bitbucket.org/esportsph/minigame-backend-golang/internal/constants/loltower"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/gamemanager"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/logger"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/models"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/settings"
	"bitbucket.org/esportsph/minigame-backend-golang/pkg/hashutil"
)

type datasource struct {
	identifier string
}

func NewDatasource(identifier string) gamemanager.Datasource {
	return &datasource{
		identifier: identifier,
	}
}

func (ds *datasource) GetIdentifier() string {
	return ds.identifier
}

func (ds *datasource) GetEventName() string {
	return constants_loltower.GameName
}

func (ds *datasource) GetGameID() int64 {
	return constants_loltower.GameID
}

func (ds *datasource) GetTableID() int64 {
	return constants_loltower.TableID
}

func (ds *datasource) GetMaxFutureHashes() int8 {
	return constants_loltower.MaxFutureHashes
}

func (ds *datasource) GetMaxFutureEvents() int8 {
	return constants_loltower.MaxFutureEvents
}

func (ds *datasource) GetMaxSequencePerHash() int {
	return settings.GetMaxHashSequenceCount().Int()
}

func (ds *datasource) GetHashSequenceResults(hashSequenceValue string) *[]models.EventResult {
	result, err := hashutil.LOLTowerGenerateResult(hashutil.NewHash(hashSequenceValue))

	if err != nil {
		logger.Error(ds.identifier, " GetHashSequenceResults error: ", err.Error())
		return nil
	}
	return &[]models.EventResult{{ResultType: constants_loltower.EventResultType, Value: result}}
}

func (ds *datasource) GetGameDuration(eventResults *[]models.EventResult) time.Duration {
	return time.Duration(constants_loltower.GameDuration) * time.Second
}
