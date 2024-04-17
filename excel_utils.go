package main

import (
	"fmt"
	"github.com/xuri/excelize/v2"
	"log"
	"strings"
)

const (
	CONTROL_SHEET_NAME = "Лист управления"
	AUTHOR_SHEET_NAME  = "Содержание"
)

type CellRange struct {
	startRow int
	startCol int
	endRow   int
	endCol   int
}

func extractControlData(file *excelize.File) [][]string {
	rows, err := file.GetRows(CONTROL_SHEET_NAME)
	if err != nil {
		log.Printf("Failed to get rows from %s due to %s", CONTROL_SHEET_NAME, err)
		return nil
	}
	return rows
}

func extractAuthorData(file *excelize.File) [][2]string {
	rows, err := file.GetRows(AUTHOR_SHEET_NAME)
	if err != nil {
		log.Printf("Failde to get rows from %s due to %s", AUTHOR_SHEET_NAME, err)
	}
	var cellRangesStr = "D5:E23,F5:G23"
	var cellRanges = strings.Split(cellRangesStr, ",")
	var cellRangesFormatted []CellRange
	for _, cellRange := range cellRanges {
		cells := strings.Split(cellRange, ":")
		startCol, startRow, err := excelize.CellNameToCoordinates(cells[0])
		if err != nil {
			log.Printf("Incorrect cell reange for Authors")
		}
		endCol, endRow, err := excelize.CellNameToCoordinates(cells[1])
		if err != nil {
			log.Printf("Incorrect cell reange for Authors")
		}
		cellRangesFormatted = append(
			cellRangesFormatted,
			CellRange{startCol: startCol - 1, startRow: startRow - 1, endCol: endCol - 1, endRow: endRow - 1})
	}

	var data [][2]string
	for _, cellRangeFormatted := range cellRangesFormatted {
		for _, row := range rows[cellRangeFormatted.startRow:cellRangeFormatted.endRow] {
			data = append(data, [2]string(row[cellRangeFormatted.startCol:cellRangeFormatted.endCol+1]))
		}
	}
	return data
}

func ExtractExcelFileData(path string) ([][]string, [][2]string) {
	f, err := excelize.OpenFile(path)
	if err != nil {
		fmt.Println(err)
		return nil, nil
	}

	return extractControlData(f), extractAuthorData(f)
}
