package utils

import log "github.com/sirupsen/logrus"

func DoOrDie(err error) {
	if err != nil {
		log.Fatalf("%+v", err)
	}
}
