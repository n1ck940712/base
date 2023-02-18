package process_bet

import (
	"crypto/sha256"
	"fmt"
	"math"
	"strings"
	"time"

	"bitbucket.org/esportsph/minigame-backend-golang/internal/constants"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/models"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/models/request"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/models/response"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/settings"
	"bitbucket.org/esportsph/minigame-backend-golang/pkg/types"
	"github.com/google/uuid"
)

const (
	BetType       = "bet"
	SelectionType = "selection"
)

type BetResult struct {
	Result                 int16
	WinLossAmount          float64
	PayoutAmount           float64
	HongkongOdds           float64
	EuroOdds               float64
	MalayOdds              float64
	OriginalOdds           float64
	PossibleWinningsAmount float64
}

type BetOpenRange struct {
	MinMS int64
	MaxMS int64
}

type BetDatasouce interface {
	GetIdentifier() string
	GetUser() *models.User
	GetEvent() *models.Event
	GetEventResults() *[]models.EventResult
	GetGameTable() *models.GameTable
	GetMemberTable() *models.MemberTable
}

type BetProcess interface {
	Placebet(betData *request.BetData) response.ResponseData
}

func GenerateTicketID(esportsID int64) string {
	now := time.Now()
	curYear := now.Year()
	yearTensValue := types.Int(math.Floor(float64(curYear%100) / float64(10)))
	yearOnesValue := types.Int(curYear % 10)
	yr1 := yearTensValue.GetASCII()
	yr2 := yearOnesValue.GetASCII()
	mo1 := types.Int(now.Month() + 12).GetASCII()

	dayTensValue := types.Int(math.Floor(float64(now.Day()) / float64(10)))
	dayOnesValue := types.Int(now.Day() % 10)
	day1 := dayTensValue.GetASCII()
	day2 := dayOnesValue.GetASCII()
	hr := types.Int(now.Hour()).GetASCII()
	m := types.Int(now.Minute()).LeadingZeroes(2)
	s := types.Int(now.Second()).LeadingZeroes(2)

	r := types.Int(4).RandSeq()
	baseTicketID := fmt.Sprintf("%v%v%v%v%v%v%v%v%vMG%v", yr1, yr2, mo1, day1, day2, esportsID, hr, m, s, r)
	signature := fmt.Sprintf("%x", sha256.Sum256([]byte(baseTicketID+settings.GetSecretKey().String())))

	return strings.ToUpper(baseTicketID + signature[0:4])
}

func GenerateReferenceNo(referenceNo string) string {
	if len(referenceNo) > 0 {
		return referenceNo
	}
	return uuid.NewString()
}

func TimeNow() time.Time {
	return time.Now()
}

func DefaultCreatedTicketStatus() int16 {
	return constants.TICKET_STATUS_PAYMENT_CONFIRMED
}
