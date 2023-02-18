package wsclient

import (
	constants_fifashootup "bitbucket.org/esportsph/minigame-backend-golang/internal/constants/fifashootup"
	constants_fishprawncrab "bitbucket.org/esportsph/minigame-backend-golang/internal/constants/fishprawncrab"
	constants_lolcouple "bitbucket.org/esportsph/minigame-backend-golang/internal/constants/lolcouple"
	constants_loltower "bitbucket.org/esportsph/minigame-backend-golang/internal/constants/loltower"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/models"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/process"
	process_fifashootup "bitbucket.org/esportsph/minigame-backend-golang/internal/process/fifashootup"
	process_fishprawncrab "bitbucket.org/esportsph/minigame-backend-golang/internal/process/fishprawncrab"
	process_lolcouple "bitbucket.org/esportsph/minigame-backend-golang/internal/process/lolcouple"
	process_loltower "bitbucket.org/esportsph/minigame-backend-golang/internal/process/loltower"
)

func (wsc *wsclient) GetProcess() process.ProcessRequest {
	if wsc.processRequest == nil {
		switch wsc.identifier {
		case constants_loltower.Identifier:
			wsc.processRequest = process_loltower.NewLOLTowerProcess(wsc)
		case constants_lolcouple.Identifier:
			wsc.processRequest = process_lolcouple.NewLOLCoupleProcess(wsc)
		case constants_fifashootup.Identifier:
			wsc.processRequest = process_fifashootup.NewFIFAShootupProcess(wsc)
		case constants_fishprawncrab.Identifier:
			wsc.processRequest = process_fishprawncrab.NewFishPrawCrabProcess(wsc)
		default:
			panic("add identifier to load process")
		}
	}
	return wsc.processRequest
}

func (wsc *wsclient) GetIdentifier() string {
	return wsc.identifier
}

func (wsc *wsclient) GetGameID() int64 {
	switch wsc.identifier {
	case constants_loltower.Identifier:
		return constants_loltower.GameID
	case constants_lolcouple.Identifier:
		return constants_lolcouple.GameID
	case constants_fifashootup.Identifier:
		return constants_fifashootup.GameID
	case constants_fishprawncrab.Identifier:
		return constants_fishprawncrab.GameID
	default:
		panic("add identifier to load game id")
	}
}

func (wsc *wsclient) GetTableID() int64 {
	switch wsc.identifier {
	case constants_loltower.Identifier:
		return constants_loltower.TableID
	case constants_lolcouple.Identifier:
		return constants_lolcouple.TableID
	case constants_fifashootup.Identifier:
		return constants_fifashootup.TableID
	case constants_fishprawncrab.Identifier:
		return constants_fishprawncrab.TableID
	default:
		panic("add identifier to load table id")
	}
}

func (wsc *wsclient) GetUser() *models.User {
	return wsc.user
}
