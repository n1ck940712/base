package process_odds

import (
	constants_lolcouple "bitbucket.org/esportsph/minigame-backend-golang/internal/constants/lolcouple"
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
		Odds3Under: constants_lolcouple.Selection1Odds.EuroToHK(2).Ptr(),
		Odds4Under: constants_lolcouple.Selection2Odds.EuroToHK(2).Ptr(),
		Odds5Under: constants_lolcouple.Selection3Odds.EuroToHK(2).Ptr(),
		Odds6Under: constants_lolcouple.Selection4Odds.EuroToHK(2).Ptr(),
		Odds7:      constants_lolcouple.Selection5Odds.EuroToHK(2).Ptr(),
		Odds8Over:  constants_lolcouple.Selection6Odds.EuroToHK(2).Ptr(),
		Odds9Over:  constants_lolcouple.Selection7Odds.EuroToHK(2).Ptr(),
		Odds10Over: constants_lolcouple.Selection8Odds.EuroToHK(2).Ptr(),
		Odds11Over: constants_lolcouple.Selection9Odds.EuroToHK(2).Ptr(),
		Odds777:    constants_lolcouple.Selection10Odds.EuroToHK(2).Ptr(),
		OddsBonus:  constants_lolcouple.SelectionBonusOdds.EuroToHK(2).Ptr(),
		OddsBunus7: constants_lolcouple.Selection5BonusOdds.EuroToHK(2).Ptr(),
	}

	return &oddsData
}
