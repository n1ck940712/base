package migrations

import (
	constants_fifashootup "bitbucket.org/esportsph/minigame-backend-golang/internal/constants/fifashootup"
	constants_fishprawncrab "bitbucket.org/esportsph/minigame-backend-golang/internal/constants/fishprawncrab"
	constants_lolcouple "bitbucket.org/esportsph/minigame-backend-golang/internal/constants/lolcouple"
	constants_loltower "bitbucket.org/esportsph/minigame-backend-golang/internal/constants/loltower"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/db"
	migrations_insert "bitbucket.org/esportsph/minigame-backend-golang/internal/migrations/insert"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/models"
	"gorm.io/gorm"
)

type insertDatasource struct {
}

func (tds *insertDatasource) GetDB() *gorm.DB {
	return db.Shared()
}

func (m *Migrations) StartInsertMigrations() {
	insert := migrations_insert.NewInsert(&insertDatasource{})
	insert.InsertGames(
		models.Game{
			ESID:      constants_lolcouple.ESGameID,
			Name:      constants_lolcouple.GameName,
			ShortName: constants_lolcouple.GameName,
			Type:      "mini_game",
			GameTables: &[]models.GameTable{
				{
					Name:            constants_lolcouple.GameName,
					MinBetAmount:    10,
					MaxBetAmount:    5000,
					MaxPayoutAmount: 15000,
					IsEnabled:       true,
				},
			},
		},
		models.Game{
			ESID:      constants_fifashootup.ESGameID,
			Name:      constants_fifashootup.GameName,
			ShortName: constants_fifashootup.GameName,
			Type:      "mini_game",
			GameTables: &[]models.GameTable{
				{
					Name:            constants_fifashootup.GameName,
					MinBetAmount:    10,
					MaxBetAmount:    3000,
					MaxPayoutAmount: 12000,
					IsEnabled:       true,
				},
			},
		},
		models.Game{
			ESID:      constants_fishprawncrab.ESGameID,
			Name:      constants_fishprawncrab.GameName,
			ShortName: constants_fishprawncrab.GameName,
			Type:      "mini_game",
			GameTables: &[]models.GameTable{
				{
					Name:            constants_fishprawncrab.GameName,
					MinBetAmount:    10,
					MaxBetAmount:    3000,
					MaxPayoutAmount: 12000,
					IsEnabled:       true,
				},
			},
		},
	)

	insert.InsertSelections(
		models.SelectionHeader{
			Name:   constants_loltower.GameName,
			Status: "active",
			GameID: constants_loltower.GameID,
			SelectionLines: &[]models.SelectionLine{
				{
					Name:       *constants_loltower.Selection1.String().Ptr(),
					Attributes: "{}",
				},
				{
					Name:       *constants_loltower.Selection2.String().Ptr(),
					Attributes: "{}",
				},
				{
					Name:       *constants_loltower.Selection3.String().Ptr(),
					Attributes: "{}",
				},
				{
					Name:       *constants_loltower.Selection4.String().Ptr(),
					Attributes: "{}",
				},
				{
					Name:       *constants_loltower.Selection5.String().Ptr(),
					Attributes: "{}",
				},
			},
		},
		models.SelectionHeader{
			Name:   constants_lolcouple.GameName,
			Status: "active",
			GameID: constants_lolcouple.GameID,
			SelectionLines: &[]models.SelectionLine{
				{
					Name:       constants_lolcouple.Selection1,
					Attributes: "{}",
				},
				{
					Name:       constants_lolcouple.Selection2,
					Attributes: "{}",
				},
				{
					Name:       constants_lolcouple.Selection3,
					Attributes: "{}",
				},
				{
					Name:       constants_lolcouple.Selection4,
					Attributes: "{}",
				},
				{
					Name:       constants_lolcouple.Selection5,
					Attributes: "{}",
				},
				{
					Name:       constants_lolcouple.Selection6,
					Attributes: "{}",
				},
				{
					Name:       constants_lolcouple.Selection7,
					Attributes: "{}",
				},
				{
					Name:       constants_lolcouple.Selection8,
					Attributes: "{}",
				},
				{
					Name:       constants_lolcouple.Selection9,
					Attributes: "{}",
				},
				{
					Name:       constants_lolcouple.Selection10,
					Attributes: "{}",
				},
			},
		},
		models.SelectionHeader{
			Name:   constants_fifashootup.GameName,
			Status: "active",
			GameID: constants_fifashootup.GameID,
			SelectionLines: &[]models.SelectionLine{
				{
					Name:       constants_fifashootup.Selection1,
					Attributes: "{}",
				},
				{
					Name:       constants_fifashootup.Selection2,
					Attributes: "{}",
				},
				{
					Name:       constants_fifashootup.Selection3,
					Attributes: "{}",
				},
			},
		},
		models.SelectionHeader{
			Name:   constants_fishprawncrab.GameName,
			Status: "active",
			GameID: constants_fishprawncrab.GameID,
			SelectionLines: &[]models.SelectionLine{
				{
					Name:       constants_fishprawncrab.Selection1,
					Attributes: "{}",
				},
				{
					Name:       constants_fishprawncrab.Selection2,
					Attributes: "{}",
				},
				{
					Name:       constants_fishprawncrab.Selection3,
					Attributes: "{}",
				},
				{
					Name:       constants_fishprawncrab.Selection4,
					Attributes: "{}",
				},
				{
					Name:       constants_fishprawncrab.Selection5,
					Attributes: "{}",
				},
				{
					Name:       constants_fishprawncrab.Selection6,
					Attributes: "{}",
				},
			},
		},
	)

	insert.InsertTestUsers(
		//loltower
		migrations_insert.TestUser{ESID: 114258, MemberCode: "testjus", TableID: constants_loltower.TableID},
		migrations_insert.TestUser{ESID: 114097, MemberCode: "testnhick", TableID: constants_loltower.TableID},
		migrations_insert.TestUser{ESID: 142174, MemberCode: "testmike", TableID: constants_loltower.TableID},
		migrations_insert.TestUser{ESID: 55555, MemberCode: "testcarlo", TableID: constants_loltower.TableID},
		migrations_insert.TestUser{ESID: 101832, MemberCode: "testgene", TableID: constants_loltower.TableID},
		migrations_insert.TestUser{ESID: 101899, MemberCode: "testpat", TableID: constants_loltower.TableID},
		migrations_insert.TestUser{ESID: 104436, MemberCode: "testkeir", TableID: constants_loltower.TableID},
		migrations_insert.TestUser{ESID: 55554, MemberCode: "testjoe", TableID: constants_loltower.TableID},
		migrations_insert.TestUser{ESID: 101833, MemberCode: "testaina", TableID: constants_loltower.TableID},
		migrations_insert.TestUser{ESID: 118097, MemberCode: "testgarry", TableID: constants_loltower.TableID},
		migrations_insert.TestUser{ESID: 166255, MemberCode: "testjomarie", TableID: constants_loltower.TableID},
		migrations_insert.TestUser{ESID: 177302, MemberCode: "testpoorman", TableID: constants_loltower.TableID},
		migrations_insert.TestUser{ESID: 177316, MemberCode: "testgarryuttog", TableID: constants_loltower.TableID},
		//lolcouple
		migrations_insert.TestUser{ESID: 114258, MemberCode: "testjus", TableID: constants_lolcouple.TableID},
		migrations_insert.TestUser{ESID: 114097, MemberCode: "testnhick", TableID: constants_lolcouple.TableID},
		migrations_insert.TestUser{ESID: 142174, MemberCode: "testmike", TableID: constants_lolcouple.TableID},
		migrations_insert.TestUser{ESID: 55555, MemberCode: "testcarlo", TableID: constants_lolcouple.TableID},
		migrations_insert.TestUser{ESID: 101832, MemberCode: "testgene", TableID: constants_lolcouple.TableID},
		migrations_insert.TestUser{ESID: 101899, MemberCode: "testpat", TableID: constants_lolcouple.TableID},
		migrations_insert.TestUser{ESID: 104436, MemberCode: "testkeir", TableID: constants_lolcouple.TableID},
		migrations_insert.TestUser{ESID: 55554, MemberCode: "testjoe", TableID: constants_lolcouple.TableID},
		migrations_insert.TestUser{ESID: 101833, MemberCode: "testaina", TableID: constants_lolcouple.TableID},
		migrations_insert.TestUser{ESID: 118097, MemberCode: "testgarry", TableID: constants_lolcouple.TableID},
		migrations_insert.TestUser{ESID: 166255, MemberCode: "testjomarie", TableID: constants_lolcouple.TableID},
		migrations_insert.TestUser{ESID: 177302, MemberCode: "testpoorman", TableID: constants_lolcouple.TableID},
		migrations_insert.TestUser{ESID: 177316, MemberCode: "testgarryuttog", TableID: constants_lolcouple.TableID},
		migrations_insert.TestUser{ESID: 152826, MemberCode: "testjack", TableID: constants_lolcouple.TableID},
		//fifashootup
		migrations_insert.TestUser{ESID: 114258, MemberCode: "testjus", TableID: constants_fifashootup.TableID},
		migrations_insert.TestUser{ESID: 114097, MemberCode: "testnhick", TableID: constants_fifashootup.TableID},
		migrations_insert.TestUser{ESID: 142174, MemberCode: "testmike", TableID: constants_fifashootup.TableID},
		migrations_insert.TestUser{ESID: 55555, MemberCode: "testcarlo", TableID: constants_fifashootup.TableID},
		migrations_insert.TestUser{ESID: 101832, MemberCode: "testgene", TableID: constants_fifashootup.TableID},
		migrations_insert.TestUser{ESID: 101899, MemberCode: "testpat", TableID: constants_fifashootup.TableID},
		migrations_insert.TestUser{ESID: 104436, MemberCode: "testkeir", TableID: constants_fifashootup.TableID},
		migrations_insert.TestUser{ESID: 55554, MemberCode: "testjoe", TableID: constants_fifashootup.TableID},
		migrations_insert.TestUser{ESID: 101833, MemberCode: "testaina", TableID: constants_fifashootup.TableID},
		migrations_insert.TestUser{ESID: 118097, MemberCode: "testgarry", TableID: constants_fifashootup.TableID},
		migrations_insert.TestUser{ESID: 166255, MemberCode: "testjomarie", TableID: constants_fifashootup.TableID},
		migrations_insert.TestUser{ESID: 177302, MemberCode: "testpoorman", TableID: constants_fifashootup.TableID},
		migrations_insert.TestUser{ESID: 177316, MemberCode: "testgarryuttog", TableID: constants_fifashootup.TableID},
		migrations_insert.TestUser{ESID: 152826, MemberCode: "testjack", TableID: constants_fifashootup.TableID},
		//fishprawncrab
		migrations_insert.TestUser{ESID: 114258, MemberCode: "testjus", TableID: constants_fishprawncrab.TableID},
		migrations_insert.TestUser{ESID: 114097, MemberCode: "testnhick", TableID: constants_fishprawncrab.TableID},
		migrations_insert.TestUser{ESID: 142174, MemberCode: "testmike", TableID: constants_fishprawncrab.TableID},
		migrations_insert.TestUser{ESID: 55555, MemberCode: "testcarlo", TableID: constants_fishprawncrab.TableID},
		migrations_insert.TestUser{ESID: 101832, MemberCode: "testgene", TableID: constants_fishprawncrab.TableID},
		migrations_insert.TestUser{ESID: 101899, MemberCode: "testpat", TableID: constants_fishprawncrab.TableID},
		migrations_insert.TestUser{ESID: 104436, MemberCode: "testkeir", TableID: constants_fishprawncrab.TableID},
		migrations_insert.TestUser{ESID: 55554, MemberCode: "testjoe", TableID: constants_fishprawncrab.TableID},
		migrations_insert.TestUser{ESID: 101833, MemberCode: "testaina", TableID: constants_fishprawncrab.TableID},
		migrations_insert.TestUser{ESID: 118097, MemberCode: "testgarry", TableID: constants_fishprawncrab.TableID},
		migrations_insert.TestUser{ESID: 166255, MemberCode: "testjomarie", TableID: constants_fishprawncrab.TableID},
		migrations_insert.TestUser{ESID: 177302, MemberCode: "testpoorman", TableID: constants_fishprawncrab.TableID},
		migrations_insert.TestUser{ESID: 177316, MemberCode: "testgarryuttog", TableID: constants_fishprawncrab.TableID},
		migrations_insert.TestUser{ESID: 152826, MemberCode: "testjack", TableID: constants_fishprawncrab.TableID},
	)
}
