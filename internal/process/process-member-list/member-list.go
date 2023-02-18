package process_member_list

import (
	"bitbucket.org/esportsph/minigame-backend-golang/internal/logger"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/models/response"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/redis"
)

const MemberListType = "member_list"

type MemberListDatasource interface {
	GetIdentifier() string
}

type memberListProcess struct {
	datasource MemberListDatasource
}

type MemberListProcess interface {
	GetMemberList() response.ResponseData
}

func NewMemberListProcess(datasource MemberListDatasource) MemberListProcess {
	return &memberListProcess{datasource: datasource}
}

func (mlp *memberListProcess) GetMemberList() response.ResponseData {
	if leaderboard, err := redis.GetLeaderboard(mlp.datasource.GetIdentifier()); err != nil {
		logger.Info(mlp.datasource.GetIdentifier(), " GetLeaderboard error: ", err.Error())
		return nil
	} else {
		return leaderboard
	}
}
