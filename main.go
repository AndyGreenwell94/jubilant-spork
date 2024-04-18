package main

import (
	"errors"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"io/fs"
	"log"
	"os"
	"path"
	"path/filepath"
	"strconv"
)

const (
	OpenLabel                      = "Выбрать"
	SelectFolderLabel              = "Выбраная Папка:"
	PlaceholderLabel               = "Placeholder"
	WindowTitle                    = "Расчет Данных ИУЛ"
	SelectTemplateLabel            = "Выбрать Фаил Шаблона:"
	SelectTemplateButton           = "Выбрать"
	SelectOutputLabel              = "Выбрать Фаил Назначения:"
	SelectOutputButton             = "Выбрать"
	RenderTemplateLabel            = "Заполнить Шаблон:"
	RenderTemplateButton           = "Выполнить"
	RenderCompleteLabel            = "Документ Сформирован"
	RenderCompleteMsgTemplate      = "Документ был успешно сформирован: %s"
	DefaultTemplatePath            = "./template.docx"
	DefaultOutputPath              = "./result.docx"
	WindowWidth                    = 1920
	WindowHeight                   = 1080
	FilenameColumnWidth            = 300
	ChecksumColumnWidth            = 200
	SizeColumnWidth                = 150
	CreatedColumnWidth             = 250
	AuthorTableColumnWidth         = 400
	ControlTableDefaultColumnWidth = 30
)

var fileTableHeaders = [5]string{"Имя Файла", "Контрольная Сумма", "Размер", "Дата Создания", ""}
var authorTableHeaders = [3]string{"Работа", "Имя", ""}

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

func NewControlSheetSelect(window fyne.Window, excelFile *string, callback func()) *fyne.Container {
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
		widget.NewButton("Parse XLSX", callback),
	)
}

func moveFile(slice [][]string, src, dst int) [][]string {
	if src == 0 && dst == -1 {
		return slice
	}
	sliceLen := len(slice)
	if src == sliceLen-1 && dst == sliceLen {
		return slice
	}
	value := slice[src]
	copy(slice[src:], slice[src+1:])
	slice = slice[:len(slice)-1]
	slice = append(slice, []string{})
	copy(slice[dst+1:], slice[dst:])
	slice[dst] = value
	return slice
}

func moveAuthor(slice [][2]string, src, dst int) [][2]string {
	if src == 0 && dst == -1 {
		return slice
	}
	sliceLen := len(slice)
	if src == sliceLen-1 && dst == sliceLen {
		return slice
	}
	value := slice[src]
	copy(slice[src:], slice[src+1:])
	slice = slice[:len(slice)-1]
	slice = append(slice, [2]string{})
	copy(slice[dst+1:], slice[dst:])
	slice[dst] = value
	return slice
}

func CreateFileDataTable(fileData *[][]string) *widget.Table {
	var arrowUp = theme.MenuDropUpIcon()
	var arrowDown = theme.MenuDropDownIcon()
	table := &widget.Table{
		Length: func() (rows int, cols int) {
			rowsCount := len(*fileData)
			if rowsCount == 0 {
				return 0, 0
			}
			colsCount := len((*fileData)[0])
			return rowsCount, colsCount + 1
		},
		CreateCell: func() fyne.CanvasObject {
			return container.NewGridWithRows(1)
		},
		UpdateCell: func(id widget.TableCellID, object fyne.CanvasObject) {},
	}
	table.ExtendBaseWidget(table)
	table.UpdateCell = func(id widget.TableCellID, object fyne.CanvasObject) {
		box := object.(*fyne.Container)
		box.RemoveAll()
		if id.Col == 4 {
			box.Add(widget.NewButtonWithIcon("", arrowUp, func() {
				*fileData = moveFile(*fileData, id.Row, id.Row-1)
				table.Refresh()
			}))
			box.Add(widget.NewButtonWithIcon("", arrowDown, func() {
				*fileData = moveFile(*fileData, id.Row, id.Row+1)
				table.Refresh()
			}))
		} else {
			cellContent := (*fileData)[id.Row][id.Col]
			label := widget.NewLabel(cellContent)
			box.Add(label)
		}
	}
	table.ShowHeaderRow = true
	table.ShowHeaderColumn = true
	table.SetColumnWidth(0, FilenameColumnWidth)
	table.SetColumnWidth(1, ChecksumColumnWidth)
	table.SetColumnWidth(2, SizeColumnWidth)
	table.SetColumnWidth(3, CreatedColumnWidth)
	table.SetColumnWidth(4, CreatedColumnWidth)
	table.UpdateHeader = func(id widget.TableCellID, template fyne.CanvasObject) {
		label := template.(*widget.Label)
		if id.Row < 0 {
			label.SetText(fileTableHeaders[id.Col])
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
	for i := range 4 {
		table.SetColumnWidth(i, ControlTableDefaultColumnWidth)
	}
	return table
}

func CreateAuthorTable(authorsData *[][2]string, distinctAuthors *[]string) *widget.Table {
	var arrowUp = theme.MenuDropUpIcon()
	var arrowDown = theme.MenuDropDownIcon()
	table := &widget.Table{
		Length: func() (rows int, cols int) {
			rowsCount := len(*authorsData)
			if rowsCount == 0 {
				return 0, 0
			}
			colsCount := len((*authorsData)[0])
			return rowsCount, colsCount + 1
		},
		CreateCell: func() fyne.CanvasObject {
			return container.NewGridWithRows(1)
		},
		UpdateCell: func(id widget.TableCellID, object fyne.CanvasObject) {},
	}
	table.UpdateCell = func(id widget.TableCellID, object fyne.CanvasObject) {
		row := (*authorsData)[id.Row]
		cellContent := ""
		if len(row) > id.Col {
			cellContent = (*authorsData)[id.Row][id.Col]
		}
		box := object.(*fyne.Container)
		box.RemoveAll()
		if id.Col == 0 {

			titleSelect := widget.NewSelect(*distinctAuthors, func(s string) {
				(*authorsData)[id.Row][id.Col] = s
			})
			titleSelect.Selected = cellContent
			box.Add(titleSelect)
		} else if id.Col == 1 {
			entry := widget.NewEntry()
			entry.SetText(cellContent)
			entry.OnChanged = func(s string) {
				(*authorsData)[id.Row][id.Col] = s
			}
			box.Add(entry)
		} else if id.Col == 2 {
			box.Add(widget.NewButtonWithIcon("", arrowUp, func() {
				*authorsData = moveAuthor(*authorsData, id.Row, id.Row-1)
				table.Refresh()
			}))
			box.Add(widget.NewButtonWithIcon("", arrowDown, func() {
				*authorsData = moveAuthor(*authorsData, id.Row, id.Row+1)
				table.Refresh()
			}))
			button := widget.NewButtonWithIcon("", theme.DeleteIcon(), func() {
				*authorsData = append((*authorsData)[:id.Row], (*authorsData)[id.Row+1:]...)
				table.Refresh()
			})
			box.Add(button)
		}
	}

	table.UpdateHeader = func(id widget.TableCellID, template fyne.CanvasObject) {
		label := template.(*widget.Label)
		if id.Row < 0 {
			label.SetText(authorTableHeaders[id.Col])
		} else if id.Col < 0 {
			label.SetText(strconv.Itoa(id.Row + 1))
		} else {
			label.SetText("")
		}
	}
	table.ExtendBaseWidget(table)
	table.ShowHeaderRow = true
	table.ShowHeaderColumn = true
	table.SetColumnWidth(0, AuthorTableColumnWidth)
	table.SetColumnWidth(1, AuthorTableColumnWidth)
	table.SetColumnWidth(2, AuthorTableColumnWidth)
	return table
}

func updateFileTable(dir string, fileTable *widget.Table, fileData *[][]string) error {
	files, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	var newFileData [][]string
	for _, file := range files {
		checksum, fileSize, createdAt, err := calculateChecksum(file.Name(), dir)
		if err != nil {
			return err
		}
		newFileData = append(newFileData, []string{file.Name(), checksum, fileSize, createdAt})
	}
	*fileData = newFileData
	fileTable.Refresh()

	return nil
}

func searchExcel(dir string, fileName string) string {
	var foundFile string
	err := filepath.Walk(dir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			fmt.Println("Error:", err)
			return err
		}

		if info.Name() == fileName {
			foundFile = path
			fmt.Println("File found:", path)
		}

		return nil
	})
	if err != nil {
		return ""
	}

	return foundFile
}

func main() {
	mainApp := app.New()
	window := mainApp.NewWindow(WindowTitle)
	window.Resize(fyne.NewSize(WindowWidth, WindowHeight))

	var fileData [][]string
	var controlData [][]string
	var authorData [][2]string
	var distinctAuthors []string
	var excelFileName, excelFileCheck, excelFileCreated, excelSize string
	workingDir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	var templateFile = path.Join(workingDir, DefaultTemplatePath)
	var outputFile = path.Join(workingDir, DefaultOutputPath)
	var excelFile = ""

	controlTable := CreateControlTable(&controlData)
	fileTable := CreateFileDataTable(&fileData)
	authorTable := CreateAuthorTable(&authorData, &distinctAuthors)
	controlGroup := container.NewVBox(
		NewFolderSelectGroup(window, func(uri fyne.ListableURI, err error) {
			err = updateFileTable(uri.Path(), fileTable, &fileData)
			if err != nil {
				dialog.NewError(err, window).Show()
				return
			}
		}),
		NewConfigGroup(window, &templateFile, &outputFile),
		NewControlSheetSelect(window, &excelFile, func() {
			controlData, authorData = ExtractExcelFileData(excelFile)
			controlTable.Refresh()
			authorTable.Refresh()
		}),
		NewRenderDocumentGroup(func() {
			renderTemplate(fileData, controlData, authorData, excelFileName, excelFileCheck, excelSize, excelFileCreated, &templateFile, &outputFile)
			dialog.NewInformation(
				RenderCompleteLabel,
				fmt.Sprintf(RenderCompleteMsgTemplate, outputFile),
				window,
			).Show()
		}),
	)
	window.SetOnDropped(func(position fyne.Position, uris []fyne.URI) {
		if len(uris) != 1 {
			dialog.NewError(
				errors.New("Можно импортировать только 1 папку."),
				window,
			).Show()
			return
		}
		folderUri := uris[0]
		info, err := os.Stat(folderUri.Path())
		if err != nil {
			dialog.NewError(err, window).Show()
			return
		}
		if !info.IsDir() {
			dialog.NewError(
				errors.New("Можно импортировать только 1 папку."),
				window,
			).Show()
			return
		}
		excelFile = searchExcel(filepath.Join(folderUri.Path(), "../.."), folderUri.Name()+".xlsx")

		err = updateFileTable(
			folderUri.Path(),
			fileTable,
			&fileData,
		)
		if err != nil {
			dialog.NewError(err, window).Show()
			return
		}
		if excelFile != "" {
			excelFileName = filepath.Base(excelFile)
			excelFileCheck, excelSize, excelFileCreated, err = calculateChecksum(
				excelFileName,
				filepath.Dir(excelFile),
			)
			if err != nil {
				dialog.NewError(err, window).Show()
				return
			}
			controlData, authorData = ExtractExcelFileData(excelFile)
			distinctAuthors = make([]string, 0)
			seen := make(map[string]bool)
			for _, row := range authorData {
				author := row[0]
				if !seen[author] {
					seen[author] = true
					distinctAuthors = append(distinctAuthors, author)
				}

			}
			controlTable.Refresh()
			authorTable.Refresh()
		}
	})
	tabs := container.NewAppTabs(
		container.NewTabItem("Лист Управленгия", controlTable),
		container.NewTabItem("Файлы", fileTable),
		container.NewTabItem("Авторы", authorTable),
	)
	window.SetContent(
		container.NewBorder(
			nil,
			nil,
			controlGroup,
			nil,
			//container.NewAdaptiveGrid(1, controlTable, fileTable, authorTable),
			tabs,
		))
	window.ShowAndRun()
}
