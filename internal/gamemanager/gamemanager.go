package gamemanager

import (
	"errors"
	"math"
	"time"

	"bitbucket.org/esportsph/minigame-backend-golang/internal/constants"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/db"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/logger"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/models"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/settings"
	"bitbucket.org/esportsph/minigame-backend-golang/pkg/hashutil"
	"bitbucket.org/esportsph/minigame-backend-golang/pkg/measure"
	"bitbucket.org/esportsph/minigame-backend-golang/pkg/types"
	"gorm.io/gorm"
)

type Datasource interface {
	GetIdentifier() string
	GetEventName() string
	GetGameID() int64
	GetTableID() int64
	GetMaxFutureHashes() int8
	GetMaxFutureEvents() int8
	GetMaxSequencePerHash() int
	GetHashSequenceResults(hashSequenceValue string) *[]models.EventResult
	GetGameDuration(eventResults *[]models.EventResult) time.Duration
}

type GameManager interface {
	CreateFutureHashes() error
	CreateFutureEvents() error
	GetFutureEvents() *[]models.Event
	GetDatasource() Datasource
}

type gameManager struct {
	datasource Datasource
}

func NewGameManagerV2(datasource Datasource) GameManager {
	return &gameManager{datasource: datasource}
}

func (gm *gameManager) CreateFutureHashes() error {
	if settings.GetEnvironment().String() != "local" {
		return nil
	}
	exec := measure.NewExecution()
	defer func() {
		logger.Debug(gm.GetIdentifier()+" CreateFutureHashes execution time: ", exec.Done())
	}()
	queuedHashes := []models.Hash{}

	if err := db.Shared().Where(models.Hash{Status: constants.HashStatusQueued, TableID: gm.datasource.GetTableID()}).Find(&queuedHashes).Error; err != nil {
		return err
	}
	remainingHashesCount := int(gm.datasource.GetMaxFutureHashes()) - len(queuedHashes)

	if remainingHashesCount <= 0 {
		logger.Debug(gm.GetIdentifier(), " CreateFutureHashes has queued (", len(queuedHashes), ")")
		return nil
	}
	createdHashes := []models.Hash{}

	if err := db.Shared().Transaction(func(tx *gorm.DB) error {
		for i := 0; i < remainingHashesCount; i++ {
			sequences := []models.HashSequence{}
			seed := hashutil.GenerateHash()

			for i := 1; i <= gm.datasource.GetMaxSequencePerHash(); i++ {
				sequences = append(sequences, models.HashSequence{
					Value:    hashutil.GenerateHash(seed),
					Sequence: i,
				})
			}
			hash := models.Hash{
				Seed:      seed,
				GameID:    gm.datasource.GetGameID(),
				TableID:   gm.datasource.GetTableID(),
				Status:    constants.HashStatusQueued,
				Sequences: &sequences,
			}
			if err := tx.Create(&hash).Error; err != nil {
				return err
			}
			createdHashes = append(createdHashes, hash)
		}
		return nil
	}); err != nil {
		logger.Error(gm.GetIdentifier(), " CreateFutureHashes transaction error: ", err.Error())
		return err
	}
	logger.Info(gm.GetIdentifier(), " CreateFutureHashes queued created (", len(createdHashes), ")")
	return nil
}

func (gm *gameManager) CreateFutureEvents() error {
	exec := measure.NewExecution()
	defer func() {
		logger.Debug(gm.GetIdentifier()+" CreateFutureEvents execution time: ", exec.Done())
	}()
	futureEvents := gm.GetFutureEvents()

	if futureEvents == nil {
		return errors.New(gm.GetIdentifier() + " CreateFutureEvents futureEvents is nil")
	}
	remainingEventsCount := int(gm.datasource.GetMaxFutureEvents()) - len(*futureEvents)

	if remainingEventsCount <= 0 { //exit since future events is complete
		logger.Debug(gm.GetIdentifier(), " CreateFutureEvents is complete (", len(*futureEvents), ")")
		return nil
	}
	selectionHeader := gm.GetSelectionHeader()

	if selectionHeader == nil {
		return errors.New(gm.GetIdentifier() + " CreateFutureEvents selectionHeader is nil")
	}
	gameTable := gm.GetGameTable()

	if gameTable == nil {
		return errors.New(gm.GetIdentifier() + " CreateFutureEvents gameTable is nil")
	}
	prevEventID := (*int64)(nil)
	startDateTime := time.Now()
	completedHashes := (*[]models.Hash)(nil)
	queuedHash := (*models.Hash)(nil)
	nextHashSequences := (*[]models.HashSequence)(nil)

	if len(*futureEvents) > 0 {
		prevEventID = (*futureEvents)[0].ID
		startDateTime = (*futureEvents)[0].StartDatetime.Add(gm.datasource.GetGameDuration((*futureEvents)[0].Results))
		completedHashes, queuedHash, nextHashSequences = gm.GetNextHashSequences(&(*futureEvents)[0].HashSequenceID, remainingEventsCount)
	} else {
		latestEvent := gm.GetLatestEvent()

		if latestEvent != nil {
			prevEventID = latestEvent.ID
			timeDiff := startDateTime.Sub(latestEvent.StartDatetime)
			defaultDuration := gm.datasource.GetGameDuration(nil)
			startDateTime = latestEvent.StartDatetime.Add(defaultDuration * time.Duration(math.Floor(timeDiff.Seconds()/defaultDuration.Seconds())))
			startDateTime = startDateTime.Add(gm.datasource.GetGameDuration(latestEvent.Results))
			completedHashes, queuedHash, nextHashSequences = gm.GetNextHashSequences(&latestEvent.HashSequenceID, remainingEventsCount)
		} else {
			completedHashes, queuedHash, nextHashSequences = gm.GetNextHashSequences(nil, remainingEventsCount)
		}
	}
	if nextHashSequences == nil {
		return errors.New(gm.GetIdentifier() + " CreateFutureEvents nextHashSequences is nil")
	}
	nextHashSequenceLen := len(*nextHashSequences)

	if nextHashSequenceLen == 0 {
		return errors.New(gm.GetIdentifier() + " CreateFutureEvents nextHashSequences is empty")
	}
	if nextHashSequenceLen < remainingEventsCount {
		logger.Warning(gm.GetIdentifier(), " CreateFutureEvents nextHashSequences contains less count than ", remainingEventsCount)
	}
	createdEvents := []models.Event{}

	if err := db.Shared().Transaction(func(tx *gorm.DB) error {
		for i := 0; i < remainingEventsCount; i++ {
			if i >= nextHashSequenceLen { //if no new hash to generate skip creation of the event
				logger.Warning(gm.GetIdentifier() + " CreateFutureEvents has sequence len is less than events to create")
				continue //skip
			}
			eventResults := gm.datasource.GetHashSequenceResults((*nextHashSequences)[i].Value)

			if eventResults == nil {
				logger.Warning(gm.GetIdentifier() + " CreateFutureEvents eventResults is nil")
				continue //skip
			}
			event := models.Event{
				Name:              gm.datasource.GetEventName(),
				StartDatetime:     startDateTime,
				Status:            constants.EVENT_STATUS_ENABLED,
				GameID:            gm.datasource.GetGameID(),
				HashSequenceID:    (*nextHashSequences)[i].ID,
				SelectionHeaderID: selectionHeader.ID,
				GroundType:        constants.DEFAULT_GROUND_TYPE,
				TableID:           gm.datasource.GetTableID(),
				PrevEventID:       prevEventID,
				MaxBet:            gameTable.MaxBetAmount,
				MaxPayout:         gameTable.MaxPayoutAmount,
				Results:           eventResults,
			}

			if err := tx.Create(&event).Error; err != nil {
				return err
			}
			startDateTime = startDateTime.Add(gm.datasource.GetGameDuration(eventResults)) //start date time for next event
			prevEventID = event.ID
			createdEvents = append(createdEvents, event)
		}

		if completedHashes != nil && len(*completedHashes) > 0 { //update completed hashes
			ids := types.Array[models.Hash](*completedHashes).Map(func(value models.Hash) any { return value.ID })

			if err := tx.Model(models.Hash{}).Where("id IN ?", ids.ToRaw()).
				Updates(models.Hash{Status: constants.HashStatusDone}).Error; err != nil {
				return err
			}
			logger.Info(constants.HashStatusActive, " -> ", constants.HashStatusDone, " hashes: ", ids)
		}

		if queuedHash != nil { //update queued hash
			if err := tx.Model(models.Hash{}).Where("id = ?", queuedHash.ID).
				Updates(models.Hash{Status: constants.HashStatusActive}).Error; err != nil {
				return err
			}
			logger.Info(constants.HashStatusQueued, " -> ", constants.HashStatusActive, " hash: ", queuedHash.ID)
		}
		return nil
	}); err != nil {
		logger.Error(gm.GetIdentifier(), " CreateFutureEvents transaction error: ", err.Error())
		return err
	}
	logger.Info(gm.GetIdentifier(), " CreateFutureEvents created (", len(createdEvents), ")")
	return nil
}

func (gm *gameManager) GetFutureEvents() *[]models.Event {
	events := []models.Event{}

	if err := db.Shared().Preload("Results").
		Where("status = ? AND mini_game_table_id = ? AND start_datetime > NOW()", constants.EVENT_STATUS_ENABLED, gm.datasource.GetTableID()).
		Order("ctime DESC").Find(&events).Error; err != nil {
		logger.Error(gm.GetIdentifier(), " GetFutureEvents error: ", err.Error())
		return nil
	}
	return &events
}

func (gm *gameManager) GetDatasource() Datasource {
	return gm.datasource
}

func (gm *gameManager) GetQueuedHashes() *[]models.Hash {
	hashes := []models.Hash{}

	if err := db.Shared().Preload("Sequences", func(tx *gorm.DB) *gorm.DB {
		return tx.Order("sequence ASC")
	}).Where(models.Hash{
		Status:  constants.HashStatusQueued,
		TableID: gm.datasource.GetTableID(),
	}).Order("ctime ASC").Find(&hashes).Error; err != nil {
		logger.Error(gm.GetIdentifier(), " GetQueuedHashes error: ", err.Error())
		return nil
	}
	return &hashes
}

func (gm *gameManager) GetActiveHashes() *[]models.Hash {
	hashes := []models.Hash{}

	if err := db.Shared().Preload("Sequences", func(tx *gorm.DB) *gorm.DB {
		return tx.Order("sequence ASC")
	}).Where(models.Hash{
		Status:  constants.HashStatusActive,
		TableID: gm.datasource.GetTableID(),
	}).Order("ctime ASC").Find(&hashes).Error; err != nil {
		logger.Error(gm.GetIdentifier(), " GetQueuedHashes error: ", err.Error())
		return nil
	}
	return &hashes
}

// returns completed hashes(to done), queued hash(to active) and next hash sequences
func (gm *gameManager) GetNextHashSequences(currentSequenceID *int64, nextCount int) (*[]models.Hash, *models.Hash, *[]models.HashSequence) {
	nextHashSequences := []models.HashSequence{}
	activeHashes := gm.GetActiveHashes()

	if activeHashes == nil {
		return nil, nil, nil
	}
	logger.Debug(gm.GetIdentifier(), " LoadHashSequences from activeHashes (", len(*activeHashes), ")")
	completedHashes := (*[]models.Hash)(nil)

	if cHashes, hashSequences := gm.LoadHashSequences(currentSequenceID, activeHashes, nextCount); hashSequences != nil {
		completedHashes = cHashes
		nextHashSequences = append(nextHashSequences, *hashSequences...)
	}
	if len(nextHashSequences) >= nextCount {
		return completedHashes, nil, &nextHashSequences
	}
	queuedHashes := gm.GetQueuedHashes()

	if queuedHashes == nil {
		logger.Error(gm.GetIdentifier(), " GetNextHashSequences queuedHashes is nil")
		return completedHashes, nil, &nextHashSequences
	}
	logger.Debug(gm.GetIdentifier(), " LoadHashSequences from queuedHashes (", len(*queuedHashes), ")")
	queuedHash := (*models.Hash)(nil)

	if len(*queuedHashes) > 0 {
		queuedHash = &(*queuedHashes)[0]
	}
	if _, hashSequences := gm.LoadHashSequences(nil, queuedHashes, nextCount-len(nextHashSequences)); hashSequences != nil {
		nextHashSequences = append(nextHashSequences, *hashSequences...)
	}
	return completedHashes, queuedHash, &nextHashSequences
}

// returns completed hashes and next sequences
func (gm *gameManager) LoadHashSequences(currentSequenceID *int64, hashes *[]models.Hash, nextCount int) (*[]models.Hash, *[]models.HashSequence) {
	if hashes == nil {
		return nil, nil
	}
	completedHashes := []models.Hash{}
	hashSequences := []models.HashSequence{}

hashes:
	for hi := 0; hi < len(*hashes); hi++ {
		hash := (*hashes)[hi]
		sequencesLen := len(*hash.Sequences)

	sequences:
		for si := 0; si < sequencesLen; si++ {
			if currentSequenceID == nil { // if nil append starting sequence
				currentSequenceID = &(*hash.Sequences)[si].ID
				hashSequences = append(hashSequences, (*hash.Sequences)[si])
			}
			if sequencesLen >= (si+2) && (*hash.Sequences)[si].ID == *currentSequenceID && ((*hash.Sequences)[si].Sequence+1) == (*hash.Sequences)[si+1].Sequence {
				hashSequence := (*hash.Sequences)[si+1]
				*currentSequenceID = hashSequence.ID
				hashSequences = append(hashSequences, hashSequence)

				if hashSequence.Sequence >= gm.datasource.GetMaxSequencePerHash() {
					break sequences
				}
				if len(hashSequences) >= nextCount {
					break hashes
				}
			}
		}
		completedHashes = append(completedHashes, hash)
	}

	return &completedHashes, &hashSequences
}

func (gm *gameManager) GetGameTable() *models.GameTable {
	gameTable := models.GameTable{
		ID:     gm.datasource.GetTableID(),
		GameID: gm.datasource.GetGameID(),
	}

	if err := db.Shared().Where(gameTable).Order("ctime DESC").First(&gameTable).Error; err != nil {
		logger.Error(gm.GetIdentifier(), " GetGameTable error: ", err.Error())
		return nil
	}

	return &gameTable
}

func (gm *gameManager) GetSelectionHeader() *models.SelectionHeader {
	selectionHeader := models.SelectionHeader{
		GameID: gm.datasource.GetGameID(),
		Status: "active",
	}

	if err := db.Shared().Preload("SelectionLines").Where(selectionHeader).Order("ctime DESC").First(&selectionHeader).Error; err != nil {
		logger.Error(gm.GetIdentifier(), " GetSelectionHeader error: ", err.Error())
		return nil
	}
	return &selectionHeader
}

func (gm *gameManager) GetLatestEvent() *models.Event {
	event := models.Event{}

	if err := db.Shared().Preload("Results").Where("mini_game_table_id = ?", gm.datasource.GetTableID()).
		Order("ctime DESC").
		First(&event).Error; err != nil {
		logger.Error(gm.GetIdentifier(), " GetLatestEvent error: ", err.Error())
		return nil
	}
	return &event
}

func (gm *gameManager) GenerateStartDateTime(fromTime time.Time) time.Time {
	return time.Now()
}

func (gm *gameManager) GetIdentifier() string {
	return gm.datasource.GetIdentifier()
}
