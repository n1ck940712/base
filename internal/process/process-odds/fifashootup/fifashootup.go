package process_odds

import (
	constants_fifashootup "bitbucket.org/esportsph/minigame-backend-golang/internal/constants/fifashootup"
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
		Odds1:  constants_fifashootup.Card1Odds.Ptr(),
		Odds2:  constants_fifashootup.Card2Odds.Ptr(),
		Odds3:  constants_fifashootup.Card3Odds.Ptr(),
		Odds4:  constants_fifashootup.Card4Odds.Ptr(),
		Odds5:  constants_fifashootup.Card5Odds.Ptr(),
		Odds6:  constants_fifashootup.Card6Odds.Ptr(),
		Odds7:  constants_fifashootup.Card7Odds.Ptr(),
		Odds8:  constants_fifashootup.Card8Odds.Ptr(),
		Odds9:  constants_fifashootup.Card9Odds.Ptr(),
		Odds10: constants_fifashootup.Card10Odds.Ptr(),
		Odds11: constants_fifashootup.Card11Odds.Ptr(),
		Odds12: constants_fifashootup.Card12Odds.Ptr(),
	}

	return &oddsData
}
