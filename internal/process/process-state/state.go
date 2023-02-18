package process_state

import (
	"bitbucket.org/esportsph/minigame-backend-golang/internal/logger"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/models/response"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/redis"
)

const StateType = "state"

type StateDatasource interface {
	GetIdentifier() string
}

type stateProcess struct {
	datasource StateDatasource
}

type StateProcess interface {
	GetState() response.ResponseData
}

func NewStateProcess(datasource StateDatasource) StateProcess {
	return &stateProcess{datasource: datasource}
}

func (sp *stateProcess) GetState() response.ResponseData {
	if state, err := redis.GetPublishState(sp.datasource.GetIdentifier()); err != nil {
		logger.Info(sp.datasource.GetIdentifier(), " GetPublishState error: ", err.Error())
		return nil
	} else {
		return state
	}
}
