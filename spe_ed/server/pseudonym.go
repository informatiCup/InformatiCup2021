// SPDX-License-Identifier: Apache-2.0
// Copyright 2020,2021 Philipp Naumann, Marcus Soll
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	  http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"encoding/json"
	"io/ioutil"
	"math/rand"
	"os"
	"strings"
	"sync"
	"time"
)

const (
	// PseudonymUpdateInterval is the interval at which pseudonyms will be updated.
	PseudonymUpdateInterval = 336 * time.Hour // 14 days
)

// Pseudonym represents the current pseudonyms used by the server.
// The pseudonyms will regularily be saved to the disc to "./pseudonyms".
// The pseudonyms will automatically be updated.
type Pseudonym struct {
	LastUpdated time.Time
	Dict        map[string]string
	l           sync.Mutex
}

// GlobalPseudonym is the global instance of Pseudonym
var GlobalPseudonym Pseudonym

// InitPseudonyms initialises the global instance of Pseudonym.
// Not safe to be used in parallel with other pseudonym functions.
func InitPseudonyms(filename string) {
	// Load Pseudonyms
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		GlobalPseudonym.Dict = make(map[string]string)
		GlobalPseudonym.LastUpdated = time.Now()
		go GlobalPseudonym.worker()
	} else {
		b, err := ioutil.ReadFile(filename)
		if err != nil {
			panic(err)
		}
		err = json.Unmarshal(b, &GlobalPseudonym)
		if err != nil {
			panic(err)
		}
		go GlobalPseudonym.worker()
	}
}

// NewPseudonym returns a new random pseudonym.
func NewPseudonym() string {
	words := make([]string, 3)
	for i := range words {
		words[i] = wordlist[rand.Intn(len(wordlist))]
	}
	return strings.Join(words, "-")
}

func (p *Pseudonym) worker() {
	for {
		p.l.Lock()
		if time.Now().Sub(p.LastUpdated) > PseudonymUpdateInterval {
			for k := range p.Dict {
				p.Dict[k] = NewPseudonym()
			}
			log.Println("pseudonym:", "updated pseudonyms")
			p.LastUpdated = time.Now()
		}
		b, err := json.Marshal(p)
		if err != nil {
			log.Println("pseudonym:", "marshal", err)
		} else {
			err = ioutil.WriteFile("./pseudonyms", b, os.ModePerm)
			if err != nil {
				log.Println("pseudonym:", "writing file", err)
			}
		}
		log.Println("pseudonym:", "saved state")
		p.l.Unlock()
		time.Sleep(10 * time.Minute)
	}
}

// Get returns the current pseudonym for a given string (e.g. player API key or AI name).
// It will create a new one if the string has no previous pseudonym associated with it.
func (p *Pseudonym) Get(API string) string {
	p.l.Lock()
	defer p.l.Unlock()
	v, ok := p.Dict[API]
	if !ok {
		v = NewPseudonym()
		p.Dict[API] = v
	}
	return v
}

var wordlist = []string{
	"Abbild",
	"Abbrecher",
	"Abendhimmel",
	"Anschluss",
	"Befehlswort",
	"Beginn",
	"Buchsbaum",
	"Butterteig",
	"Crêpe",
	"Cookie",
	"Computer",
	"Christenheit",
	"Dampflokomotive",
	"Dekor",
	"Diskette",
	"Düsenantrieb",
	"Evakuierung",
	"Exil",
	"Extrem",
	"Effekt",
	"Fragenkreis",
	"Frachtbrief",
	"Forschung",
	"Flügel",
	"Gehirn",
	"Geldfälschung",
	"Gemeinschaftsangelegenheiten",
	"Germanistenkongress",
	"Herz",
	"Heizöl",
	"Hochwasser",
	"Holunderbaum",
	"Ingenieur",
	"Informationsverarbeitungsprozess",
	"Implementierungsdetail",
	"Isomorphismus",
	"Jazz",
	"Jugend",
	"Jahrgang",
	"Jachtklub",
	"Kernfusion",
	"Keyboard",
	"Kabel",
	"Klagebegründung",
	"Lilie",
	"Lied",
	"Lichtgeschwindigkeit",
	"Legitimität",
	"Mittelmeer",
	"Modell",
	"Motiv",
	"Musik",
	"Nummerncode",
	"Nussknacker",
	"Norm",
	"Neuwagen",
	"Orientteppich",
	"Orgelklang",
	"Omnibus",
	"Optik",
	"Paket",
	"Phosphor",
	"Panik",
	"Papierfetzen",
	"Qualifikation",
	"Quellcode",
	"Quadratwurzel",
	"Qualm",
	"Radar",
	"Raddampfer",
	"Reisegeschwindigkeit",
	"Rekordflug",
	"Sicherung",
	"Schienennetz",
	"Schadensersatzforderung",
	"Schwefel",
	"Transaktion",
	"Taxi",
	"Testfall",
	"Troll",
	"Urteil",
	"Upload",
	"Untersuchungsausschuss",
	"Ultraschall",
	"Veranlassung",
	"Vergnügung",
	"Verkehr",
	"Vorteil",
	"Wachhund",
	"Wagenabteil",
	"Walnuss",
	"Wasserfahrzeug",
	"XML",
	"Xylofon",
	"X",
	"Xenolith",
	"Yeti",
	"Yoga",
	"Ysop",
	"Yack",
	"Zahlenspiel",
	"Zeilennummer",
	"Zeppelin",
	"Zentimeter",
}
