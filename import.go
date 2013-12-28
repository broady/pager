package pager

import (
	"io/ioutil"
	"log"
	"path/filepath"

	proto "code.google.com/p/goprotobuf/proto"
	"config"
)

func readNotifications() {
	dirs, err := ioutil.ReadDir("notifications")
	check("could not read dir", err)
	for _, configFile := range dirs {
		n := configFile.Name()
		if n[0] == '.' {
			continue
		}
		pb, err := ioutil.ReadFile(filepath.Join("notifications", n))
		check("could not read", err, n)
		c := new(config.NotificationSequence)
		err = proto.UnmarshalText(string(pb), c)
		check("could not unmarshal", err, n)
		notifications[n] = c
	}
}

func readConfigs() {
	// TODO: add error checking for regexp.
	dirs, err := ioutil.ReadDir("configs")
	check("could not read dir", err)
	for _, configFile := range dirs {
		n := configFile.Name()
		if n[0] == '.' {
			continue
		}
		pb, err := ioutil.ReadFile(filepath.Join("configs", n))
		check("could not read", err, n)
		c := new(config.PagerConfig)
		err = proto.UnmarshalText(string(pb), c)
		check("could not unmarshal", err, n)
		configs[n] = c
	}
}

func check(msg string, err error, a ...interface{}) {
	if err != nil {
		if len(a) != 0 {
			log.Fatalf(msg+": %v", a, err)
		} else {
			log.Fatalf(msg+": %v", err)
		}
	}
}
