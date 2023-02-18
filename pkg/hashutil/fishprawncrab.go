package hashutil

import (
	"fmt"

	"bitbucket.org/esportsph/minigame-backend-golang/pkg/types"
)

func FishPrawnCrabGenerateResult(hash Hash) (string, error) {
	offsetHash := hash.Extract(0, 8)

	if err := offsetHash.Error(); err != nil {
		return "", err
	}
	offsetValue := int(offsetHash.Int64() % 8)
	dice1Hash := hash.Extract(offsetValue, 8)

	if err := dice1Hash.Error(); err != nil {
		return "", err
	}
	dice1Value := int8(dice1Hash.Int64()%6) + 1
	dice2Hash := hash.Extract(offsetValue+8, 8)

	if err := dice2Hash.Error(); err != nil {
		return "", err
	}
	dice2Value := int8(dice2Hash.Int64()%6) + 1
	dice3Hash := hash.Extract(offsetValue+16, 8)

	if err := dice3Hash.Error(); err != nil {
		return "", err
	}
	dice3Value := int8(dice3Hash.Int64()%6) + 1

	return fmt.Sprintf("%v,%v,%v", GetSelectionName(dice1Value), GetSelectionName(dice2Value), GetSelectionName(dice3Value)), nil
}

func GetSelectionName(value int8) string {
	switch value {
	case 1:
		return "tiger"
	case 2:
		return "ground"
	case 3:
		return "chicken"
	case 4:
		return "fish"
	case 5:
		return "prawn"
	case 6:
		return "crab"
	default:
		panic("fishprawncrab GetSelectionName value " + types.Int(value).String() + " is not supported")
	}
}
