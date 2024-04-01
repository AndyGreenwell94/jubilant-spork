package main

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"hash/crc32"
	"io"
	"log"
	"os"
)

func NewFolderSelect(window fyne.Window, callback func(uri fyne.ListableURI, err error)) *fyne.Container {
	folderSelectLabel := widget.NewLabel("Select Folder:")
	folderSelectButton := widget.NewButton("open", func() {
		dialog.ShowFolderOpen(callback, window)
	})
	return container.NewVBox(folderSelectLabel, folderSelectButton)
}

func NewFileTable(data *[][]string) *widget.Table {
	table := widget.NewTableWithHeaders(
		func() (rows int, cols int) {
			rowsCount := len(*data)
			if rowsCount == 0 {
				return 0, 0
			}
			colsCount := len((*data)[0])
			return rowsCount, colsCount
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("placeholder")
		},
		func(id widget.TableCellID, object fyne.CanvasObject) {
			label := object.(*widget.Label)
			label.SetText((*data)[id.Row][id.Col])
		},
	)
	return table
}

func main() {
	mainApp := app.New()
	window := mainApp.NewWindow("Check Sum")
	window.Resize(fyne.NewSize(1200, 800))
	var data [][]string
	table := NewFileTable(&data)
	updateTableOnDataChange := func(uri fyne.ListableURI, err error) {
		dir := uri.Path()
		files, err := os.ReadDir(dir)
		if err != nil {
			log.Fatal(err)
		}

		for _, file := range files {
			filePath := fmt.Sprintf("%s/%s", dir, file.Name())
			f, err := os.Open(filePath)
			if err != nil {
				log.Fatal(err)
			}

			hasher := crc32.NewIEEE()
			if _, err := io.Copy(hasher, f); err != nil {
				log.Fatal(err)
			}
			checksum := hasher.Sum32()
			fileInfo, err := file.Info()
			if err != nil {
				log.Fatal(err)
			}
			fileSize := fmt.Sprintf("%d", fileInfo.Size())

			data = append(data, []string{file.Name(), fmt.Sprintf("%x", checksum), fileSize})
		}
		table.Refresh()
	}
	window.SetContent(
		container.NewGridWithColumns(
			2,
			NewFolderSelect(window, updateTableOnDataChange),
			table,
		),
	)
	window.ShowAndRun()
}
