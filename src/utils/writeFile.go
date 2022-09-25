package utils

import (
	"io"
	"log"
	"os"
)

func checkFileIsExist(filename string) bool {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return false
	}
	return true
}

func WriteFile(filename string, content string) {
	var f *os.File
	var err1 error
	if checkFileIsExist(filename) { // if file exists
		err1 = os.Remove(filename) // remove this file
	}
	f, err1 = os.Create(filename) // if file not exists, create the file

	if err1 != nil {
		log.Fatal(err1)
	}
	defer f.Close()
	io.WriteString(f, content)
}
