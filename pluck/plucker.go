package pluck

import (
	"bufio"
	"bytes"
	"encoding/json"
	"html"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/BurntSushi/toml"
	"github.com/pkg/errors"
	"github.com/schollz/pluck/pluck/striphtml"
	log "github.com/sirupsen/logrus"
)

// Config specifies parameters for plucking
type Config struct {
	Activators  []string // must be found in order, before capturing commences
	Permanent   int      // number of activators that stay permanently (counted from left to right)
	Deactivator string   // restarts capturing
	Finisher    string   // finishes capturing this pluck
	Limit       int      // specifies the number of times capturing can occur
	Name        string   // the key in the returned map, after completion
	Sanitize    bool
	Maximum     int // maximum number of characters for a capture
}

type configs struct {
	Pluck []Config
}

// Plucker stores the result and the types of things to pluck
type Plucker struct {
	pluckers []pluckUnit
	result   map[string]interface{}
}

type pluckUnit struct {
	config       Config
	activators   [][]byte
	permanent    int
	maximum      int
	deactivator  []byte
	finisher     []byte
	captured     [][]byte
	numActivated int
	captureByte  []byte
	captureI     int
	activeI      int
	deactiveI    int
	finisherI    int
	isFinished   bool
}

// New returns a new plucker
// which can later have items added to it
// or can load a config file
// and then can be used to parse.
func New() (*Plucker, error) {
	log.SetLevel(log.WarnLevel)
	p := new(Plucker)
	p.pluckers = []pluckUnit{}
	return p, nil
}

// Verbose toggles debug mode
func (p *Plucker) Verbose(makeVerbose bool) {
	if makeVerbose {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.WarnLevel)
	}
}

// Configuration returns an array of the current
// Config for each plucker.
func (p *Plucker) Configuration() (c []Config) {
	c = make([]Config, len(p.pluckers))
	for i, unit := range p.pluckers {
		c[i] = unit.config
	}
	return
}

// Add adds a unit
// to pluck with specified parameters
func (p *Plucker) Add(c Config) {
	var u pluckUnit
	u.config = c
	if u.config.Limit == 0 {
		u.config.Limit = -1
	}
	if u.config.Name == "" {
		u.config.Name = strconv.Itoa(len(p.pluckers))
	}
	u.activators = make([][]byte, len(c.Activators))
	for i := range c.Activators {
		u.activators[i] = []byte(c.Activators[i])
	}
	u.permanent = c.Permanent
	u.deactivator = []byte(c.Deactivator)
	if len(c.Finisher) > 0 {
		u.finisher = []byte(c.Finisher)
	} else {
		u.finisher = nil
	}
	u.maximum = -1
	if c.Maximum > 0 {
		u.maximum = c.Maximum
	}
	u.captureByte = make([]byte, 100000)
	u.captured = [][]byte{}
	p.pluckers = append(p.pluckers, u)
	log.Infof("Added plucker %+v", c)
}

// Load will load a YAML configuration file of untis
// to pluck with specified parameters
func (p *Plucker) Load(f string) (err error) {
	tomlData, err := ioutil.ReadFile(f)
	if err != nil {
		return errors.Wrap(err, "problem opening "+f)
	}
	log.Debugf("toml string: %s", string(tomlData))
	p.LoadFromString(string(tomlData))
	return
}

// LoadFromString will load a YAML configuration file of untis
// to pluck with specified parameters
func (p *Plucker) LoadFromString(tomlString string) (err error) {
	var conf configs
	_, err = toml.Decode(tomlString, &conf)
	log.Debugf("Loaded toml: %+v", conf)
	for i := range conf.Pluck {
		var c Config
		c.Activators = conf.Pluck[i].Activators
		c.Deactivator = conf.Pluck[i].Deactivator
		c.Finisher = conf.Pluck[i].Finisher
		c.Limit = conf.Pluck[i].Limit
		c.Name = conf.Pluck[i].Name
		c.Permanent = conf.Pluck[i].Permanent
		c.Sanitize = conf.Pluck[i].Sanitize
		c.Maximum = conf.Pluck[i].Maximum
		p.Add(c)
	}
	return
}

// PluckString takes a string as input
// and uses the specified parameters and generates
// a map (p.result) with the finished results.
// The streaming can be enabled by setting it to true.
func (p *Plucker) PluckString(s string, stream ...bool) (err error) {
	r := bufio.NewReader(strings.NewReader(s))
	if len(stream) > 0 && stream[0] {
		return p.PluckStream(r)
	}
	return p.Pluck(r)
}

// PluckFile takes a file as input
// and uses the specified parameters and generates
// a map (p.result) with the finished results. The streaming
// can be enabled by setting it to true.
func (p *Plucker) PluckFile(f string, stream ...bool) (err error) {
	r1, err := os.Open(f)
	defer r1.Close()
	if err != nil {
		return
	}
	r := bufio.NewReader(r1)
	if len(stream) > 0 && stream[0] {
		return p.PluckStream(r)
	}
	return p.Pluck(r)
}

// PluckURL takes a URL as input
// and uses the specified parameters and generates
// a map (p.result) with the finished results
func (p *Plucker) PluckURL(url string, stream ...bool) (err error) {
	client := &http.Client{}
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return
	}
	request.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:52.0) Gecko/20100101 Firefox/52.0")
	resp, err := client.Do(request)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	r := bufio.NewReader(resp.Body)
	if len(stream) > 0 && stream[0] {
		return p.PluckStream(r)
	}
	return p.Pluck(r)
}

// Pluck takes a buffered reader stream and
// extracts the text from it. This spawns a thread for
// each plucker and copies the entire buffer to memory,
// so that each plucker works in parallel.
func (p *Plucker) Pluck(r *bufio.Reader) (err error) {
	allBytes, _ := r.ReadBytes(0)
	var wg sync.WaitGroup
	wg.Add(len(p.pluckers))
	for i := 0; i < len(p.pluckers); i++ {
		go func(i int, allBytes []byte) {
			defer wg.Done()
			for _, curByte := range allBytes {
				if p.pluckers[i].numActivated < len(p.pluckers[i].activators) {
					// look for activators
					if curByte == p.pluckers[i].activators[p.pluckers[i].numActivated][p.pluckers[i].activeI] {
						p.pluckers[i].activeI++
						if p.pluckers[i].activeI == len(p.pluckers[i].activators[p.pluckers[i].numActivated]) {
							log.Info(string(curByte), "Activated")
							p.pluckers[i].numActivated++
							p.pluckers[i].activeI = 0
						}
					} else {
						p.pluckers[i].activeI = 0
					}
				} else {
					// add to capture
					p.pluckers[i].captureByte[p.pluckers[i].captureI] = curByte
					p.pluckers[i].captureI++
					// look for deactivators
					if curByte == p.pluckers[i].deactivator[p.pluckers[i].deactiveI] {
						p.pluckers[i].deactiveI++
						if p.pluckers[i].deactiveI == len(p.pluckers[i].deactivator) {
							log.Info(string(curByte), "Deactivated")
							// add capture
							log.Info(string(p.pluckers[i].captureByte[:p.pluckers[i].captureI-len(p.pluckers[i].deactivator)]))
							tempByte := make([]byte, p.pluckers[i].captureI-len(p.pluckers[i].deactivator))
							copy(tempByte, p.pluckers[i].captureByte[:p.pluckers[i].captureI-len(p.pluckers[i].deactivator)])
							if p.pluckers[i].config.Sanitize {
								tempByte = bytes.Replace(tempByte, []byte("\\u003c"), []byte("<"), -1)
								tempByte = bytes.Replace(tempByte, []byte("\\u003e"), []byte(">"), -1)
								tempByte = bytes.Replace(tempByte, []byte("\\u0026"), []byte("&"), -1)
								tempByte = []byte(striphtml.StripTags(html.UnescapeString(string(tempByte))))
							}
							tempByte = bytes.TrimSpace(tempByte)
							if p.pluckers[i].maximum < 1 || len(tempByte) < p.pluckers[i].maximum {
								p.pluckers[i].captured = append(p.pluckers[i].captured, tempByte)
							}
							// reset
							p.pluckers[i].numActivated = p.pluckers[i].permanent
							p.pluckers[i].deactiveI = 0
							p.pluckers[i].captureI = 0
						}
					} else {
						p.pluckers[i].activeI = 0
						p.pluckers[i].deactiveI = 0
					}
				}

				// look for finisher
				if p.pluckers[i].finisher != nil && len(p.pluckers[i].captured) > 0 {
					if curByte == p.pluckers[i].finisher[p.pluckers[i].finisherI] {
						p.pluckers[i].finisherI++
						if p.pluckers[i].finisherI == len(p.pluckers[i].finisher) {
							log.Info(string(curByte), "Finished")
							p.pluckers[i].isFinished = true
						}
					} else {
						p.pluckers[i].finisherI = 0
					}
				}

				if len(p.pluckers[i].captured) == p.pluckers[i].config.Limit {
					p.pluckers[i].isFinished = true
				}
				if p.pluckers[i].isFinished {
					break
				}
			}
			log.Infof("plucker %d finished", i)
		}(i, allBytes)
	}
	wg.Wait()
	p.generateResult()
	return
}

// PluckStream takes a buffered reader stream and streams one
// byte at a time and processes all pluckers serially and
// simultaneously.
func (p *Plucker) PluckStream(r *bufio.Reader) (err error) {
	var finished bool
	for {
		curByte, errRead := r.ReadByte()
		if errRead == io.EOF || finished {
			break
		}
		finished = true
		for i := range p.pluckers {
			if p.pluckers[i].isFinished {
				continue
			}
			finished = false
			if p.pluckers[i].numActivated < len(p.pluckers[i].activators) {
				// look for activators
				if curByte == p.pluckers[i].activators[p.pluckers[i].numActivated][p.pluckers[i].activeI] {
					p.pluckers[i].activeI++
					if p.pluckers[i].activeI == len(p.pluckers[i].activators[p.pluckers[i].numActivated]) {
						log.Info(string(curByte), "Activated")
						p.pluckers[i].numActivated++
						p.pluckers[i].activeI = 0
					}
				} else {
					p.pluckers[i].activeI = 0
				}
			} else {
				// add to capture
				p.pluckers[i].captureByte[p.pluckers[i].captureI] = curByte
				p.pluckers[i].captureI++
				// look for deactivators
				if curByte == p.pluckers[i].deactivator[p.pluckers[i].deactiveI] {
					p.pluckers[i].deactiveI++
					if p.pluckers[i].deactiveI == len(p.pluckers[i].deactivator) {
						log.Info(string(curByte), "Deactivated")
						// add capture
						log.Info(string(p.pluckers[i].captureByte[:p.pluckers[i].captureI-len(p.pluckers[i].deactivator)]))
						tempByte := make([]byte, p.pluckers[i].captureI-len(p.pluckers[i].deactivator))
						copy(tempByte, p.pluckers[i].captureByte[:p.pluckers[i].captureI-len(p.pluckers[i].deactivator)])
						if p.pluckers[i].config.Sanitize {
							tempByte = bytes.Replace(tempByte, []byte("\\u003c"), []byte("<"), -1)
							tempByte = bytes.Replace(tempByte, []byte("\\u003e"), []byte(">"), -1)
							tempByte = bytes.Replace(tempByte, []byte("\\u0026"), []byte("&"), -1)
							tempByte = []byte(striphtml.StripTags(html.UnescapeString(string(tempByte))))
						}
						tempByte = bytes.TrimSpace(tempByte)
						p.pluckers[i].captured = append(p.pluckers[i].captured, tempByte)
						// reset
						p.pluckers[i].numActivated = p.pluckers[i].permanent
						p.pluckers[i].deactiveI = 0
						p.pluckers[i].captureI = 0
					}
				} else {
					p.pluckers[i].activeI = 0
					p.pluckers[i].deactiveI = 0
				}
			}

			// look for finisher
			if p.pluckers[i].finisher != nil {
				if curByte == p.pluckers[i].finisher[p.pluckers[i].finisherI] {
					p.pluckers[i].finisherI++
					if p.pluckers[i].finisherI == len(p.pluckers[i].finisher) {
						log.Info(string(curByte), "Finished")
						p.pluckers[i].isFinished = true
					}
				} else {
					p.pluckers[i].finisherI = 0
				}
			}

			if len(p.pluckers[i].captured) == p.pluckers[i].config.Limit {
				p.pluckers[i].isFinished = true
			}
		}
	}
	p.generateResult()
	return
}

func (p *Plucker) generateResult() {
	p.result = make(map[string]interface{})
	for i := range p.pluckers {
		if len(p.pluckers[i].captured) == 1 {
			p.result[p.pluckers[i].config.Name] = string(p.pluckers[i].captured[0])
		} else {
			results := make([]string, len(p.pluckers[i].captured))
			for j, r := range p.pluckers[i].captured {
				results[j] = string(r)
			}
			p.result[p.pluckers[i].config.Name] = results
		}
	}
}

// ResultJSON returns the result, formatted as JSON.
// If their are no results, it returns an empty string.
func (p *Plucker) ResultJSON(indent ...bool) string {
	totalResults := 0
	for key := range p.result {
		b, _ := json.Marshal(p.result[key])
		totalResults += len(b)
	}
	if totalResults == len(p.result)*2 { // results == 2 because its just []
		return ""
	}
	var err error
	var resultJSON []byte
	if len(indent) > 0 && indent[0] {
		resultJSON, err = json.MarshalIndent(p.result, "", "    ")
	} else {
		resultJSON, err = json.Marshal(p.result)
	}
	if err != nil {
		log.Error(errors.Wrap(err, "result marshalling failed"))
	}
	return string(resultJSON)
}

// Result returns the raw result
func (p *Plucker) Result() map[string]interface{} {
	return p.result
}
