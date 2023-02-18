package gamemanager

import (
	"strconv"
	"testing"
	"time"

	constants_loltower "bitbucket.org/esportsph/minigame-backend-golang/internal/constants/loltower"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/db"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/models"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/settings"
	"bitbucket.org/esportsph/minigame-backend-golang/pkg/types"
	"gorm.io/gorm"
)

type testDatasource struct {
}

func (tds *testDatasource) GetIdentifier() string {
	return *types.String("testGameManager").Ptr()
}

func (tds *testDatasource) GetEventName() string {
	return "LOL Tower"
}

func (tds *testDatasource) GetGameID() int64 {
	return 5
}

func (tds *testDatasource) GetTableID() int64 {
	return 11
}

func (ds *testDatasource) GetMaxFutureHashes() int8 {
	return 1
}

func (tds *testDatasource) GetMaxFutureEvents() int8 {
	return 3
}

func (tds *testDatasource) GetMaxSequencePerHash() int {
	return settings.GetMaxHashSequenceCount().Int()
}

func (tds *testDatasource) GetHashSequenceResults(hashSequenceValue string) *[]models.EventResult {
	cards := []int{1, 2, 3, 4, 5}

	hexNum := hashSequenceValue[0:16]
	dec, _ := strconv.ParseInt(hexNum, 16, 64)
	res1 := int((dec % 5))
	cards = append(cards[:res1], cards[res1+1:]...)

	hexNum2 := hashSequenceValue[16:32]
	dec2, _ := strconv.ParseInt(hexNum2, 16, 64)
	res2 := int((dec2 % 4))
	cards = append(cards[:res2], cards[res2+1:]...)

	return &[]models.EventResult{
		{
			ResultType: 28,
			Value:      types.Array[int](cards).Join(","),
		},
	}
}

func (tds *testDatasource) GetGameDuration(eventResults *[]models.EventResult) time.Duration {
	return time.Duration(constants_loltower.GameDuration) * time.Second
}

func (tds *testDatasource) GetDB() *gorm.DB {
	return db.Shared()
}

func TestCreateFutureEvents(t *testing.T) {
	gameManager := NewGameManagerV2(&testDatasource{})

	if err := gameManager.CreateFutureEvents(); err != nil {
		t.Fatal(err.Error())
	}
}
