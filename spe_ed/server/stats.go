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
	"bytes"
	"fmt"
	"html/template"
	"io"
	"sync"
	"time"
)

// PlayerStats contains the statistics of a single player.
type PlayerStats struct {
	Key       string
	Pseudonym string
	Bot       bool
}

// GameStats contains the statistics of a game.
type GameStats struct {
	Key     string
	Start   time.Time
	Players map[int]PlayerStats
}

// SendStat is used to add new Games to the statistics.
// Will block before InitStats is called.
var SendStat chan<- GameStats

// DeleteStat is used to remove games from the statistics. Games are identified by their key.
// Will block before InitStats is called.
var DeleteStat chan<- string

// SendLobby is used to add new keys to the lobby statistics.
// Will block before InitStats is called.
var SendLobby chan<- string

// DeleteLobby is used to remove keys from the lobby statistics.
// Will block before InitStats is called.
var DeleteLobby chan<- string

// GetStatPage can be used to get a reader containing a HTML-page with the current stats.
// The provided channel must be non-blocking or will be ignored.
// Will block before InitStats is called.
var GetStatPage chan<- chan io.Reader

var statsTemplate *template.Template

type statsTemplateStruct struct {
	Time       time.Time
	GameStats  map[string]GameStats
	LobbyStats map[string]bool
	LobbyTime  time.Duration
}

var statsOnce sync.Once
var statsMap map[string]GameStats
var lobbyMap map[string]bool

// InitStats will initialise the statistics routines. Successive calls have no effect.
func InitStats() {
	statsOnce.Do(func() {
		statsTemplate = template.Must(template.New("stats").Parse(`
		<!DOCTYPE HTML>
		<html lang="en">
		<body>
			<p>Time: {{ .Time.UTC.Format "2006-01-02T15:04:05Z07:00" }}</p>
			<p>Lobby max. wait time: {{.LobbyTime}}
			<h1>Lobby</h1>
			{{ if .LobbyStats }}
			<ul>
			{{ range $key, $value := .LobbyStats }}
				<li>{{ $key }}</li>
			{{ end }}
			</ul>
			{{ else }}
			<p>Empty</p>
			{{ end }}
			<h1>Games</h1>
			{{ range $gameID, $game := .GameStats }}
				<h2>{{ $gameID }}</h2>
				<p>Start: {{$game.Start.UTC.Format "2006-01-02T15:04:05Z07:00"}}</p>
				<h3>Players</h3>
				<table>
					<tr>
						<th>ID</th>
						<th>Key</th>
						<th>Pseudonym</th>
						<th>Bot</th>
					<tr>
					{{ range $playerID, $player := $game.Players }}				
					<tr>
						<td>{{ $playerID }}</td>
						<td>{{ $player.Key }}</td>
						<td>{{ $player.Pseudonym }}</td>
						<td>{{ $player.Bot }}</td>
					</tr>
					{{ end }}
				</table>
			{{ end }}
		</body>
		</html>
	`))

		s := make(chan GameStats)
		d := make(chan string)
		g := make(chan chan io.Reader, 10)
		sl := make(chan string)
		dl := make(chan string)
		statsMap = make(map[string]GameStats)
		lobbyMap = make(map[string]bool)
		SendStat = s
		DeleteStat = d
		GetStatPage = g
		SendLobby = sl
		DeleteLobby = dl
		go workerStats(s, d, sl, dl, g)
	})
}

func workerStats(send <-chan GameStats, deleteStats <-chan string, sendLobby <-chan string, deleteLobby <-chan string, get <-chan chan io.Reader) {
	for {
		select {
		case gs := <-send:
			statsMap[gs.Key] = gs
		case k := <-deleteStats:
			delete(statsMap, k)
		case k := <-sendLobby:
			lobbyMap[k] = true
		case k := <-deleteLobby:
			delete(lobbyMap, k)
		case g := <-get:
			var buf bytes.Buffer
			err := statsTemplate.Execute(&buf, statsTemplateStruct{Time: time.Now(), GameStats: statsMap, LobbyStats: lobbyMap, LobbyTime: maxWaitTime})
			if err != nil {
				fmt.Println("error rendering stats:", err)
			}
			select {
			case g <- &buf:
				// Ok
			default:
				// Ignore
			}
		}
	}
}
