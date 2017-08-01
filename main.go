package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
)

func init() {
	// log.SetFormatter(&log.JSONFormatter{})
	// log.SetOutput(os.Stdout)
	log.SetLevel(log.DebugLevel)
}

func main() {
	t := time.Now()
	r, _ := os.Open("test.txt")
	type Search struct {
		After          []byte   // wait until `After` appears beforing looking for activator
		Activators     [][]byte // wait until all these are seen to activate
		Deactivator    []byte
		Limit          int
		Captured       [][]byte
		numActivated   int
		numDeactivated int
		capture        []byte
		i              int
	}
	s := Search{
		After:        []byte("laskjdf"),
		Activators:   [][]byte{[]byte(`<option class="level-0" `), []byte(`>`)},
		Deactivator:  []byte(`</option>`),
		Limit:        10,
		Captured:     [][]byte{},
		numActivated: 0,
		capture:      make([]byte, 500),
		i:            0,
	}

	// Find largest activator/deactivator
	maxSize := 0
	for _, phrase := range append(s.Activators, s.Deactivator) {
		if len(phrase) > maxSize {
			maxSize = len(phrase)
		}
	}

	bytesRead := make([]byte, maxSize*2+2)
	for {
		n, errRead := r.Read(bytesRead)
	StartAgain:
		// Loop, and keep looping while we find activators
		// increment startIndex upon finding activators
		startIndex := 0
		for {
			if s.numActivated == len(s.Activators) {
				log.Debug("Activated")
				break
			}
			index := bytes.Index(bytesRead[startIndex:], s.Activators[s.numActivated])
			if index > -1 {
				log.Debug("Activating")
				startIndex = startIndex + index + len(s.Activators[s.numActivated])
				s.numActivated++
				continue
			}
			break
		}

		if s.numActivated == len(s.Activators) {
			// Check if deactivated
			deactivated := false
			endIndex := bytes.Index(bytesRead[startIndex:], s.Deactivator)
			if endIndex > -1 {
				log.Debug("Deactivating")
				deactivated = true
				endIndex += startIndex
			} else {
				endIndex = n
			}

			// add bytes from known activator to deactivator
			log.Debug(startIndex, endIndex)
			if startIndex > endIndex {
				// edge case where you have another thing starting as one is ending
				log.Debug("Edge case")
				for _, b := range bytesRead[0:endIndex] {
					s.capture[s.i] = b
					s.i++
				}
				log.Debugf("Added %d bytes", endIndex)
			} else {
				for _, b := range bytesRead[startIndex:endIndex] {
					s.capture[s.i] = b
					s.i++
				}
				log.Debugf("Added %d bytes", endIndex-startIndex)
			}

			if deactivated {
				log.Debug("Reseting")
				// add and reset
				s.Captured = append(s.Captured, s.capture[:s.i])
				s.numActivated = 0
				s.i = 0
				s.capture = make([]byte, 1000)

				bytesRead = bytesRead[endIndex+len(s.Deactivator):]
				goto StartAgain
			}
		}

		if errRead == io.EOF {
			break
		}
	}
	r.Close()
	fmt.Println(time.Since(t))
	for _, r := range s.Captured {
		fmt.Println(string(r))
	}
}
