package gameloop

import (
	"context"
	"time"

	"bitbucket.org/esportsph/minigame-backend-golang/internal/utils"
)

type GamePhaseCallback func(elapsedTime Milliseconds, phase GamePhase)

type GamePhase interface {
	GetIdentifier() string                      //get identifier
	GetDuration() Milliseconds                  //get duration end ms - start ms
	IsTriggered() bool                          //is triggered
	Trigger()                                   //is triggered true
	ResetTrigger()                              //is triggered false
	Sleep(duration Milliseconds)                //sleep with duration
	CancelSleep()                               //cancel sleep
	SetStartMSOffset(milliseconds Milliseconds) //set start offset (used to extend or lessen duration  for sleep duration)
	SetEndMSOffset(milliseconds Milliseconds)   //set end offset (used to extend or lessen duration  for sleep duration)
	Run(elapsedTime Milliseconds) Milliseconds  //perform the callback stored in phase
}

type gamePhase struct {
	identifier    string
	startMS       Milliseconds
	startMSOffset Milliseconds
	endMS         Milliseconds
	endMSOffset   Milliseconds
	isTriggered   bool
	callback      *GamePhaseCallback
	sleepCtx      context.Context
	sleepCancel   func()
}

func NewGamePhase(identifier string, startMS Milliseconds, endMS Milliseconds, callback GamePhaseCallback) GamePhase {
	ctx, cancelFunc := utils.CancelContext()

	return &gamePhase{
		identifier:  identifier,
		startMS:     startMS,
		endMS:       endMS,
		callback:    &callback,
		sleepCtx:    ctx,
		sleepCancel: cancelFunc,
	}
}

func (gp *gamePhase) GetIdentifier() string {
	return gp.identifier
}

func (gp *gamePhase) GetDuration() Milliseconds {
	return gp.GetEndMS() - gp.GetStartMS()
}

func (gp *gamePhase) IsTriggered() bool {
	return gp.isTriggered
}

func (gp *gamePhase) Trigger() {
	gp.isTriggered = true
}

func (gp *gamePhase) ResetTrigger() {
	gp.isTriggered = false
}

func (gp *gamePhase) Sleep(duration Milliseconds) {
	utils.Sleep(time.Duration(duration)*time.Millisecond, gp.sleepCtx)
}

func (gp *gamePhase) CancelSleep() {
	gp.sleepCancel()
}

func (gp *gamePhase) SetStartMSOffset(milliseconds Milliseconds) {
	gp.startMSOffset = milliseconds
}

func (gp *gamePhase) SetEndMSOffset(milliseconds Milliseconds) {
	gp.endMSOffset = milliseconds
}

func (gp *gamePhase) Run(elapsedTime Milliseconds) Milliseconds {
	if !gp.IsTriggered() && gp.GetStartMS() <= elapsedTime && elapsedTime <= gp.GetEndMS() {
		(*gp.callback)(elapsedTime, gp)
		gp.Trigger()
		return gp.GetEndMS() - elapsedTime
	} else if !gp.IsTriggered() && elapsedTime > gp.GetStartMS() && elapsedTime > gp.GetEndMS() {
		gp.Trigger()
	}
	return 0
}

func (gp *gamePhase) GetStartMS() Milliseconds {
	return gp.startMS + gp.startMSOffset
}

func (gp *gamePhase) GetEndMS() Milliseconds {
	return gp.endMS + gp.endMSOffset
}

type gamePhases []GamePhase

func (phases gamePhases) Run(elapsedTime Milliseconds, preloadCallback func()) {
	for i := 0; i < len(phases); i++ {
		if remainingDuration := phases[i].Run(elapsedTime); remainingDuration > 0 {
			if phases.IsTriggered() {
				go preloadCallback()
			}
			phases[i].Sleep(remainingDuration)
		}
	}
}

func (phases gamePhases) CancelSleep() {
	for i := 0; i < len(phases); i++ {
		phases[i].CancelSleep()
	}
}

func (phases gamePhases) TotalDuration() Milliseconds {
	totalDuration := Milliseconds(0)

	for i := 0; i < len(phases); i++ {
		totalDuration += phases[i].GetDuration()
	}
	return totalDuration
}

func (phases gamePhases) IsTriggered() bool {
	for i := 0; i < len(phases); i++ {
		if !phases[i].IsTriggered() {
			return false
		}
	}
	return true
}

func (phases gamePhases) Trigger() {
	for i := 0; i < len(phases); i++ {
		phases[i].Trigger()
	}
}

func (phases gamePhases) ResetTrigger() {
	for i := 0; i < len(phases); i++ {
		phases[i].ResetTrigger()
	}
}

func (phases gamePhases) ResetOffset() {
	for i := 0; i < len(phases); i++ {
		phases[i].SetStartMSOffset(0)
		phases[i].SetEndMSOffset(0)
	}
}
