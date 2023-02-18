package betsimulator

import (
	"fmt"
	"testing"
	"time"

	"bitbucket.org/esportsph/minigame-backend-golang/internal/logger"
)

func TestBetSimulatorGenerateProbability(t *testing.T) {

	logger.Info("probility: ", generateProbability(1000))
}

func TestBetSimulator(t *testing.T) {
	fakemembers := []BetSimulatorData{}
	for i := 1; i <= 200; i++ { //generate members
		fakemembers = append(fakemembers, &FakeMember{
			BetProbability: 20,
			SessionMinute:  120,
			Name:           fmt.Sprint("fakemember", i),
		})
	}

	betSimulator := NewBetSimulator()
	betSimulator.SetPhaseCallback(func(phase BetSimulatorPhase) {
		logger.Info("Bet Simulator Phase: ", phase)
	})
	betSimulator.SetActiveBettorsCallback(func(data BetSimulatorData) {
		logger.Info("Bet Placed by: ", data.GetID())
	})
	betSimulator.SetWinningBettorsCallback(func(data BetSimulatorData) {
		logger.Info("Bet Won by: ", data.GetID())
	})
	betSimulator.SetBettors(fakemembers)
	logger.Info("SetBettors SessionBettors: ", len(betSimulator.SessionBettors()))
	betSimulator.StartBetting()
	time.Sleep(7 * time.Second)
	betSimulator.StartResulting()
	time.Sleep(5 * time.Second)
	logger.Info("Bettors: ", len(betSimulator.Bettors()))
	logger.Info("Session bettors: ", len(betSimulator.SessionBettors()))
	logger.Info("Active bettors: ", len(betSimulator.ActiveBettors()))
	logger.Info("Winning bettors: ", len(betSimulator.WinningBettors()))

}
