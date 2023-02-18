package betsimulator_loltower

import (
	"bitbucket.org/esportsph/minigame-backend-golang/internal/betsimulator"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/constants"
	constants_loltower "bitbucket.org/esportsph/minigame-backend-golang/internal/constants/loltower"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/db"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/logger"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/models"
	"bitbucket.org/esportsph/minigame-backend-golang/pkg/types"
)

type leaderboardUser struct {
	Name string `json:"name"`
}

func (lbu *leaderboardUser) EncryptedID() string {
	return types.String(lbu.Name).Bytes().SHA256()
}

func (lbu *leaderboardUser) EncryptedName() string {
	return string(types.String(lbu.Name).Mask("*", constants_loltower.MemberCodeMaskCount, 4))
}

type BetSimulatorLOLTower struct {
	bs          betsimulator.BetSimulator
	leaderBoard map[string]int8 //leaderBoard["{name}"]{level_won} //enhancement save to redis
}

func NewBetSimulator() *BetSimulatorLOLTower {
	bs := betsimulator.NewBetSimulator()

	bs.SetActiveBettorsCallback(func(data betsimulator.BetSimulatorData) {
		// logger.Debug(constants_loltower.Identifier, " Bet Placed by: ", data.GetID())
	})
	bs.SetWinningBettorsCallback(func(data betsimulator.BetSimulatorData) {
		// logger.Debug(constants_loltower.Identifier, " Bet Won by: ", data.GetID())
	})
	bs.SetPhaseCallback(func(phase betsimulator.BetSimulatorPhase) {
		switch phase {
		case betsimulator.StartBetting:
			logger.Debug(constants_loltower.Identifier, " BetSimulatorLOLTower Phase: ", phase, " bettors: ", len(bs.SessionBettors()))
		default:
			logger.Debug(constants_loltower.Identifier, " BetSimulatorLOLTower Phase: ", phase)
		}
	})
	return &BetSimulatorLOLTower{bs: bs}
}

func (bslt *BetSimulatorLOLTower) UpdateBettors() {
	fakeMembers := []models.FakeMember{}

	if err := db.Shared().Where(models.FakeMember{MiniGameTableID: constants_loltower.TableID}).Find(&fakeMembers).Error; err != nil {
		logger.Error(constants_loltower.Identifier, " Get fake members error: ", err.Error())
	}
	bettors := []betsimulator.BetSimulatorData{}

	for _, fakeMember := range fakeMembers {
		bettor := fakeMember
		bettors = append(bettors, &bettor)
	}
	bslt.bs.SetBettors(bettors)
}

func (bslt *BetSimulatorLOLTower) StartBetting() {
	bslt.Cleanup()
	bslt.UpdateBettors()
	bslt.bs.StartBetting()
}

func (bslt *BetSimulatorLOLTower) StartResulting() {
	bslt.bs.StartResulting()
}

func (bslt *BetSimulatorLOLTower) UpdateResults() {
	uLeaderBoard := map[string]int8{}

	for _, bsd := range bslt.bs.WinningBettors() {
		if level, ok := bslt.leaderBoard[bsd.GetName()]; ok { //found
			uLeaderBoard[bsd.GetName()] = level + 1
		} else {
			uLeaderBoard[bsd.GetName()] = 1
		}
	}
	bslt.leaderBoard = uLeaderBoard
}

func (bslt *BetSimulatorLOLTower) GetLeaderBoard() (level int8, leaderboard []leaderboardUser) {
	uLeaderBoard := []leaderboardUser{}
	uLevel := int8(0)

	for uName, level := range bslt.leaderBoard {
		if level >= constants_loltower.MaxLevel {
			continue
		}
		if level > uLevel {
			uLevel = level
			uLeaderBoard = []leaderboardUser{} //reset container

			uLeaderBoard = append(uLeaderBoard, leaderboardUser{Name: uName})
		} else if level == uLevel {
			uLeaderBoard = append(uLeaderBoard, leaderboardUser{Name: uName})
		}
	}
	return uLevel, uLeaderBoard
}

func (bslt *BetSimulatorLOLTower) GetChampions() []leaderboardUser {
	uLeaderBoard := []leaderboardUser{}

	for uName, level := range bslt.leaderBoard {
		if level == constants_loltower.MaxLevel {
			uLeaderBoard = append(uLeaderBoard, leaderboardUser{Name: uName})
		}
	}
	return uLeaderBoard
}

func (bslt *BetSimulatorLOLTower) Cleanup() {
	leaderBoard := bslt.leaderBoard
	//remove champions
	for uName, level := range leaderBoard {
		if level >= constants.LOL_TOWER_MAX_LEVEL {
			delete(bslt.leaderBoard, uName)
		}
	}
}
