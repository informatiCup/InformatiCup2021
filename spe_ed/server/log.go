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
	"encoding/base32"
	"encoding/json"
	"errors"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pierrec/lz4/v4"
)

const logPath = "./log/"

var disableLogging = false

func init() {
	err := os.MkdirAll(logPath, os.ModePerm)
	if err != nil {
		panic(err)
	}
}

type playerLog struct {
	APIKey    string
	Pseudonym string
	AI        string
}

// Logger allows for games to be saved to a lz4-compressed file, thus making them analyseable later.
type Logger struct {
	file   *os.File
	w      *lz4.Writer
	data   chan []byte
	closed bool
}

// GetLogger returns a logger and a game name to log a game to. All actions are saved in a lz4-compressed file.
// If disableLogging is set to true, logger is nil.
func GetLogger() (*Logger, string, error) {
	prefix := make([]byte, 10)
	rand.Read(prefix)
	id := base32.StdEncoding.EncodeToString(prefix)
	if disableLogging {
		return nil, id, errors.New("logging disabled")
	}
	filename := strings.Join([]string{time.Now().Format(time.RFC3339), "-", id, ".json.lz4"}, "")
	filename = filepath.Join(logPath, filename)

	var err error
	l := new(Logger)

	l.file, err = os.Create(filename)
	if err != nil {
		return nil, id, err
	}
	l.w = lz4.NewWriter(l.file)
	l.data = make(chan []byte, 10)
	go l.worker()
	return l, id, nil
}

// LogPlayer writes the player map to the log file.
// Should be called once in the beginning.
func (l *Logger) LogPlayer(p map[int]*Player) {
	metadata := make(map[int]playerLog)

	for k, v := range p {
		pl := playerLog{
			APIKey:    v.api,
			Pseudonym: v.realName,
			AI:        "",
		}

		if v.underlyingAI != nil {
			pl.AI = v.underlyingAI.Name()
		}
		metadata[k] = pl
	}

	b, err := json.Marshal(metadata)
	if err != nil {
		log.Println("logger:", err)
		return
	}
	l.data <- b
}

// LogState writes the game state to the log file.
func (l *Logger) LogState(g *Game) {
	if l.closed {
		log.Println("logger: writing while closed")
		return
	}

	b, err := json.Marshal(g)
	if err != nil {
		log.Println("logger:", err)
	}
	l.data <- b
}

// Close closes the log file.
func (l *Logger) Close() {
	if !l.closed {
		close(l.data)
		l.closed = true
	}
}

func (l *Logger) worker() {
	for b := range l.data {
		if l.w == nil {
			// Invalid logger - ignore
			continue
		}

		_, err := l.w.Write(b)
		if err != nil {
			log.Println("logger:", err)
		}
		_, err = l.w.Write([]byte("\n"))
		if err != nil {
			log.Println("logger:", err)
		}
	}
	if l.w != nil {
		l.w.Close()
	}
	l.w = nil
	if l.file != nil {
		l.file.Close()
	}
	l.file = nil
}
