package pluck

import (
	"bufio"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// Config specifies parameters for plucking
type Config struct {
	Activators  []string // must be found in order, before capturing commences
	Deactivator string   // stops capturing
	Limit       int      // specifies the number of times capturing can occur
	Name        string   // the key in the returned map, after completion
}

type configs struct {
	Pluck []Config
}

type plucker struct {
	pluckers []pluckUnit
	result   map[string]interface{}
}

type pluckUnit struct {
	config       Config
	activators   [][]byte
	deactivator  []byte
	captured     [][]byte
	numActivated int
	captureByte  []byte
	captureI     int
	activeI      int
	deactiveI    int
}

// New returns a new plucker
// which can later have items added to it
// or can load a config file
// and then can be used to parse.
func New() (*plucker, error) {
	log.SetLevel(log.WarnLevel)
	p := new(plucker)
	p.pluckers = []pluckUnit{}
	return p, nil
}

// Verbose toggles debug mode
func (p *plucker) Verbose(makeVerbose bool) {
	if makeVerbose {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.WarnLevel)
	}
}

// Add adds a unit
// to pluck with specified parameters
func (p *plucker) Add(c Config) {
	var u pluckUnit
	u.config = c
	if u.config.Limit == 0 {
		u.config.Limit = 1
	}
	if u.config.Name == "" {
		u.config.Name = strconv.Itoa(len(p.pluckers))
	}
	u.activators = make([][]byte, len(c.Activators))
	for i := range c.Activators {
		u.activators[i] = []byte(c.Activators[i])
	}
	u.deactivator = []byte(c.Deactivator)
	u.captureByte = make([]byte, 10000)
	u.captured = [][]byte{}
	p.pluckers = append(p.pluckers, u)
}

// Load will load a YAML configuration file of untis
// to pluck with specified parameters
func (p *plucker) Load(f string) (err error) {
	var conf configs
	tomlData, err := ioutil.ReadFile(f)
	if err != nil {
		return errors.Wrap(err, "problem opening "+f)
	}
	log.Debugf("toml string: %s", string(tomlData))
	_, err = toml.Decode(string(tomlData), &conf)
	log.Debugf("Loaded toml: %+v", conf)
	for i := range conf.Pluck {
		var c Config
		c.Activators = conf.Pluck[i].Activators
		c.Deactivator = conf.Pluck[i].Deactivator
		c.Limit = conf.Pluck[i].Limit
		c.Name = conf.Pluck[i].Name
		p.Add(c)
	}
	return
}

// PluckString takes a string as input
// and uses the specified parameters and generates
// a map (p.result) with the finished results
func (p *plucker) PluckString(s string) (err error) {
	r := bufio.NewReader(strings.NewReader(s))
	return p.pluck(r)
}

// PluckFile takes a file as input
// and uses the specified parameters and generates
// a map (p.result) with the finished results
func (p *plucker) PluckFile(f string) (err error) {
	r1, err := os.Open(f)
	defer r1.Close()
	if err != nil {
		return
	}
	r := bufio.NewReader(r1)
	return p.pluck(r)
}

// PluckWeb takes a URL as input
// and uses the specified parameters and generates
// a map (p.result) with the finished results
func (p *plucker) PluckURL(url string) (err error) {
	resp, err := http.Get(url)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	r := bufio.NewReader(resp.Body)
	return p.pluck(r)
}

func (p *plucker) pluck(r *bufio.Reader) (err error) {
	for {
		curByte, errRead := r.ReadByte()
		allLimitsReached := true
		for i := range p.pluckers {
			if len(p.pluckers[i].captured) == p.pluckers[i].config.Limit {
				continue
			} else {
				allLimitsReached = false
			}
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
						p.pluckers[i].captured = append(p.pluckers[i].captured, tempByte)
						// reset
						p.pluckers[i].numActivated = 0
						p.pluckers[i].deactiveI = 0
						p.pluckers[i].captureI = 0
					}
				} else {
					p.pluckers[i].activeI = 0
				}
			}

		}

		if errRead == io.EOF || allLimitsReached {
			break
		}

		// TODO: Also break if everything has reached its limit
	}

	// Generate result
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

	return
}

// Returns the result, formatted as JSON
func (p *plucker) ResultJSON() string {
	resultJson, err := json.MarshalIndent(p.result, "", "    ")
	if err != nil {
		log.Error(errors.Wrap(err, "result marshalling failed"))
	}
	return string(resultJson)
}

// Returns the result
func (p *plucker) Result() map[string]interface{} {
	return p.result
}
