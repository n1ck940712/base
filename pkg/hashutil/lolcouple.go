package hashutil

import (
	"encoding/json"
	"fmt"
)

func LOLCoupleGenerateResult(hash Hash) (string, error) {
	results := []string{}

	for i := 0; i < 3; i++ {
		femaleI := 2 * i
		femaleIndex := femaleI * 8
		femaleHash := hash.Extract(femaleIndex, 8)
		femaleValue := int8(femaleHash.Int64()%6) + 1

		if err := femaleHash.Error(); err != nil {
			return "", err
		}
		maleI := femaleI + 1
		maleIndex := maleI * 8
		maleHash := hash.Extract(maleIndex, 8)
		maleValue := int8(maleHash.Int64()%6) + 1

		if err := maleHash.Error(); err != nil {
			return "", err
		}
		coupleValue := femaleValue + maleValue

		results = append(results, fmt.Sprintf("%v:%v", maleValue, femaleValue)) //male:female
		if coupleValue != 7 {
			break
		}
	}
	result, err := json.Marshal(results)

	if err != nil {
		return "", err
	}
	return string(result), nil
}
