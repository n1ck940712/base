package csv

import (
	"testing"

	"bitbucket.org/esportsph/minigame-backend-golang/pkg/types"
)

func TestCSV(t *testing.T) {
	csvFile := NewCSV("test.csv")

	csvFile.AddRowValues("column 1", "column 2", "column 3")
	for i := 0; i < 3000; i += 3 {
		csvFile.AddRowValues(*types.Int(i).String().Ptr(), *types.Int(i + 1).String().Ptr(), *types.Int(i + 2).String().Ptr())
	}
	if err := csvFile.Create(); err != nil {
		t.Fatal(err.Error())
	}
}
