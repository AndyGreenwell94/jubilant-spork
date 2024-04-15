package main

import (
	"fmt"
	"github.com/xuri/excelize/v2"
)

func OpenExcelFile(path string) {
	f, err := excelize.OpenFile(path)
	if err != nil {
		fmt.Println(err)
		return
	}
	// Get all the sheet names from the Excel file
	rows, err := f.GetRows("Лист управления")
	if err != nil {
		return
	}
	for _, row := range rows {
		for i, col := range row {
			fmt.Println(i, col)
		}
		fmt.Println("----------")
	}
}
