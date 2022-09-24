package utils

import "time"

func GenerateID(name string) string {
	timeStamp := time.Now().Format("2006-01-02 15:04:05")
	return name + " Machine " + timeStamp
}
