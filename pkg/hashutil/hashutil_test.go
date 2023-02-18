package hashutil

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"bitbucket.org/esportsph/minigame-backend-golang/pkg/csv"
	"bitbucket.org/esportsph/minigame-backend-golang/pkg/types"
)

func TestLOLTowerResults(t *testing.T) {
	hash := "b1987538713ae6e46f825c70f2fdc06bef5b8659d25daf2a4bbe07d8d1eb7cda"
	result := lolTowerGenerateHashResult(hash)

	println("win: ", lolTowerResultsToString(result))
}

func TestFIFAShootupResults(t *testing.T) {
	tHashes := []struct {
		hash     Hash
		expected string
	}{
		{hash: NewHash("b613679a0814d9ec772f95d778c35fc5ff1697c493715653c6c712144292c5ad"), expected: "3-hearts,8-spades,8-hearts"},   //v1:8-hearts,9-spades,10-diamonds, v2:3-hearts,8-spades,8-hearts
		{hash: NewHash("e9bf61c0e9bf61c0e9bf61c011cbe805aae14fb49b148940593c7d5b46a1b283"), expected: "5-diamonds,5-spades,k-hearts"}, //v1:k-hearts,k-spades,k-clubs, v2:5-diamonds,5-spades,k-hearts
		{hash: NewHash("e9bf61c0e9bf61c0e9bf61c011cbe80511cbe80511cbe80511cbe80511cbe805"), expected: "5-spades,5-hearts,k-spades"},   //v1:k-hearts,k-clubs,k-diamonds, v2:5-spades,5-hearts,k-spades
		{hash: NewHash("e5d23126b6c34e0e6554030bab389c02d39ac636f36b918b83500a8dd184f336"), expected: "j-hearts,k-clubs,2-clubs"},     //v1:2-clubs,k-clubs,2-diamonds, v2:j-hearts,k-clubs,2-clubs
	}

	for i, tHash := range tHashes {
		if selections, result, err := FIFAShootupGenerateResult(tHash.hash); err != nil {
			t.Fatal(err)
		} else if selections != tHash.expected {
			t.Fatal(fmt.Errorf("case %v result selection (%v) expected hash (%v)", i+1, selections, tHash.expected))
		} else {
			println("selections: ", selections, " result: ", result)
		}
	}
}

func TestLOLTowerCompareResults(t *testing.T) {
	for i := 0; i < 1_000_000; i++ {
		generatedHash := GenerateHash()
		hash := NewHash(generatedHash)

		hashResult := lolTowerGenerateHashResult(generatedHash)
		hashResultStr := lolTowerResultsToString(hashResult)
		newHashResult, err := LOLTowerGenerateResult(hash)

		if err != nil {
			t.Fatal(err.Error())
		}
		if hashResultStr != newHashResult {
			t.Fatal("hashResultStr != newHashResult (" + hashResultStr + " != " + newHashResult + ")" + " at index " + *types.Int(i).String().Ptr())
		}
	}
	println("success old implementation is same with new implementation")
}

func TestCSVLOLTowerResults(t *testing.T) {
	hashCount := 100_000
	selections := []int{1, 2, 3, 4, 5}
	csvLOLTower := csv.NewCSV("loltower-result")
	csvLOLTowerVis := csv.NewCSV("loltower-result-visualize")

	csvLOLTower.AddRowValues("hash",
		"first16 from bomb",
		"last16 from bomb",
		"first8_last8 from bomb",
		"first8_second8_third8 from win",
		"bomb combination",
		"win combination",
		"%2==1?bomb:win",
	)
	csvLOLTowerVis.AddRowValues("Legend:", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "")
	csvLOLTowerVis.AddRowValues("g0", "first16 from bomb", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "")
	csvLOLTowerVis.AddRowValues("g1", "last16 from bomb", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "")
	csvLOLTowerVis.AddRowValues("g2", "first8_last8 from bomb", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "")
	csvLOLTowerVis.AddRowValues("g3", "first8_second8_third8 from win", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "")
	csvLOLTowerVis.AddRowValues("g4", "bomb combination", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "")
	csvLOLTowerVis.AddRowValues("g5", "result combination", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "")
	csvLOLTowerVis.AddRowValues("g6", "%2==1?bomb:win", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "")
	csvLOLTowerVis.AddRowValues("loss_value", "from bomb generated result", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "")
	csvLOLTowerVis.AddRowValues("win_value", "from win generated result", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "")
	csvLOLTowerVis.AddRowValues("", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "", "")
	csvLOLTowerVis.AddRowValues("hash",
		"g0_loss_value_1", "g0_loss_value_2", "g0_loss_value_3", "g0_loss_value_4", "g0_loss_value_5", "",
		"g1_loss_value_1", "g1_loss_value_2", "g1_loss_value_3", "g1_loss_value_4", "g1_loss_value_5", "",
		"g2_loss_value_1", "g2_loss_value_2", "g2_loss_value_3", "g2_loss_value_4", "g2_loss_value_5", "",
		"g3_win_value_1", "g3_win_value_2", "g3_win_value_3", "g3_win_value_4", "g3_win_value_5", "",
		"g4_loss_value_1", "g4_loss_value_2", "g4_loss_value_3", "g4_loss_value_4", "g4_loss_value_5", "",
		"g5_win_value_1", "g5_win_value_2", "g5_win_value_3", "g5_win_value_4", "g5_win_value_5", "",
		"g6_loss_win_value_1", "g6_loss_win_value_2", "g6_loss_win_value_3", "g6_loss_win_value_4", "g6_loss_win_value_5", "",
	)

	hashes := []string{}

	for i := 0; i < hashCount; i++ {
		hashes = append(hashes, GenerateHash())
	}
	first16Selections := map[int]map[int]int{
		1: {},
		2: {},
		3: {},
		4: {},
		5: {},
	} //map[<selection>]map[<repeat>]<count>
	first16Streak := map[int]int{
		1: 0,
		2: 0,
		3: 0,
		4: 0,
		5: 0,
	} //map[<selection>]<streak_count>
	first16Appearance := map[int]int{
		1: 0,
		2: 0,
		3: 0,
		4: 0,
		5: 0,
	}
	last16Selections := map[int]map[int]int{
		1: {},
		2: {},
		3: {},
		4: {},
		5: {},
	} //map[<selection>]map[<repeat>]<count>
	last16Streak := map[int]int{
		1: 0,
		2: 0,
		3: 0,
		4: 0,
		5: 0,
	} //map[<selection>]<streak_count>
	last16Appearance := map[int]int{
		1: 0,
		2: 0,
		3: 0,
		4: 0,
		5: 0,
	} //map[<selection>]<streak_count>
	first8Last8Selections := map[int]map[int]int{
		1: {},
		2: {},
		3: {},
		4: {},
		5: {},
	} //map[<selection>]map[<repeat>]<count>
	first8Last8Streak := map[int]int{
		1: 0,
		2: 0,
		3: 0,
		4: 0,
		5: 0,
	} //map[<selection>]<streak_count>
	first8Last8Appearance := map[int]int{
		1: 0,
		2: 0,
		3: 0,
		4: 0,
		5: 0,
	} //map[<selection>]<streak_count>
	first8Second8Third8Selections := map[int]map[int]int{
		1: {},
		2: {},
		3: {},
		4: {},
		5: {},
	} //map[<selection>]map[<repeat>]<count>
	first8Second8Third8Streak := map[int]int{
		1: 0,
		2: 0,
		3: 0,
		4: 0,
		5: 0,
	} //map[<selection>]<streak_count>
	first8Second8Third8Appearance := map[int]int{
		1: 0,
		2: 0,
		3: 0,
		4: 0,
		5: 0,
	} //map[<selection>]<streak_count>
	bombComboSelections := map[int]map[int]int{
		1: {},
		2: {},
		3: {},
		4: {},
		5: {},
	} //map[<selection>]map[<repeat>]<count>
	bombComboStreak := map[int]int{
		1: 0,
		2: 0,
		3: 0,
		4: 0,
		5: 0,
	} //map[<selection>]<streak_count>
	bombComboAppearance := map[int]int{
		1: 0,
		2: 0,
		3: 0,
		4: 0,
		5: 0,
	} //map[<selection>]<streak_count>
	resultComboSelections := map[int]map[int]int{
		1: {},
		2: {},
		3: {},
		4: {},
		5: {},
	} //map[<selection>]map[<repeat>]<count>
	resultComboStreak := map[int]int{
		1: 0,
		2: 0,
		3: 0,
		4: 0,
		5: 0,
	} //map[<selection>]<streak_count>
	resultComboAppearance := map[int]int{
		1: 0,
		2: 0,
		3: 0,
		4: 0,
		5: 0,
	} //map[<selection>]<streak_count>
	finalImplSelections := map[int]map[int]int{
		1: {},
		2: {},
		3: {},
		4: {},
		5: {},
	} //map[<selection>]map[<repeat>]<count>
	finalImplStreak := map[int]int{
		1: 0,
		2: 0,
		3: 0,
		4: 0,
		5: 0,
	} //map[<selection>]<streak_count>
	finalImplAppearance := map[int]int{
		1: 0,
		2: 0,
		3: 0,
		4: 0,
		5: 0,
	}

	for i := 0; i < len(hashes); i++ {
		resultFirst16 := lolTowerGenerateHashResultFromBomb(hashes[i], SliceIndex{0, 8}, SliceIndex{8, 16})
		resultLast16 := lolTowerGenerateHashResultFromBomb(hashes[i], SliceIndex{48, 56}, SliceIndex{56, 64})
		resultFirst8Last8 := lolTowerGenerateHashResultFromBomb(hashes[i], SliceIndex{0, 8}, SliceIndex{56, 64})
		resultFirst8Second8Third8 := lolTowerGenerateHashResultFromWin(hashes[i], SliceIndex{0, 8}, SliceIndex{8, 16}, SliceIndex{16, 24})
		bcIndex, resultBombCombo := lolTowerGenerateBombCombinationResult(hashes[i], SliceIndex{0, 8})
		rcIndex, resultResultCombo := lolTowerGenerateResultCombinationResult(hashes[i], SliceIndex{0, 8})
		resultFinalImpl := lolTowerGenerateHashResult(hashes[i])

		for _, selection := range selections {
			if types.Array[int](resultFirst16).Constains(selection) {
				first16Appearance[selection] += 1
			}
			if types.Array[int](resultLast16).Constains(selection) {
				last16Appearance[selection] += 1
			}
			if types.Array[int](resultFirst8Last8).Constains(selection) {
				first8Last8Appearance[selection] += 1
			}
			if types.Array[int](resultFirst8Second8Third8).Constains(selection) {
				first8Second8Third8Appearance[selection] += 1
			}
			if types.Array[int](resultBombCombo).Constains(selection) {
				bombComboAppearance[selection] += 1
			}
			if types.Array[int](resultResultCombo).Constains(selection) {
				resultComboAppearance[selection] += 1
			}
			if types.Array[int](resultFinalImpl).Constains(selection) {
				finalImplAppearance[selection] += 1
			}
		}

		if i > 0 {
			prevResultFirst16 := lolTowerGenerateHashResultFromBomb(hashes[i-1], SliceIndex{0, 8}, SliceIndex{8, 16})
			prevResultLast16 := lolTowerGenerateHashResultFromBomb(hashes[i-1], SliceIndex{48, 56}, SliceIndex{56, 64})
			prevResultFirst8Last8 := lolTowerGenerateHashResultFromBomb(hashes[i-1], SliceIndex{0, 8}, SliceIndex{56, 64})
			prevResultFirst8Second8Third8 := lolTowerGenerateHashResultFromWin(hashes[i-1], SliceIndex{0, 8}, SliceIndex{8, 16}, SliceIndex{16, 24})
			_, prevResultBombCombo := lolTowerGenerateBombCombinationResult(hashes[i-1], SliceIndex{0, 8})
			_, prevResultResultCombo := lolTowerGenerateResultCombinationResult(hashes[i-1], SliceIndex{0, 8})
			prevResultFinalImpl := lolTowerGenerateHashResult(hashes[i-1])

			for _, selection := range selections {
				//first16
				if types.Array[int](prevResultFirst16).Constains(selection) && types.Array[int](resultFirst16).Constains(selection) {
					first16Streak[selection] += 1
				} else {
					if first16Streak[selection] > 0 {
						if repeat, ok := first16Selections[selection][first16Streak[selection]]; ok {
							first16Selections[selection][first16Streak[selection]] = repeat + 1
						} else {
							first16Selections[selection][first16Streak[selection]] = 1
						}
					}
					first16Streak[selection] = 0
				}
				if (len(hashes)-1) == i && first16Streak[selection] > 0 { //save on last index
					if repeat, ok := first16Selections[selection][first16Streak[selection]]; ok {
						first16Selections[selection][first16Streak[selection]] = repeat + 1
					} else {
						first16Selections[selection][first16Streak[selection]] = 1
					}
				}

				//last16
				if types.Array[int](prevResultLast16).Constains(selection) && types.Array[int](resultLast16).Constains(selection) {
					last16Streak[selection] += 1
				} else {
					if last16Streak[selection] > 0 {
						if repeat, ok := last16Selections[selection][last16Streak[selection]]; ok {
							last16Selections[selection][last16Streak[selection]] = repeat + 1
						} else {
							last16Selections[selection][last16Streak[selection]] = 1
						}
					}
					last16Streak[selection] = 0
				}
				if (len(hashes)-1) == i && last16Streak[selection] > 0 { //save on last index
					if repeat, ok := last16Selections[selection][last16Streak[selection]]; ok {
						last16Selections[selection][last16Streak[selection]] = repeat + 1
					} else {
						last16Selections[selection][last16Streak[selection]] = 1
					}
				}

				//first8Last8
				if types.Array[int](prevResultFirst8Last8).Constains(selection) && types.Array[int](resultFirst8Last8).Constains(selection) {
					first8Last8Streak[selection] += 1
				} else {
					if first8Last8Streak[selection] > 0 {
						if repeat, ok := first8Last8Selections[selection][first8Last8Streak[selection]]; ok {
							first8Last8Selections[selection][first8Last8Streak[selection]] = repeat + 1
						} else {
							first8Last8Selections[selection][first8Last8Streak[selection]] = 1
						}
					}
					first8Last8Streak[selection] = 0
				}
				if (len(hashes)-1) == i && first8Last8Streak[selection] > 0 { //save on last index
					if repeat, ok := first8Last8Selections[selection][first8Last8Streak[selection]]; ok {
						first8Last8Selections[selection][first8Last8Streak[selection]] = repeat + 1
					} else {
						first8Last8Selections[selection][first8Last8Streak[selection]] = 1
					}
				}

				//first8Second8Third8
				if types.Array[int](prevResultFirst8Second8Third8).Constains(selection) && types.Array[int](resultFirst8Second8Third8).Constains(selection) {
					first8Second8Third8Streak[selection] += 1
				} else {
					if first8Second8Third8Streak[selection] > 0 {
						if repeat, ok := first8Second8Third8Selections[selection][first8Second8Third8Streak[selection]]; ok {
							first8Second8Third8Selections[selection][first8Second8Third8Streak[selection]] = repeat + 1
						} else {
							first8Second8Third8Selections[selection][first8Second8Third8Streak[selection]] = 1
						}
					}
					first8Second8Third8Streak[selection] = 0
				}
				if (len(hashes)-1) == i && first8Second8Third8Streak[selection] > 0 { //save on last index
					if repeat, ok := first8Second8Third8Selections[selection][first8Second8Third8Streak[selection]]; ok {
						first8Second8Third8Selections[selection][first8Second8Third8Streak[selection]] = repeat + 1
					} else {
						first8Second8Third8Selections[selection][first8Second8Third8Streak[selection]] = 1
					}
				}

				//bombCombo
				if types.Array[int](prevResultBombCombo).Constains(selection) && types.Array[int](resultBombCombo).Constains(selection) {
					bombComboStreak[selection] += 1
				} else {
					if bombComboStreak[selection] > 0 {
						if repeat, ok := bombComboSelections[selection][bombComboStreak[selection]]; ok {
							bombComboSelections[selection][bombComboStreak[selection]] = repeat + 1
						} else {
							bombComboSelections[selection][bombComboStreak[selection]] = 1
						}
					}
					bombComboStreak[selection] = 0
				}
				if (len(hashes)-1) == i && bombComboStreak[selection] > 0 { //save on last index
					if repeat, ok := bombComboSelections[selection][bombComboStreak[selection]]; ok {
						bombComboSelections[selection][bombComboStreak[selection]] = repeat + 1
					} else {
						bombComboSelections[selection][bombComboStreak[selection]] = 1
					}
				}

				//resultCombo
				if types.Array[int](prevResultResultCombo).Constains(selection) && types.Array[int](resultResultCombo).Constains(selection) {
					resultComboStreak[selection] += 1
				} else {
					if resultComboStreak[selection] > 0 {
						if repeat, ok := resultComboSelections[selection][resultComboStreak[selection]]; ok {
							resultComboSelections[selection][resultComboStreak[selection]] = repeat + 1
						} else {
							resultComboSelections[selection][resultComboStreak[selection]] = 1
						}
					}
					resultComboStreak[selection] = 0
				}
				if (len(hashes)-1) == i && resultComboStreak[selection] > 0 { //save on last index
					if repeat, ok := resultComboSelections[selection][resultComboStreak[selection]]; ok {
						resultComboSelections[selection][resultComboStreak[selection]] = repeat + 1
					} else {
						resultComboSelections[selection][resultComboStreak[selection]] = 1
					}
				}

				//finalImpl
				if types.Array[int](prevResultFinalImpl).Constains(selection) && types.Array[int](resultFinalImpl).Constains(selection) {
					finalImplStreak[selection] += 1
				} else {
					if finalImplStreak[selection] > 0 {
						if repeat, ok := finalImplSelections[selection][finalImplStreak[selection]]; ok {
							finalImplSelections[selection][finalImplStreak[selection]] = repeat + 1
						} else {
							finalImplSelections[selection][finalImplStreak[selection]] = 1
						}
					}
					finalImplStreak[selection] = 0
				}
				if (len(hashes)-1) == i && finalImplStreak[selection] > 0 { //save on last index
					if repeat, ok := finalImplSelections[selection][finalImplStreak[selection]]; ok {
						finalImplSelections[selection][finalImplStreak[selection]] = repeat + 1
					} else {
						finalImplSelections[selection][finalImplStreak[selection]] = 1
					}
				}
			}
		}

		csvLOLTower.AddRowValues(hashes[i],
			lolTowerResultsToString(resultFirst16),
			lolTowerResultsToString(resultLast16),
			lolTowerResultsToString(resultFirst8Last8),
			lolTowerResultsToString(resultFirst8Second8Third8),
			lolTowerResultsToString(resultBombCombo)+" index:"+fmt.Sprint(bcIndex),
			lolTowerResultsToString(resultResultCombo)+" index:"+fmt.Sprint(rcIndex),
			lolTowerResultsToString(resultFinalImpl),
		)

		csvLOLTowerVisRowValues := []string{hashes[i]}
		csvLOLTowerVisRowValues = append(csvLOLTowerVisRowValues, lolTowerResultsToRowValues(resultFirst16)...)
		csvLOLTowerVisRowValues = append(csvLOLTowerVisRowValues, "")
		csvLOLTowerVisRowValues = append(csvLOLTowerVisRowValues, lolTowerResultsToRowValues(resultLast16)...)
		csvLOLTowerVisRowValues = append(csvLOLTowerVisRowValues, "")
		csvLOLTowerVisRowValues = append(csvLOLTowerVisRowValues, lolTowerResultsToRowValues(resultFirst8Last8)...)
		csvLOLTowerVisRowValues = append(csvLOLTowerVisRowValues, "")
		csvLOLTowerVisRowValues = append(csvLOLTowerVisRowValues, lolTowerResultsToRowValues(resultFirst8Second8Third8)...)
		csvLOLTowerVisRowValues = append(csvLOLTowerVisRowValues, "")
		csvLOLTowerVisRowValues = append(csvLOLTowerVisRowValues, lolTowerResultsToRowValues(resultBombCombo)...)
		csvLOLTowerVisRowValues = append(csvLOLTowerVisRowValues, "")
		csvLOLTowerVisRowValues = append(csvLOLTowerVisRowValues, lolTowerResultsToRowValues(resultResultCombo)...)
		csvLOLTowerVisRowValues = append(csvLOLTowerVisRowValues, "")
		csvLOLTowerVisRowValues = append(csvLOLTowerVisRowValues, lolTowerResultsToRowValues(resultFinalImpl)...)
		csvLOLTowerVisRowValues = append(csvLOLTowerVisRowValues, "")
		csvLOLTowerVis.AddRowValues(csvLOLTowerVisRowValues...)
	}

	if err := csvLOLTower.Create(); err != nil {
		t.Fatal(err)
	}

	if err := csvLOLTowerVis.Create(); err != nil {
		t.Fatal(err)
	}

	maxRepeat := 0
	for _, selection := range selections {
		for key := range first16Selections[selection] {
			if key > maxRepeat {
				maxRepeat = key
			}
		}
		for key := range last16Selections[selection] {
			if key > maxRepeat {
				maxRepeat = key
			}
		}
		for key := range first8Last8Selections[selection] {
			if key > maxRepeat {
				maxRepeat = key
			}
		}
		for key := range first8Second8Third8Selections[selection] {
			if key > maxRepeat {
				maxRepeat = key
			}
		}
		for key := range bombComboSelections[selection] {
			if key > maxRepeat {
				maxRepeat = key
			}
		}
		for key := range resultComboSelections[selection] {
			if key > maxRepeat {
				maxRepeat = key
			}
		}
		for key := range finalImplSelections[selection] {
			if key > maxRepeat {
				maxRepeat = key
			}
		}
	}
	first16MoreColumns := []string{"first16 from bomb", `result\repeat`}
	last16MoreColumns := []string{"last16 from bomb", `result\repeat`}
	first8Last8MoreColumns := []string{"first8_last8 from bomb", `result\repeat`}
	first8Second8Third8MoreColumns := []string{"first8_second8_third8 from win", `result\repeat`}
	bombComboMoreColumns := []string{"bomb combination", `result\repeat`}
	resultComboMoreColumns := []string{"result combination", `result\repeat`}
	finalImplMoreColumns := []string{"%2==1?bomb:win", `result\repeat`}

	for i := 1; i <= maxRepeat; i++ {
		column := fmt.Sprint(i)

		first16MoreColumns = append(first16MoreColumns, column)
		last16MoreColumns = append(last16MoreColumns, column)
		first8Last8MoreColumns = append(first8Last8MoreColumns, column)
		first8Second8Third8MoreColumns = append(first8Second8Third8MoreColumns, column)
		bombComboMoreColumns = append(bombComboMoreColumns, column)
		resultComboMoreColumns = append(resultComboMoreColumns, column)
		finalImplMoreColumns = append(finalImplMoreColumns, column)
	}
	first16MoreColumns = append(first16MoreColumns, "appearance")
	last16MoreColumns = append(last16MoreColumns, "appearance")
	first8Last8MoreColumns = append(first8Last8MoreColumns, "appearance")
	first8Second8Third8MoreColumns = append(first8Second8Third8MoreColumns, "appearance")
	bombComboMoreColumns = append(bombComboMoreColumns, "appearance")
	resultComboMoreColumns = append(resultComboMoreColumns, "appearance")
	finalImplMoreColumns = append(finalImplMoreColumns, "appearance")
	csvLOLTowerMore := csv.NewCSV("loltower-result-more")

	csvLOLTowerMore.AddRowValues(first16MoreColumns...)
	for _, selection := range selections {
		rowValues := []string{"", fmt.Sprint(selection)}
		for i := 1; i <= maxRepeat; i++ {
			if repeatValue, ok := first16Selections[selection][i]; ok {
				rowValues = append(rowValues, fmt.Sprint(repeatValue))
			} else {
				rowValues = append(rowValues, "0")
			}
		}
		rowValues = append(rowValues, fmt.Sprint(first16Appearance[selection]))
		csvLOLTowerMore.AddRowValues(rowValues...)
	}
	csvLOLTowerMore.AddRowValues(last16MoreColumns...)
	for _, selection := range selections {
		rowValues := []string{"", fmt.Sprint(selection)}
		for i := 1; i <= maxRepeat; i++ {
			if repeatValue, ok := last16Selections[selection][i]; ok {
				rowValues = append(rowValues, fmt.Sprint(repeatValue))
			} else {
				rowValues = append(rowValues, "0")
			}
		}
		rowValues = append(rowValues, fmt.Sprint(last16Appearance[selection]))
		csvLOLTowerMore.AddRowValues(rowValues...)
	}
	csvLOLTowerMore.AddRowValues(first8Last8MoreColumns...)
	for _, selection := range selections {
		rowValues := []string{"", fmt.Sprint(selection)}
		for i := 1; i <= maxRepeat; i++ {
			if repeatValue, ok := first8Last8Selections[selection][i]; ok {
				rowValues = append(rowValues, fmt.Sprint(repeatValue))
			} else {
				rowValues = append(rowValues, "0")
			}
		}
		rowValues = append(rowValues, fmt.Sprint(first8Last8Appearance[selection]))
		csvLOLTowerMore.AddRowValues(rowValues...)
	}
	csvLOLTowerMore.AddRowValues(first8Second8Third8MoreColumns...)
	for _, selection := range selections {
		rowValues := []string{"", fmt.Sprint(selection)}
		for i := 1; i <= maxRepeat; i++ {
			if repeatValue, ok := first8Second8Third8Selections[selection][i]; ok {
				rowValues = append(rowValues, fmt.Sprint(repeatValue))
			} else {
				rowValues = append(rowValues, "0")
			}
		}
		rowValues = append(rowValues, fmt.Sprint(first8Second8Third8Appearance[selection]))
		csvLOLTowerMore.AddRowValues(rowValues...)
	}
	csvLOLTowerMore.AddRowValues(bombComboMoreColumns...)
	for _, selection := range selections {
		rowValues := []string{"", fmt.Sprint(selection)}
		for i := 1; i <= maxRepeat; i++ {
			if repeatValue, ok := bombComboSelections[selection][i]; ok {
				rowValues = append(rowValues, fmt.Sprint(repeatValue))
			} else {
				rowValues = append(rowValues, "0")
			}
		}
		rowValues = append(rowValues, fmt.Sprint(bombComboAppearance[selection]))
		csvLOLTowerMore.AddRowValues(rowValues...)
	}
	csvLOLTowerMore.AddRowValues(resultComboMoreColumns...)
	for _, selection := range selections {
		rowValues := []string{"", fmt.Sprint(selection)}
		for i := 1; i <= maxRepeat; i++ {
			if repeatValue, ok := resultComboSelections[selection][i]; ok {
				rowValues = append(rowValues, fmt.Sprint(repeatValue))
			} else {
				rowValues = append(rowValues, "0")
			}
		}
		rowValues = append(rowValues, fmt.Sprint(resultComboAppearance[selection]))
		csvLOLTowerMore.AddRowValues(rowValues...)
	}
	csvLOLTowerMore.AddRowValues(finalImplMoreColumns...)
	for _, selection := range selections {
		rowValues := []string{"", fmt.Sprint(selection)}
		for i := 1; i <= maxRepeat; i++ {
			if repeatValue, ok := finalImplSelections[selection][i]; ok {
				rowValues = append(rowValues, fmt.Sprint(repeatValue))
			} else {
				rowValues = append(rowValues, "0")
			}
		}
		rowValues = append(rowValues, fmt.Sprint(finalImplAppearance[selection]))
		csvLOLTowerMore.AddRowValues(rowValues...)
	}

	if err := csvLOLTowerMore.Create(); err != nil {
		t.Fatal(err)
	}
}

func TestCSVLOLCoupleResults(t *testing.T) {
	resultsCount := 100_000
	csvLOLCouple := csv.NewCSV("lolcouple-result")

	csvLOLCouple.AddRowValues("Game Hash", "Round 1", "Round 2", "Round 3")

	for i := 0; i < resultsCount; i++ {
		hash := NewHash(GenerateHash())
		csvRow := []string{hash.Raw()}
		result, rErr := LOLCoupleGenerateResult(hash)

		if rErr != nil {
			t.Fatal(rErr)
		}
		results := []string{}

		if err := json.Unmarshal([]byte(result), &results); err != nil {
			t.Fatal("json Unmarshal error: " + err.Error())
		}
		for i := 0; i < len(results); i++ {
			csvRow = append(csvRow, results[i])
		}
		csvLOLCouple.AddRowValues(csvRow...)

	}
	if err := csvLOLCouple.Create(); err != nil {
		t.Fatal(err)
	}
}

func TestCSVSoccerShootoutResults(t *testing.T) {
	resultsCount := 100_000
	csvSoccerShootout := csv.NewCSV("soccershootout-result")

	csvSoccerShootout.AddRowValues("Game Hash", "Left Card", "Right Card", "Result Card", "Result")

	for i := 0; i < resultsCount; i++ {
		hash := NewHash(GenerateHash())
		csvRow := []string{hash.Raw()}
		selections, result, err := FIFAShootupGenerateResult(hash)

		if err != nil {
			t.Fatal(err)
		}
		selectionsArr := strings.Split(selections, ",")

		csvRow = append(csvRow, selectionsArr[0])
		csvRow = append(csvRow, selectionsArr[1])
		csvRow = append(csvRow, selectionsArr[2])
		csvRow = append(csvRow, result)

		csvSoccerShootout.AddRowValues(csvRow...)

	}
	if err := csvSoccerShootout.Create(); err != nil {
		t.Fatal(err)
	}
}
