package main

import (
	"fmt"
	"github.com/AndyGreenwell94/docxt"
	"github.com/xuri/excelize/v2"
	"hash/crc32"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type CheckedFile struct {
	FileName  string
	Checksum  string
	FileSize  string
	CreatedAt string
}

type Author struct {
	Name  string
	Title string
}

type RenderData struct {
	Items   []CheckedFile
	Excel   CheckedFile
	Control map[string]string
	Authors []Author
}

func calculateChecksum(fileName string, dir string) (string, string, string, error) {
	filePath := filepath.Join(dir, fileName)
	file, err := os.Open(filePath)
	if err != nil {
		return "", "", "", err
	}

	defer func() {
		closeErr := file.Close()
		if closeErr != nil && err == nil {
			err = closeErr
		}
	}()

	hasher := crc32.NewIEEE()
	if _, err = io.Copy(hasher, file); err != nil {
		return "", "", "", err
	}
	checksum := hasher.Sum32()
	fileInfo, err := file.Stat()
	if err != nil {
		return "", "", "", err
	}

	return strings.ToUpper(fmt.Sprintf("%x", checksum)), strconv.FormatInt(fileInfo.Size(), 10), fileInfo.ModTime().Format("2006.01.02_15:04"), err
}

func renderTemplate(files [][]string, controlData [][]string, authorsData [][2]string, excelFileName, excelCheck, excelSize, excelCreatedAt string, templateFile *string, outputFile *string) {
	template, err := docxt.OpenTemplate(*templateFile)
	if err != nil {
		log.Fatal(err)
	}
	renderData := new(RenderData)
	for _, file := range files {
		renderData.Items = append(renderData.Items, CheckedFile{
			FileName:  file[0],
			Checksum:  file[1],
			FileSize:  file[2],
			CreatedAt: file[3],
		})
	}
	renderData.Excel = CheckedFile{FileName: excelFileName, Checksum: excelCheck, FileSize: excelSize, CreatedAt: excelCreatedAt}
	renderData.Control = make(map[string]string)
	for rowNum, controlRow := range controlData {
		for colNum, controlCol := range controlRow {
			name, err := excelize.CoordinatesToCellName(colNum+1, rowNum+1)
			if err != nil {
				log.Printf("Error")
			}
			renderData.Control[name] = controlCol
		}
	}
	for _, authorData := range authorsData {
		renderData.Authors = append(renderData.Authors, Author{
			Name:  authorData[1],
			Title: authorData[0],
		})
	}

	if err := template.RenderTemplate(renderData); err != nil {
		log.Fatal(err)
	}
	if err := template.Save(*outputFile); err != nil {
		log.Fatal(err)
	}
}
