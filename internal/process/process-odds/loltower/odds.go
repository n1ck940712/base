package process_odds

import (
	constants_loltower "bitbucket.org/esportsph/minigame-backend-golang/internal/constants/loltower"
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
		OddsLevel1:  constants_loltower.Level1Odds.EuroToHK(2).Ptr(),
		OddsLevel2:  constants_loltower.Level2Odds.EuroToHK(2).Ptr(),
		OddsLevel3:  constants_loltower.Level3Odds.EuroToHK(2).Ptr(),
		OddsLevel4:  constants_loltower.Level4Odds.EuroToHK(2).Ptr(),
		OddsLevel5:  constants_loltower.Level5Odds.EuroToHK(2).Ptr(),
		OddsLevel6:  constants_loltower.Level6Odds.EuroToHK(2).Ptr(),
		OddsLevel7:  constants_loltower.Level7Odds.EuroToHK(2).Ptr(),
		OddsLevel8:  constants_loltower.Level8Odds.EuroToHK(2).Ptr(),
		OddsLevel9:  constants_loltower.Level9Odds.EuroToHK(2).Ptr(),
		OddsLevel10: constants_loltower.Level10Odds.EuroToHK(2).Ptr(),
	}

	return &oddsData
}
