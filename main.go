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
		Deactivators   [][]byte
		Limit          int
		Captured       [][]byte
		numActivated   int
		numDeactivated int
		capture        []byte
		i              int
	}
	s := Search{
		After:          []byte("laskjdf"),
		Activators:     [][]byte{[]byte(`<option class="level-0" `), []byte(`>`)},
		Deactivators:   [][]byte{[]byte(`</option>`)},
		Limit:          10,
		Captured:       [][]byte{},
		numActivated:   0,
		numDeactivated: 0,
		capture:        make([]byte, 50),
		i:              0,
	}

	// Find largest activator/deactivator
	maxSize := 0
	for _, phrase := range append(s.Activators, s.Deactivators...) {
		if len(phrase) > maxSize {
			maxSize = len(phrase)
		}
	}

	bytesRead := make([]byte, maxSize*2+2)
	for {
		n, errRead := r.Read(bytesRead)
		indexActive := -1

		// Only try to activate if there's still things to check
		if s.numActivated < len(s.Activators) {
			indexActive = bytes.Index(bytesRead, s.Activators[s.numActivated])
			if indexActive > -1 {
				log.Debug("Activating")
				s.numActivated++
			}
		}

		if len(s.Activators) == s.numActivated && len(s.Captured) < s.Limit {

			// Check if we should deactivate
			indexDeactivate := bytes.Index(bytesRead, s.Deactivators[s.numDeactivated])
			if indexDeactivate > -1 {
				log.Debug("Deactivating")
				s.numDeactivated++
				if len(s.Deactivators) == s.numDeactivated {
					// add bytes up to deactivator
					for _, b := range bytesRead[n-indexDeactivate:] {
						s.capture[s.i] = b
						s.i++
					}
					log.Debugf("Added %d bytes", n-indexDeactivate)

					// add and reset
					s.Captured = append(s.Captured, s.capture[:s.i])
					s.numActivated = 0
					s.numDeactivated = 0
					s.i = 0
					s.capture = make([]byte, 1000)
				}
			}

			// if not reset, then add bytes
			if s.numActivated == len(s.Activators) {
				if indexActive > -1 {
					indexActive += len(s.Activators[s.numActivated-1])
				} else {
					indexActive = 0
				}
				for _, b := range bytesRead[indexActive:] {
					s.capture[s.i] = b
					s.i++
				}
				log.Debugf("Added %d bytes", n)
			}

		}

		if errRead == io.EOF {
			break
		}
	}
	r.Close()
	fmt.Printf("%+v", s)
	fmt.Println(time.Since(t))
	for _, r := range s.Captured {
		fmt.Println(string(r))
	}
}
