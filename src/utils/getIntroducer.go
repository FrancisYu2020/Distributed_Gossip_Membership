package utils

import (
	"fmt"
	"os"
)

func IsIntroducer() bool {
	if _, err := os.Stat("~/introducer"); err == nil {
		fmt.Println("I am the introducer!");
	}
}