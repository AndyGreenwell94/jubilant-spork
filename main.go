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
	SelectOutputLabel         = "Выбрать Фаил Назначения:"
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
			if err != nil || uri == nil {
				return
			}
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
				if err != nil || closer == nil {
					return
				}
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
				if err != nil || closer == nil {
					return
				}
				*outputFile = closer.URI().Path()
				selectedOutputPath.SetText(*outputFile)
				err = closer.Close()
				if err != nil {
					return
				}
			}, window)
			fileSaveDialog.SetFilter(storage.NewExtensionFileFilter([]string{".doc", ".docx"}))
			fileSaveDialog.SetFileName("result.docx")
			fileSaveDialog.Show()
		}),
	)
}

func NewRenderDocumentGroup(callback func()) *fyne.Container {
	renderDocumentLabel := widget.NewLabel(RenderTemplateLabel)
	renderDocumentButton := widget.NewButton(RenderTemplateButton, callback)
	return container.NewVBox(renderDocumentLabel, renderDocumentButton)
}

func NewControlSheetSelect(window fyne.Window, excelFile *string, calback func()) *fyne.Container {
	labelText := "Selected XLSX: %s"
	label := widget.NewLabel(fmt.Sprintf(labelText, ""))
	return container.NewVBox(
		label,
		widget.NewButton("Open", func() {
			dialog.ShowFileOpen(func(closer fyne.URIReadCloser, err error) {
				if err != nil || closer == nil {
					return
				}
				*excelFile = closer.URI().Path()
				label.SetText(fmt.Sprintf(labelText, *excelFile))
				err = closer.Close()
				if err != nil {
					return
				}
			}, window)
		}),
		widget.NewButton("Parse XLSX", calback),
	)
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
		label := template.(*widget.Label)
		if id.Row < 0 {
			label.SetText(TableHeaders[id.Col])
		} else if id.Col < 0 {
			label.SetText(strconv.Itoa(id.Row + 1))
		} else {
			label.SetText("")
		}
	}
	return table
}

func CreateControlTable(controlData *[][]string) *widget.Table {
	table := widget.NewTableWithHeaders(
		func() (rows int, cols int) {
			rowsCount := len(*controlData)
			if rowsCount == 0 {
				return 0, 0
			}
			colsCount := 0
			for rowIndex := range len(*controlData) {
				rowLen := len((*controlData)[rowIndex])
				if rowLen > colsCount {
					colsCount = rowLen
				}
			}
			return rowsCount, colsCount
		},
		func() fyne.CanvasObject {
			label := widget.NewLabel(PlaceholderLabel)
			label.MinSize()
			return label
		},
		func(id widget.TableCellID, object fyne.CanvasObject) {
			row := (*controlData)[id.Row]
			label := object.(*widget.Label)
			cellContent := ""
			if len(row) > id.Col {
				cellContent = (*controlData)[id.Row][id.Col]
			}
			label.SetText(cellContent)
		},
	)
	return table
}

func CreateAuthorTable(authorData *[][2]string) *widget.Table {
	table := widget.NewTableWithHeaders(
		func() (rows int, cols int) {
			rowsCount := len(*authorData)
			if rowsCount == 0 {
				return 0, 0
			}
			colsCount := 0
			for rowIndex := range len(*authorData) {
				rowLen := len((*authorData)[rowIndex])
				if rowLen > colsCount {
					colsCount = rowLen
				}
			}
			return rowsCount, colsCount
		},
		func() fyne.CanvasObject {
			label := widget.NewLabel(PlaceholderLabel)
			label.MinSize()
			return label
		},
		func(id widget.TableCellID, object fyne.CanvasObject) {
			row := (*authorData)[id.Row]
			label := object.(*widget.Label)
			cellContent := ""
			if len(row) > id.Col {
				cellContent = (*authorData)[id.Row][id.Col]
			}
			label.SetText(cellContent)
		},
	)
	return table
}

func updateFileTable(dir string, fileTable *widget.Table, fileData *[][]string) {
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

	var fileData = [][]string{{"", "", "", ""}}
	var controlData = [][]string{{}}
	var authorData = [][2]string{{}}
	workingDir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	var templateFile = path.Join(workingDir, DefaultTemplatePath)
	var outputFile = path.Join(workingDir, DefaultOutputPath)
	var excelFile = ""

	controlTable := CreateControlTable(&controlData)
	fileTable := CreateFileDataTable(&fileData)
	authorTable := CreateAuthorTable(&authorData)
	controlGroup := container.NewVBox(
		NewFolderSelectGroup(window, func(uri fyne.ListableURI, err error) {
			updateFileTable(uri.Path(), fileTable, &fileData)
		}),
		NewConfigGroup(window, &templateFile, &outputFile),
		NewControlSheetSelect(window, &excelFile, func() {
			controlData, authorData = ExtractExcelFileData(excelFile)
			controlTable.Refresh()
			authorTable.Refresh()
		}),
		NewRenderDocumentGroup(func() {
			renderTemplate(fileData, controlData, authorData, &templateFile, &outputFile)
			dialog.NewInformation(
				RenderCompleteLabel,
				fmt.Sprintf(RenderCompleteMsgTemplate, outputFile),
				window,
			).Show()
		}),
	)
	window.SetOnDropped(func(position fyne.Position, uris []fyne.URI) {
		fmt.Println(uris)
		updateFileTable(
			uris[0].Path(),
			fileTable,
			&fileData,
		)
	})
	window.SetContent(
		container.NewBorder(
			nil,
			nil,
			controlGroup,
			nil,
			container.NewVSplit(controlTable, container.NewHSplit(fileTable, authorTable)),
		))
	window.ShowAndRun()
}
