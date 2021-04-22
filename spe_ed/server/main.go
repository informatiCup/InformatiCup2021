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

// This package contains the server for the game "spe_ed".
//
// Prequisites
//
// - Install go
//
// Build
//
// `go build`
//
// Run
//
// `./server`
//
// Options
//
// see `./server -help`
//
// More Information
// See https://github.com/informatiCup/InformatiCup2021/
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	golog "log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"
)

var (
	log           *golog.Logger
	disableTime   bool
	serverAddress = "localhost:10101"
	statsEnabled  bool
	keyFile       = "./keys"
	pseudonymFile = "./pseudonyms"
)

func init() {
	// Random
	rand.Seed(time.Now().Unix())
}

func main() {
	flag.BoolVar(&disableLogging, "disableLogging", false, "Disables logging of games")
	wait := flag.String("wait", "5m", "Waiting time for new games. Must be at least 0s (0=instant start for debugging). Value must be parseable by time.Duration")
	flag.BoolVar(&disableTime, "disableTime", false, "Disables time endpoint")
	flag.StringVar(&serverAddress, "address", serverAddress, "Address of the server")
	flag.BoolVar(&statsEnabled, "stats", false, "Enables stats on /spe_ed_stats")
	flag.StringVar(&keyFile, "keyfile", keyFile, "Path to key file")
	flag.StringVar(&pseudonymFile, "pseudonymfile", pseudonymFile, "Path to pseudonym file. Will be created if non-existing")
	ais := flag.String("ais", "", fmt.Sprintf("Comma seperated list of ais which should be used. Must be at least %d", PlayersPerGame))
	listais := flag.Bool("listais", false, "Lists all ai names and exits")
	logfilename := flag.String("logfile", "", "If set, logging will be done to file instead of to stdout")
	flag.Parse()

	if *listais {
		fmt.Println(GetAINames())
		return
	}

	if *ais != "" {
		err := UpdateAIPool(strings.Split(*ais, ","))
		if err != nil {
			panic(err)
		}
	}

	{
		var err error
		maxWaitTime, err = time.ParseDuration(*wait)

		if err != nil {
			panic(err)
		}
		if maxWaitTime < 0 {
			panic("waiting time too small")
		}
	}

	if *logfilename == "" {
		log = golog.New(os.Stdout, "spe_ed server ", golog.LstdFlags)
	} else {
		f, err := os.OpenFile(*logfilename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			panic(err)
		}
		defer f.Close()
		log = golog.New(f, "", golog.LstdFlags)
	}

	InitPseudonyms(pseudonymFile)
	InitKeys(keyFile)

	http.HandleFunc("/spe_ed", endpoint)

	if statsEnabled {
		InitStats()
		http.HandleFunc("/spe_ed_stats", func(rw http.ResponseWriter, r *http.Request) {
			rw.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
			request := make(chan io.Reader, 1)
			GetStatPage <- request
			stats := <-request
			_, err := io.Copy(rw, stats)
			if err != nil {
				log.Println("error copying stats:", err)
			}
		})
	}

	if !disableTime {
		http.HandleFunc("/spe_ed_time", func(rw http.ResponseWriter, r *http.Request) {
			now := time.Now().UTC()
			b, err := json.Marshal(struct {
				Time         string `json:"time"`
				Milliseconds int    `json:"milliseconds"`
			}{Time: now.Format(time.RFC3339), Milliseconds: now.Nanosecond() * int(time.Nanosecond) / int(time.Millisecond)})
			if err != nil {
				log.Println("time:", err)
				rw.WriteHeader(http.StatusInternalServerError)
				rw.Write([]byte(err.Error()))
				return
			}
			rw.Header().Set("Content-Type", "application/json")
			rw.Header().Set("Access-Control-Allow-Origin", "*")
			rw.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
			rw.Write(b)
		})
	}
	log.Fatal(http.ListenAndServe(serverAddress, nil))
}
