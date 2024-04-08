package main

import (
	"fmt"
	docxt "github.com/legion-zver/go-docx-templates"
	"hash/crc32"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func calculateChecksum(fileName string, dir string) (string, string, string, error) {
	filePath := filepath.Join(dir, fileName)
	file, err := os.Open(filePath)
	if err != nil {
		return "", "", "", err
	}

	defer func() {
		closeErr := file.Close()
		if closeErr != nil && err == nil { // Update the err return value if it's nil.
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

	return strings.ToUpper(fmt.Sprintf("%x", checksum)), strconv.FormatInt(fileInfo.Size(), 10), fileInfo.ModTime().Format("2006-01-02 15:04"), err
}

func renderTemplate(files [][]string, templateFile *string, outputFile *string) {
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

	if err := template.RenderTemplate(renderData); err != nil {
		log.Fatal(err)
	}
	if err := template.Save(*outputFile); err != nil {
		log.Fatal(err)
	}
}
