package migrations_insert

import (
	"fmt"

	"bitbucket.org/esportsph/minigame-backend-golang/internal/logger"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/models"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/settings"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/utils"
	"gorm.io/gorm"
)

type TestUser struct {
	ESID       int64
	MemberCode string
	TableID    int64
}

type TestGame struct {
	Name            string
	ShortName       string
	MinBetAmount    float64
	MaxBetAmount    float64
	MaxPayoutAmount float64
}

type Datasource interface {
	GetDB() *gorm.DB
}

type Insert interface {
	InsertTestUsers(users ...TestUser)
	InsertGames(games ...models.Game)
	InsertSelections(selections ...models.SelectionHeader)
}

type insert struct {
	datasource Datasource
}

func NewInsert(datasource Datasource) Insert {
	return &insert{datasource: datasource}
}

// inserts in users if not existing and insert to members if not existing
func (im *insert) InsertTestUsers(users ...TestUser) {
	if settings.GetLoggerLevel().String() != "local" {
		logger.Info("InsertTestUsers is only supported on local")
		return
	}
	for i := 0; i < len(users); i++ {
		user := models.User{EsportsID: users[i].ESID}

		if im.datasource.GetDB().Where(user).Find(&user).RowsAffected == 0 {
			newUser := models.User{
				Username:         fmt.Sprint("esports-2-", users[i].ESID),
				UserType:         2,
				EsportsID:        users[i].ESID,
				EsportsPartnerID: 2,
				MemberCode:       users[i].MemberCode,
				CurrencyCode:     "RMB",
				ExchangeRate:     1.000000,
				SleepStatus:      0,
			}

			if err := im.datasource.GetDB().Create(&newUser).Error; err != nil {
				logger.Error("InsertTestUsers Create user error: ", err.Error())
			} else {
				logger.Info("created user: ", utils.PrettyJSON(newUser))
			}
		}
		member := models.MemberTable{UserID: user.ID, TableID: users[i].TableID}

		if im.datasource.GetDB().Where(member).Find(&member).RowsAffected == 0 {
			newMember := models.MemberTable{UserID: user.ID, TableID: users[i].TableID}

			if err := im.datasource.GetDB().Create(&newMember).Error; err != nil {
				logger.Error("InsertTestUsers Create member error: ", err.Error())
			} else {
				logger.Info("created member: ", utils.PrettyJSON(newMember), " - ", users[i].TableID)
			}
		} else {
			logger.Info("Member: (", users[i].MemberCode, " - ", users[i].TableID, ")  already existing")
		}
	}
}

// inserts in selection headers if not existing
func (im *insert) InsertSelections(selections ...models.SelectionHeader) {
	for i := 0; i < len(selections); i++ {
		if im.datasource.GetDB().Where(selections[i]).Find(&selections[i]).RowsAffected == 0 {
			if err := im.datasource.GetDB().Create(&selections[i]).Error; err != nil {
				logger.Info("to insert Selection header: ", utils.PrettyJSON(selections[i]))
				logger.Error("Insert Selection header: (", selections[i].Name, ") error: ", err.Error())
			} else {
				logger.Info("Inserted Selection header: (", selections[i].Name, ")")
			}
		} else {
			logger.Info("Selection header: (", selections[i].Name, ") already existing")
		}
	}
}

func (im *insert) InsertGames(games ...models.Game) {
	if settings.GetEnvironment().String() != "local" {
		logger.Info("InsertGames is only supported on local")
		return
	}
	for i := 0; i < len(games); i++ {
		if im.datasource.GetDB().Where(games[i]).Find(&games[i]).RowsAffected == 0 {
			if err := im.datasource.GetDB().Create(&games[i]).Error; err != nil {
				logger.Error("Insert Game: (", games[i].Name, ") error: ", err.Error())
				continue
			}
			logger.Info("Inserted Game: (", games[i].Name, ")")
		} else {
			logger.Info("Game: (", games[i].Name, ") already existing")
		}
	}
}
