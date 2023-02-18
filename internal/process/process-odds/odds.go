package process_odds

import "bitbucket.org/esportsph/minigame-backend-golang/internal/models/response"

const OddsType = "odds"

type OddsDatasource interface {
	GetIdentifier() string
}

type OddsProcess interface {
	GetOdds() response.ResponseData
}
