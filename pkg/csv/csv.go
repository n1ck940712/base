package csv

import (
	ecsv "encoding/csv"
	"errors"
	"os"
)

type CSV interface {
	AddRowValues(values ...string)
	InsertRowValues(index int, values ...string)
	Create() error
}

type csv struct {
	filename string
	rows     [][]string
}

func NewCSV(filename string) CSV {
	return &csv{filename: filename}
}

func (c *csv) AddRowValues(values ...string) {
	c.rows = append(c.rows, values)
}

func (c *csv) InsertRowValues(index int, values ...string) {
	inserted := false
	rows := [][]string{}

	for i := 0; i < len(c.rows); i++ {
		if i == index {
			rows = append(rows, values)
			inserted = true
		}
		rows = append(rows, c.rows[i])
	}
	if !inserted { //if not inserted add instead
		c.AddRowValues(values...)
	}
	c.rows = rows
}

func (c *csv) Create() error {
	if c.filename == "" {
		return errors.New("filename cannot be empty")
	}
	csvFile, err := os.Create(c.filename + ".csv")

	if err != nil {
		return errors.New("failed creating csv error: " + err.Error())
	}
	defer csvFile.Close()
	csvWriter := ecsv.NewWriter(csvFile)

	defer csvWriter.Flush()
	if err := csvWriter.WriteAll(c.rows); err != nil {
		return errors.New("failed writing rows csv error: " + err.Error())
	}
	return nil
}
