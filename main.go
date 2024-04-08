package main

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
	"log"
	"os"
	"path"
	"strconv"
)

const (
	OpenLabel                 = "Выбрать"
	SelectFolderLabel         = "Выбраная Папка:"
	PlaceholderLabel          = "Placeholder"
	WindowTitle               = "Расчет Данных ИУЛ"
	SelectTemplateLabel       = "Выбрать Фаил Шаблона:"
	SelectTemplateButton      = "Выбрать"
	SelectOutputLabel         = "Выбрать Фаил Назначения"
	SelectOutputButton        = "Выбрать"
	RenderTemplateLabel       = "Заполнить Шаблон:"
	RenderTemplateButton      = "Выполнить"
	RenderCompleteLabel       = "Документ Сформирован"
	RenderCompleteMsgTemplate = "Документ был успешно сформирован:\n %s"
	DefaultTemplatePath       = "./template.docx"
	DefaultOutputPath         = "./result.docx"
	WindowWidth               = 1400
	WindowHeight              = 800
	FilenameColumnWidth       = 300
	ChecksumColumnWidth       = 200
	SizeColumnWidth           = 150
	CreatedColumnWidth        = 250
)

var TableHeaders = [4]string{"Имя Файла", "Контрольная Сумма", "Размер", "Дата Создания"}

func NewFolderSelectGroup(window fyne.Window, callback func(uri fyne.ListableURI, err error)) *fyne.Container {
	label := widget.NewLabel(SelectFolderLabel)
	selectedFolderLabel := widget.NewLabel("")
	button := widget.NewButton(OpenLabel, func() {
		dialog.ShowFolderOpen(func(uri fyne.ListableURI, err error) {
			callback(uri, err)
			dir := uri.Path()
			selectedFolderLabel.SetText(dir)
		}, window)
	})
	return container.NewVBox(label, selectedFolderLabel, button)
}

func NewConfigGroup(window fyne.Window, templateFile *string, outputFile *string) *fyne.Container {
	selectedTemplatePath := widget.NewLabel(*templateFile)
	selectedOutputPath := widget.NewLabel(*outputFile)
	return container.NewVBox(
		widget.NewLabel(SelectTemplateLabel),
		selectedTemplatePath,
		widget.NewButton(SelectTemplateButton, func() {
			templateOpenDialog := dialog.NewFileOpen(func(closer fyne.URIReadCloser, err error) {
				*templateFile = closer.URI().Path()
				selectedTemplatePath.SetText(*templateFile)
				err = closer.Close()
				if err != nil {
					log.Fatal(err)
				}
			}, window)
			templateOpenDialog.SetFilter(storage.NewExtensionFileFilter([]string{".doc", ".docx"}))
			templateOpenDialog.Show()
		}),
		widget.NewLabel(SelectOutputLabel),
		selectedOutputPath,
		widget.NewButton(SelectOutputButton, func() {
			fileSaveDialog := dialog.NewFileSave(func(closer fyne.URIWriteCloser, err error) {
				*outputFile = closer.URI().Path()
				selectedOutputPath.SetText(*outputFile)
				err = closer.Close()
				if err != nil {
					return
				}
			}, window)
			fileSaveDialog.SetFilter(storage.NewExtensionFileFilter([]string{".doc", ".docx"}))
			fileSaveDialog.Show()
		}),
	)
}

func NewRenderDocumentGroup(callback func()) *fyne.Container {
	renderDocumentLabel := widget.NewLabel(RenderTemplateLabel)
	renderDocumentButton := widget.NewButton(RenderTemplateButton, callback)
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
			label := widget.NewLabel(PlaceholderLabel)
			label.MinSize()
			return label
		},
		func(id widget.TableCellID, object fyne.CanvasObject) {
			cellContent := (*fileData)[id.Row][id.Col]
			label := object.(*widget.Label)
			label.SetText(cellContent)
		},
	)
	table.SetColumnWidth(0, FilenameColumnWidth)
	table.SetColumnWidth(1, ChecksumColumnWidth)
	table.SetColumnWidth(2, SizeColumnWidth)
	table.SetColumnWidth(3, CreatedColumnWidth)
	table.UpdateHeader = func(id widget.TableCellID, template fyne.CanvasObject) {
		l := template.(*widget.Label)
		if id.Row < 0 {
			l.SetText(TableHeaders[id.Col])
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
	window := mainApp.NewWindow(WindowTitle)
	window.Resize(fyne.NewSize(WindowWidth, WindowHeight))

	var fileData [][]string
	workingDir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	var templateFile = path.Join(workingDir, DefaultTemplatePath)
	var outputFile = path.Join(workingDir, DefaultOutputPath)

	fileTable := CreateFileDataTable(&fileData)
	controlGroup := container.NewVBox(
		NewFolderSelectGroup(window, func(uri fyne.ListableURI, err error) { updateTable(uri, fileTable, &fileData) }),
		NewConfigGroup(window, &templateFile, &outputFile),
		NewRenderDocumentGroup(func() {
			renderTemplate(fileData, &templateFile, &outputFile)
			dialog.NewInformation(
				RenderCompleteLabel,
				fmt.Sprintf(RenderCompleteMsgTemplate, outputFile),
				window,
			).Show()
		}),
	)
	window.SetContent(
		container.NewBorder(
			nil,
			nil,
			controlGroup,
			nil,
			fileTable,
		))
	window.ShowAndRun()
}
