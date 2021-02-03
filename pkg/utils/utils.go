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

func JsonString(obj interface{}) string {
	bytes, err := json.MarshalIndent(obj, "", "  ")
	DoOrDie(err)
	return string(bytes)
}

func PrintJson(obj interface{}) {
	fmt.Printf("%s\n", JsonString(obj))
}
