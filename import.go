package pager

import (
	"io/ioutil"
	"log"
	"path/filepath"

	proto "code.google.com/p/goprotobuf/proto"
	"config"
)

func readNotifications() {
	dirs, err := ioutil.ReadDir("pagers")
	check("could not read dir", err)
	for _, configFile := range dirs {
		n := configFile.Name()
		if n[0] == '.' {
			continue
		}
		pb, err := ioutil.ReadFile(filepath.Join("pagers", n))
		check("could not read", err, n)
		c := new(config.NotificationSequence)
		err = proto.UnmarshalText(string(pb), c)
		check("could not unmarshal", err, n)
		pagers[n] = c
	}
}

func readMatchers() {
	// TODO: add error checking for regexp.
	dirs, err := ioutil.ReadDir("matchers")
	check("could not read dir", err)
	for _, configFile := range dirs {
		n := configFile.Name()
		if n[0] == '.' {
			continue
		}
		pb, err := ioutil.ReadFile(filepath.Join("matchers", n))
		check("could not read", err, n)
		c := new(config.PagerConfig)
		err = proto.UnmarshalText(string(pb), c)
		check("could not unmarshal", err, n)
		matchers[n] = c
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
