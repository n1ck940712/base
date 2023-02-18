package process_bet

import (
	"strings"

	constants_fifashootup "bitbucket.org/esportsph/minigame-backend-golang/internal/constants/fifashootup"
	"bitbucket.org/esportsph/minigame-backend-golang/internal/utils"
	"bitbucket.org/esportsph/minigame-backend-golang/pkg/types"
)

func GenerateResult(leftCard string, rightCard string, resultCard string) string {
	leftRange, middleRange, rightRange := GenerateLMRRanges(leftCard, rightCard)

	if types.Array[string](leftRange).Constains(strings.ToUpper(resultCard)) {
		return constants_fifashootup.Selection1
	} else if types.Array[string](middleRange).Constains(strings.ToUpper(resultCard)) {
		return constants_fifashootup.Selection2
	} else if types.Array[string](rightRange).Constains(strings.ToUpper(resultCard)) {
		return constants_fifashootup.Selection3
	}
	return ""
}

func GenerateLMRRanges(leftCard string, rightCard string) (leftRange []string, middleRange []string, rightRange []string) {
	if leftCard == rightCard {
		switch leftCard {
		case "a":
			value := GenerateCardInt(leftCard)

			return GenerateRange(value, value), []string{}, GenerateRange(value+1, GenerateCardInt("k"))
		case "k":
			value := GenerateCardInt(leftCard)

			return GenerateRange(GenerateCardInt("a"), value-1), []string{}, GenerateRange(value, value)
		default:
			value := GenerateCardInt(leftCard)

			return GenerateRange(GenerateCardInt("a"), value-1), GenerateRange(value, value), GenerateRange(value+1, GenerateCardInt("k"))
		}
	}
	leftCardValue := GenerateCardInt(leftCard)
	rightCardValue := GenerateCardInt(rightCard)
	if (rightCardValue - leftCardValue) == 1 {
		return GenerateRange(GenerateCardInt("a"), leftCardValue), []string{}, GenerateRange(rightCardValue, GenerateCardInt("k"))
	}
	return GenerateRange(GenerateCardInt("a"), leftCardValue), GenerateRange(leftCardValue+1, rightCardValue-1), GenerateRange(rightCardValue, GenerateCardInt("k"))
}

func GenerateCardInt(card string) int8 {
	switch card {
	case "a":
		return 1
	case "j":
		return 11
	case "q":
		return 12
	case "k":
		return 13
	default:
		return types.String(card).Int().Int8()
	}
}

func GenerateCardString(value int8) string {
	switch value {
	case 1:
		return "a"
	case 11:
		return "j"
	case 12:
		return "q"
	case 13:
		return "k"
	default:
		return string(types.Int(value).String())
	}
}

func GenerateOdds(cRange []string) types.Odds {
	switch len(cRange) {
	case 0:
		return 0
	case 1:
		return constants_fifashootup.Card1Odds
	case 2:
		return constants_fifashootup.Card2Odds
	case 3:
		return constants_fifashootup.Card3Odds
	case 4:
		return constants_fifashootup.Card4Odds
	case 5:
		return constants_fifashootup.Card5Odds
	case 6:
		return constants_fifashootup.Card6Odds
	case 7:
		return constants_fifashootup.Card7Odds
	case 8:
		return constants_fifashootup.Card8Odds
	case 9:
		return constants_fifashootup.Card9Odds
	case 10:
		return constants_fifashootup.Card10Odds
	case 11:
		return constants_fifashootup.Card11Odds
	case 12:
		return constants_fifashootup.Card12Odds
	default:
		panic("invalid card range" + utils.PrettyJSON(cRange))
	}
}

func GenerateRange(min int8, max int8) []string {
	itemRange := []string{}
	for i := min; i <= max; i++ {
		itemRange = append(itemRange, strings.ToUpper(GenerateCardString(i)))
	}
	return itemRange
}
