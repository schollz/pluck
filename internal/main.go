package main

import (
	"bufio"
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
	"strconv"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

func init() {
	// log.SetFormatter(&log.JSONFormatter{})
	// log.SetOutput(os[i].Stdout)
	log.SetLevel(log.WarnLevel)
}

func main() {

}

func parseFile(f string) string {

	r1, _ := os.Open("test.txt")
	r := bufio.NewReader(r1)

	type Config struct {
		Activators  []string
		Deactivator string
		Name        string
		Limit       int
	}
	var config []Config
	b, _ := ioutil.ReadFile("config.yaml")
	err := yaml.Unmarshal([]byte(b), &config)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	type Search struct {
		Activators   [][]byte // wait until all these are seen to activate
		Deactivator  []byte
		Limit        int
		Name         string
		captured     [][]byte
		numActivated int
		captureByte  []byte
		captureI     int
		activeI      int
		deactiveI    int
	}
	s := make([]Search, len(config))
	for i := range config {
		if config[i].Limit == 0 {
			config[i].Limit = 1
		}
		if config[i].Name == "" {
			config[i].Name = strconv.Itoa(i)
		}
		s[i] = Search{
			Activators:   make([][]byte, len(config[i].Activators)),
			Deactivator:  []byte(config[i].Deactivator),
			Limit:        config[i].Limit,
			Name:         config[i].Name,
			captured:     [][]byte{},
			numActivated: 0,
			captureByte:  make([]byte, 10000),
			captureI:     0,
			activeI:      0,
			deactiveI:    0,
		}
		for j := range config[i].Activators {
			s[i].Activators[j] = []byte(config[i].Activators[j])
		}
	}

	for {
		curByte, errRead := r.ReadByte()
		for i := range s {
			if len(s[i].captured) == s[i].Limit {
				continue
			}
			if s[i].numActivated < len(s[i].Activators) {
				// look for activators
				if curByte == s[i].Activators[s[i].numActivated][s[i].activeI] {
					s[i].activeI++
					if s[i].activeI == len(s[i].Activators[s[i].numActivated]) {
						log.Info(string(curByte), "Activated")
						s[i].numActivated++
						s[i].activeI = 0
					}
				} else {
					s[i].activeI = 0
				}
			} else {
				// add to capture
				s[i].captureByte[s[i].captureI] = curByte
				s[i].captureI++
				// look for deactivators
				if curByte == s[i].Deactivator[s[i].deactiveI] {
					s[i].deactiveI++
					if s[i].deactiveI == len(s[i].Deactivator) {
						log.Info(string(curByte), "Deactivated")
						// add capture
						log.Info(string(s[i].captureByte[:s[i].captureI-len(s[i].Deactivator)]))
						tempByte := make([]byte, s[i].captureI-len(s[i].Deactivator))
						copy(tempByte, s[i].captureByte[:s[i].captureI-len(s[i].Deactivator)])
						s[i].captured = append(s[i].captured, tempByte)
						// reset
						s[i].numActivated = 0
						s[i].deactiveI = 0
						s[i].captureI = 0
					}
				} else {
					s[i].activeI = 0
				}
			}

		}

		if errRead == io.EOF {
			break
		}
	}
	r1.Close()
	result := make(map[string]interface{})
	for i := range s {
		if len(s[i].captured) == 1 {
			result[s[i].Name] = string(s[i].captured[0])
		} else {
			results := make([]string, len(s[i].captured))
			for j, r := range s[i].captured {
				results[j] = string(r)
			}
			result[s[i].Name] = results
		}
	}
	resultJson, err := json.MarshalIndent(result, "", "    ")
	if err != nil {
		log.Error(errors.Wrap(err, "result marshalling failed"))
	}
	return string(resultJson)
}
