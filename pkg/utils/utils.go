package utils

import (
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"sigs.k8s.io/yaml"
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

func YamlString(obj interface{}) string {
	bytes, err := yaml.Marshal(obj)
	DoOrDie(err)
	return string(bytes)
}

func PrintJson(obj interface{}) {
	fmt.Printf("%s\n", JsonString(obj))
}
