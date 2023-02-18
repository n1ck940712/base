package migrations_insert

import (
	"testing"

	constants_loltower "bitbucket.org/esportsph/minigame-backend-golang/internal/constants/loltower"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/db"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/models"
	"gorm.io/gorm"
)

type testDatasource struct {
}

func (tds *testDatasource) GetDB() *gorm.DB {
	return db.Shared()
}

func TestInsert(t *testing.T) {
	insert := NewInsert(&testDatasource{})

	insert.InsertSelections(
		models.SelectionHeader{
			Name:   "LOL Tower",
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
	)

	insert.InsertGames(
		models.Game{
			ESID:      34,
			Name:      "LOL Couple",
			ShortName: "LOL Couple",
			Type:      "mini_game",
			GameTables: &[]models.GameTable{
				{
					Name:            "LOL Couple",
					MinBetAmount:    10,
					MaxBetAmount:    5000,
					MaxPayoutAmount: 15000,
					IsEnabled:       true,
				},
			},
		},
	)

	insert.InsertTestUsers(
		TestUser{ESID: 114258, MemberCode: "testjus", TableID: constants_loltower.TableID},
		TestUser{ESID: 114097, MemberCode: "testnhick", TableID: constants_loltower.TableID},
		TestUser{ESID: 142174, MemberCode: "testmike", TableID: constants_loltower.TableID},
		TestUser{ESID: 55555, MemberCode: "testcarlo", TableID: constants_loltower.TableID},
		TestUser{ESID: 101832, MemberCode: "testgene", TableID: constants_loltower.TableID},
		TestUser{ESID: 101899, MemberCode: "testpat", TableID: constants_loltower.TableID},
		TestUser{ESID: 104436, MemberCode: "testkeir", TableID: constants_loltower.TableID},
		TestUser{ESID: 55554, MemberCode: "testjoe", TableID: constants_loltower.TableID},
		TestUser{ESID: 101833, MemberCode: "testaina", TableID: constants_loltower.TableID},
		TestUser{ESID: 118097, MemberCode: "testgarry", TableID: constants_loltower.TableID},
		TestUser{ESID: 166255, MemberCode: "testjomarie", TableID: constants_loltower.TableID},
		TestUser{ESID: 177302, MemberCode: "testpoorman", TableID: constants_loltower.TableID},
		TestUser{ESID: 177316, MemberCode: "testgarryuttog", TableID: constants_loltower.TableID},
	)
}
