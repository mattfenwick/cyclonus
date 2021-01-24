package utils

import (
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
)

func DoOrDie(err error) {
	if err != nil {
		log.Fatalf("%+v", err)
	}
}

func PrintJson(obj interface{}) {
	bytes, err := json.MarshalIndent(obj, "", "  ")
	DoOrDie(err)
	fmt.Printf("%s\n", bytes)
}
