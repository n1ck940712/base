package gamemanager_fifashootup

import (
	"time"

	constants_fifashootup "bitbucket.org/esportsph/minigame-backend-golang/internal/constants/fifashootup"
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
	return constants_fifashootup.GameName
}

func (ds *datasource) GetGameID() int64 {
	return constants_fifashootup.GameID
}

func (ds *datasource) GetTableID() int64 {
	return constants_fifashootup.TableID
}

func (ds *datasource) GetMaxFutureHashes() int8 {
	return constants_fifashootup.MaxFutureHashes
}

func (ds *datasource) GetMaxFutureEvents() int8 {
	return constants_fifashootup.MaxFutureEvents
}

func (ds *datasource) GetMaxSequencePerHash() int {
	return settings.GetMaxHashSequenceCount().Int()
}

func (ds *datasource) GetHashSequenceResults(hashSequenceValue string) *[]models.EventResult {
	selections, result, err := hashutil.FIFAShootupGenerateResult(hashutil.NewHash(hashSequenceValue))

	if err != nil {
		logger.Error(ds.identifier, " GetHashSequenceResults error: ", err.Error())
		return nil
	}
	return &[]models.EventResult{
		{
			ResultType: constants_fifashootup.EventResultType1,
			Value:      selections,
		},
		{
			ResultType: constants_fifashootup.EventResultType2,
			Value:      result,
		},
	}
}

func (ds *datasource) GetGameDuration(eventResults *[]models.EventResult) time.Duration {
	totalDuration := constants_fifashootup.StartBetMS + constants_fifashootup.StopBetMS + constants_fifashootup.ShowResultMS

	return time.Duration(totalDuration) * time.Millisecond
}
