package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"log"
	"os"
	"strconv"
)

const (
	OPEN_LABEL             = "Выбрать"
	SELECT_FOLDER_LABEL    = "Выбраная Папка:"
	PLACEHOLDER_LABEL      = "Placeholder"
	WINDOW_TITLE           = "Расчет Данных ИУЛ"
	SELECT_TEMPLATE_LABEL  = "Выбрать Фаил Шаблона:"
	SELECT_TEMPLATE_BUTTON = "Выбрать"
	SELECT_OUTPUT_LABEL    = "Выбрать Фаил Назначения"
	SELECT_OUTPUT_BUTTON   = "Выбрать"
	RENDER_TEMPLATE_LABEL  = "Заполнить Шаблон:"
	RENDER_TEMPLATE_BUTTON = "Выполнить"
	WINDOW_WIDTH           = 1400
	WINDOW_HEIGHT          = 800
	FILENAME_COLUMN_WIDTH  = 300
	CHECKSUM_COLUMN_WIDTH  = 200
	SIZE_COLUMN_WIDTH      = 150
	CREATED_COLUMN_WIDTH   = 250
)

var HEADERS = []string{"Имя Файла", "Контрольная Сумма", "Размер", "Дата Создания"}

type CheckedFile struct {
	FileName  string
	Checksum  string
	FileSize  string
	CreatedAt string
}

type RenderData struct {
	Items []CheckedFile
}

func NewFolderSelector(window fyne.Window, callback func(uri fyne.ListableURI, err error)) *fyne.Container {
	label := widget.NewLabel(SELECT_FOLDER_LABEL)
	selectedFolderLabel := widget.NewLabel("")
	button := widget.NewButton(OPEN_LABEL, func() {
		dialog.ShowFolderOpen(func(uri fyne.ListableURI, err error) {
			callback(uri, err)
			dir := uri.Path()
			selectedFolderLabel.SetText(dir)
		}, window)
	})
	return container.NewVBox(label, selectedFolderLabel, button)
}

func NewRenderDocumentGroup(callback func()) *fyne.Container {
	renderDocumentLabel := widget.NewLabel(RENDER_TEMPLATE_LABEL)
	renderDocumentButton := widget.NewButton(RENDER_TEMPLATE_BUTTON, callback)
	return container.NewVBox(renderDocumentLabel, renderDocumentButton)
}

func CreateFileDataTable(fileData *[][]string) *widget.Table {
	table := widget.NewTableWithHeaders(
		func() (rows int, cols int) {
			rowsCount := len(*fileData)
			if rowsCount == 0 {
				return 0, 0
			}
			colsCount := len((*fileData)[0])
			return rowsCount, colsCount
		},
		func() fyne.CanvasObject {
			label := widget.NewLabel(PLACEHOLDER_LABEL)
			label.MinSize()
			return label
		},
		func(id widget.TableCellID, object fyne.CanvasObject) {
			cellContent := (*fileData)[id.Row][id.Col]
			label := object.(*widget.Label)
			label.SetText(cellContent)
		},
	)
	table.SetColumnWidth(0, FILENAME_COLUMN_WIDTH)
	table.SetColumnWidth(1, CHECKSUM_COLUMN_WIDTH)
	table.SetColumnWidth(2, SIZE_COLUMN_WIDTH)
	table.SetColumnWidth(3, CREATED_COLUMN_WIDTH)
	table.UpdateHeader = func(id widget.TableCellID, template fyne.CanvasObject) {
		l := template.(*widget.Label)
		if id.Row < 0 {
			l.SetText(HEADERS[id.Col])
		} else if id.Col < 0 {
			l.SetText(strconv.Itoa(id.Row + 1))
		} else {
			l.SetText("")
		}
	}
	return table
}

func updateTable(uri fyne.ListableURI, fileTable *widget.Table, fileData *[][]string) {
	dir := uri.Path()
	files, err := os.ReadDir(dir)
	if err != nil {
		log.Fatal(err)
	}
	var newFileData [][]string
	for _, file := range files {
		checksum, fileSize, createdAt, err := calculateChecksum(file.Name(), dir)
		if err != nil {
			log.Fatal(err)
		}
		newFileData = append(newFileData, []string{file.Name(), checksum, fileSize, createdAt})
	}
	*fileData = newFileData
	fileTable.Refresh()
}

func main() {
	mainApp := app.New()
	window := mainApp.NewWindow(WINDOW_TITLE)
	window.Resize(fyne.NewSize(WINDOW_WIDTH, WINDOW_HEIGHT))
	var fileData [][]string
	var templateFile = "./template.docx"
	var outputFile = "./result.docx"
	renderDocumentBlock := NewRenderDocumentGroup(func() { renderTemplate(fileData, &templateFile, &outputFile) })
	fileTable := CreateFileDataTable(&fileData)
	folderSelector := NewFolderSelector(window, func(uri fyne.ListableURI, err error) { updateTable(uri, fileTable, &fileData) })

	selectedTemplatePath := widget.NewLabel(templateFile)
	selectedOutputPath := widget.NewLabel(outputFile)
	configGroup := container.NewVBox(
		widget.NewLabel(SELECT_TEMPLATE_LABEL),
		selectedTemplatePath,
		widget.NewButton(SELECT_TEMPLATE_BUTTON, func() {
			dialog.ShowFileOpen(func(closer fyne.URIReadCloser, err error) {
				templateFile = closer.URI().Path()
				selectedTemplatePath.SetText(templateFile)
				err = closer.Close()
				if err != nil {
					log.Fatal(err)
				}
			}, window)
		}),
		widget.NewLabel(SELECT_OUTPUT_LABEL),
		selectedOutputPath,
		widget.NewButton(SELECT_OUTPUT_BUTTON, func() {
			dialog.ShowFileSave(func(closer fyne.URIWriteCloser, err error) {
				outputFile = closer.URI().Path()
				selectedOutputPath.SetText(outputFile)
				err = closer.Close()
				if err != nil {
					return
				}
			}, window)
		}),
	)

	controlGroup := container.NewVBox(folderSelector, configGroup, renderDocumentBlock)

	content := container.NewBorder(
		nil,
		nil,
		controlGroup,
		nil,
		fileTable,
	)
	window.SetContent(content)
	window.ShowAndRun()
}
