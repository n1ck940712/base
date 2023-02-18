package gamemanager_fishprawncrab

import (
	"time"

	constants_fishprawncrab "bitbucket.org/esportsph/minigame-backend-golang/internal/constants/fishprawncrab"
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
	return constants_fishprawncrab.GameName
}

func (ds *datasource) GetGameID() int64 {
	return constants_fishprawncrab.GameID
}

func (ds *datasource) GetTableID() int64 {
	return constants_fishprawncrab.TableID
}

func (ds *datasource) GetMaxFutureHashes() int8 {
	return constants_fishprawncrab.MaxFutureHashes
}

func (ds *datasource) GetMaxFutureEvents() int8 {
	return constants_fishprawncrab.MaxFutureEvents
}

func (ds *datasource) GetMaxSequencePerHash() int {
	return settings.GetMaxHashSequenceCount().Int()
}

func (ds *datasource) GetHashSequenceResults(hashSequenceValue string) *[]models.EventResult {
	result, err := hashutil.FishPrawnCrabGenerateResult(hashutil.NewHash(hashSequenceValue))

	if err != nil {
		logger.Error(ds.identifier, " GetHashSequenceResults error: ", err.Error())
		return nil
	}
	return &[]models.EventResult{{ResultType: constants_fishprawncrab.EventResultType, Value: result}}
}

func (ds *datasource) GetGameDuration(eventResults *[]models.EventResult) time.Duration {
	totalDuration := constants_fishprawncrab.StartBetMS + constants_fishprawncrab.StopBetMS + constants_fishprawncrab.ShowResultMS

	return time.Duration(totalDuration) * time.Millisecond
}
