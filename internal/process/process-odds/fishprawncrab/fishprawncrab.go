package process_odds

import (
	constants_fishprawncrab "bitbucket.org/esportsph/minigame-backend-golang/internal/constants/fishprawncrab"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/models/response"
	process_odds "bitbucket.org/esportsph/minigame-backend-golang/internal/process/process-odds"
)

type OddsProcess interface {
	process_odds.OddsProcess
}
type oddsProcess struct {
	datasource process_odds.OddsDatasource
}

func NewOddsProcess(datasource process_odds.OddsDatasource) OddsProcess {
	return &oddsProcess{datasource: datasource}
}

func (op *oddsProcess) GetOdds() response.ResponseData {
	oddsData := response.Oddsdata{
		OddsSingle: constants_fishprawncrab.SingleOdds.Ptr(),
		OddsDouble: constants_fishprawncrab.DoubleOdds.Ptr(),
		OddsTriple: constants_fishprawncrab.TripleOdds.Ptr(),
	}

	return &oddsData
}
