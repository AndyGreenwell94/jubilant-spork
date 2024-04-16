package main

import (
	"fmt"
	"github.com/xuri/excelize/v2"
)

func convertToAlphabetic(n int) string {
	result := ""
	for n > 0 {
		mod := (n - 1) % 26
		result = string('A'+mod) + result
		n = (n - mod) / 26
	}
	return result
}

func OpenExcelFile(path string) map[string]string {
	f, err := excelize.OpenFile(path)
	if err != nil {
		fmt.Println(err)
		return map[string]string{}
	}
	// Get all the sheet names from the Excel file
	rows, err := f.GetRows("Лист управления")
	if err != nil {
		return map[string]string{}
	}
	data := map[string]string{}
	for rowNumber, row := range rows {
		for collNumber, cell := range row {
			data[fmt.Sprintf("%s%d", convertToAlphabetic(collNumber+1), rowNumber+1)] = cell
		}
	}
	return data
}
