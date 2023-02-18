package hashutil

import (
	"strconv"

	"bitbucket.org/esportsph/minigame-backend-golang/pkg/types"
)

func LOLTowerGenerateResult(hash Hash) (string, error) {
	if (hash.Extract(0, 8).Int64() % 2) == 1 {
		return lolTowerGenerateResultFromBomb(hash.Extract(8, 56))
	}
	return lolTowerGenerateResultFromWin(hash.Extract(8, 56))
}

func lolTowerGenerateResultFromBomb(hash Hash) (string, error) {
	cards := types.Array[int]{1, 2, 3, 4, 5}

	for i := 0; i < 2; i++ {
		eHash := hash.Extract(i*8, 8)
		popIndex := int(eHash.Int64()) % len(cards)

		if err := eHash.Error(); err != nil {
			return "", err
		}
		cards.PopIndex(popIndex)
	}
	return cards.Join(","), nil
}

func lolTowerGenerateResultFromWin(hash Hash) (string, error) {
	sCards := types.Array[int]{1, 2, 3, 4, 5}
	cards := types.Array[int]{}

	for i := 0; i < 3; i++ {
		eHash := hash.Extract(i*8, 8)
		popIndex := int(eHash.Int64()) % len(sCards)

		if err := eHash.Error(); err != nil {
			return "", err
		}

		if popValue := sCards.PopIndex(popIndex); popValue != nil {
			cards = append(cards, *popValue)
		}
	}

	return cards.Join(","), nil
}

// old implementation
func lolTowerGenerateHashResult(hash string) []int {
	hexFirst8Str := hash[0:8] //used first8 to check result path
	hexFirst8Int, _ := strconv.ParseInt(hexFirst8Str, 16, 64)

	if (hexFirst8Int % 2) == 1 {
		return lolTowerGenerateHashResultFromBomb(hash,
			SliceIndex{PrevIndex: 8, PostIndex: 16},  //bomb1
			SliceIndex{PrevIndex: 16, PostIndex: 24}, //bomb2
		)
	}
	return lolTowerGenerateHashResultFromWin(hash,
		SliceIndex{PrevIndex: 8, PostIndex: 16},  //result1
		SliceIndex{PrevIndex: 16, PostIndex: 24}, //result2
		SliceIndex{PrevIndex: 24, PostIndex: 32}, //result3
	)
}

func lolTowerGenerateHashResultFromBomb(hash string, bomb1 SliceIndex, bomb2 SliceIndex) []int {
	cards := []int{1, 2, 3, 4, 5}

	hexNum := hash[bomb1.PrevIndex:bomb1.PostIndex]
	dec, _ := strconv.ParseInt(hexNum, 16, 64)
	res1 := int((dec % 5))
	cards = append(cards[:res1], cards[res1+1:]...)

	hexNum2 := hash[bomb2.PrevIndex:bomb2.PostIndex]
	dec2, _ := strconv.ParseInt(hexNum2, 16, 64)
	res2 := int((dec2 % 4))
	cards = append(cards[:res2], cards[res2+1:]...)

	return cards
}

func lolTowerGenerateHashResultFromWin(hash string, results ...SliceIndex) []int {
	cards := types.Array[int]{1, 2, 3, 4, 5}
	rCards := []int{}

	for i := 0; i < len(results); i++ {
		hexStr := hash[results[i].PrevIndex:results[i].PostIndex]
		hexInt, _ := strconv.ParseInt(hexStr, 16, 64)
		res := int(hexInt) % len(cards)
		rCards = append(rCards, cards[res])
		cards.PopIndex(res)
	}

	return rCards
}

func lolTowerGenerateBombCombinationResult(hash string, sliceIndex SliceIndex) (indx int, result []int) {
	cards := []int{1, 2, 3, 4, 5}

	bombCombinations := [][]int{
		{1, 2},
		{1, 3},
		{1, 4},
		{1, 5},
		{2, 3},
		{2, 4},
		{2, 5},
		{3, 4},
		{3, 5},
		{4, 5},
	}
	hexStr := hash[sliceIndex.PrevIndex:sliceIndex.PostIndex]
	hexInt, _ := strconv.ParseInt(hexStr, 16, 64)
	index := int(hexInt) % len(bombCombinations)
	bombResults := bombCombinations[index]
	results := []int{}

	for i := 0; i < len(cards); i++ {
		if !types.Array[int](bombResults).Constains(cards[i]) {
			results = append(results, cards[i])
		}

	}
	return index, results
}

func lolTowerGenerateResultCombinationResult(hash string, sliceIndex SliceIndex) (indx int, result []int) {
	resultCombinations := [][]int{
		{1, 2, 3},
		{1, 2, 4},
		{1, 2, 5},
		{1, 3, 4},
		{1, 3, 5},
		{1, 4, 5},
		{2, 3, 4},
		{2, 3, 5},
		{2, 4, 5},
		{3, 4, 5},
	}
	hexStr := hash[sliceIndex.PrevIndex:sliceIndex.PostIndex]
	hexInt, _ := strconv.ParseInt(hexStr, 16, 64)
	index := int(hexInt) % len(resultCombinations)

	return index, resultCombinations[index]
}

func lolTowerResultsToString(values []int) string {
	return types.Array[int](values).Join(",")
}

func lolTowerResultsToRowValues(values []int) []string {
	results := []int{1, 2, 3, 4, 5}
	rValues := []string{}

	for _, result := range results {
		if types.Array[int](values).Constains(result) {
			rValues = append(rValues, "x")
		} else {
			rValues = append(rValues, "")
		}
	}
	return rValues
}
