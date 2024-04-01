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
	"path/filepath"
	"strconv"
)

const (
	OPEN_LABEL          = "Open"
	SELECT_FOLDER_LABEL = "Select Folder:"
	PLACEHOLDER_LABEL   = "placeholder"
	WINDOW_TITLE        = "Check Sum"
	WINDOW_WIDTH        = 1200
	WINDOW_HEIGHT       = 800
	GRID_COLUMNS        = 2
)

func NewFolderSelector(window fyne.Window, callback func(uri fyne.ListableURI, err error)) *fyne.Container {
	folderSelectLabel := widget.NewLabel(SELECT_FOLDER_LABEL)
	folderSelectButton := widget.NewButton(OPEN_LABEL, func() {
		dialog.ShowFolderOpen(callback, window)
	})
	return container.NewVBox(folderSelectLabel, folderSelectButton)
}

func CreateFileDataTable(fileData *[][]string) *widget.Table {
	return widget.NewTableWithHeaders(
		func() (rows int, cols int) {
			rowsCount := len(*fileData)
			if rowsCount == 0 {
				return 0, 0
			}
			colsCount := len((*fileData)[0])
			return rowsCount, colsCount
		},
		func() fyne.CanvasObject {
			return widget.NewLabel(PLACEHOLDER_LABEL)
		},
		func(id widget.TableCellID, object fyne.CanvasObject) {
			cellContent := (*fileData)[id.Row][id.Col]
			label := object.(*widget.Label)
			label.SetText(cellContent)
		},
	)
}

func calculateChecksum(fileName string, dir string) (string, string, error) {
	filePath := filepath.Join(dir, fileName)
	file, err := os.Open(filePath)
	if err != nil {
		return "", "", err
	}

	defer func() {
		closeErr := file.Close()
		if closeErr != nil && err == nil { // Update the err return value if it's nil.
			err = closeErr
		}
	}()

	hasher := crc32.NewIEEE()
	if _, err = io.Copy(hasher, file); err != nil {
		return "", "", err
	}
	checksum := hasher.Sum32()
	fileInfo, err := file.Stat()
	if err != nil {
		return "", "", err
	}

	return fmt.Sprintf("%x", checksum), strconv.FormatInt(fileInfo.Size(), 10), err
}

func updateTable(uri fyne.ListableURI, fileTable *widget.Table, fileData *[][]string) {
	dir := uri.Path()
	files, err := os.ReadDir(dir)
	if err != nil {
		log.Fatal(err)
	}
	var newFileData [][]string
	for _, file := range files {
		checksum, fileSize, err := calculateChecksum(file.Name(), dir)
		if err != nil {
			log.Fatal(err)
		}
		newFileData = append(newFileData, []string{file.Name(), checksum, fileSize})
	}
	*fileData = newFileData
	fileTable.Refresh()
}

func main() {
	mainApp := app.New()
	window := mainApp.NewWindow(WINDOW_TITLE)
	window.Resize(fyne.NewSize(WINDOW_WIDTH, WINDOW_HEIGHT))
	var fileData [][]string
	fileTable := CreateFileDataTable(&fileData)
	window.SetContent(
		container.NewGridWithColumns(
			GRID_COLUMNS,
			NewFolderSelector(window, func(uri fyne.ListableURI, err error) { updateTable(uri, fileTable, &fileData) }),
			fileTable,
		),
	)
	window.ShowAndRun()
}
