package hashutil

import (
	"fmt"

	"bitbucket.org/esportsph/minigame-backend-golang/pkg/types"
)

func FIFAShootupGenerateResult(hash Hash) (selections string, result string, err error) {
	_selections := []([]int8){}
	offsetHash := hash.Extract(0, 4)

	if err := offsetHash.Error(); err != nil {
		return "", "", err
	}
	offsetValue := int(offsetHash.Int64()%4) + 4

	for i := 0; i < 3; i++ {
		cardHash := hash.Extract(offsetValue+i*8, 8)

		if err := cardHash.Error(); err != nil {
			return "", "", err
		}
		cardValue := int8(cardHash.Int64()%13) + 1
		suitHash := hash.Extract(offsetValue+(i+3)*8, 8)

		if err := suitHash.Error(); err != nil {
			return "", "", err
		}
		suitValue := int8(suitHash.Int64()%4) + 1

		for types.Array[[]int8](_selections).Constains([]int8{cardValue, suitValue}) {
			suitValue = (suitValue % 4) + 1
		}
		_selections = append(_selections, []int8{cardValue, suitValue})
	}
	if _selections[0][0] > _selections[1][0] {
		_selections[0], _selections[1] = _selections[1], _selections[0]
	}
	selections = fmt.Sprintf("%v-%v,%v-%v,%v-%v", fifaShootupCard(_selections[0][0]), fifaShootupSuit(_selections[0][1]), fifaShootupCard(_selections[1][0]), fifaShootupSuit(_selections[1][1]), fifaShootupCard(_selections[2][0]), fifaShootupSuit(_selections[2][1]))
	result = fifaShootupGenerateResult(_selections[0][0], _selections[1][0], _selections[2][0])

	return selections, result, nil
}

func fifaShootupCard(card int8) string {
	switch card {
	case 1:
		return "a"
	case 2, 3, 4, 5, 6, 7, 8, 9, 10:
		return fmt.Sprint(card)
	case 11:
		return "j"
	case 12:
		return "q"
	case 13:
		return "k"
	default:
		panic("unsupported card " + fmt.Sprint(card))
	}
}

func fifaShootupSuit(suit int8) string {
	switch suit {
	case 1:
		return "spades"
	case 2:
		return "hearts"
	case 3:
		return "clubs"
	case 4:
		return "diamonds"
	default:
		panic("unsupported suit " + fmt.Sprint(suit))
	}
}

func fifaShootupGenerateResult(leftValue int8, rightValue int8, resultValue int8) string {
	leftRange, middleRange, rightRange := fifaShootupGenerateLMRRanges(leftValue, rightValue)

	if types.Array[int8](leftRange).Constains(resultValue) {
		return "left"
	} else if types.Array[int8](middleRange).Constains(resultValue) {
		return "middle"
	} else if types.Array[int8](rightRange).Constains(resultValue) {
		return "right"
	}
	return ""
}

func fifaShootupGenerateLMRRanges(leftValue int8, rightValue int8) (leftRange []int8, middleRange []int8, rightRange []int8) {
	if leftValue == rightValue {
		switch leftValue {
		case 1:
			return fifaShootupGenerateRange(leftValue, leftValue), []int8{}, fifaShootupGenerateRange(leftValue+1, 13)
		case 13:
			return fifaShootupGenerateRange(1, leftValue-1), []int8{}, fifaShootupGenerateRange(leftValue, leftValue)
		default:
			return fifaShootupGenerateRange(1, leftValue-1), fifaShootupGenerateRange(leftValue, leftValue), fifaShootupGenerateRange(leftValue+1, 13)
		}
	}
	if (rightValue - leftValue) == 1 {
		return fifaShootupGenerateRange(1, leftValue), []int8{}, fifaShootupGenerateRange(rightValue, 13)
	}
	return fifaShootupGenerateRange(1, leftValue), fifaShootupGenerateRange(leftValue+1, rightValue-1), fifaShootupGenerateRange(rightValue, 13)
}

func fifaShootupGenerateRange(min int8, max int8) []int8 {
	valueRange := []int8{}
	for i := min; i <= max; i++ {
		valueRange = append(valueRange, i)
	}
	return valueRange
}
