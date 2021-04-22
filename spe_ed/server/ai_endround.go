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

import "sync"

func init() {
	err := RegisterAI("EndRound", func() AI { return new(EndRound) })
	if err != nil {
		panic(err)
	}
}

// EndRound is a simple AI that always returns the "change_nothing" action.
type EndRound struct {
	l sync.Mutex

	i chan string
}

// GetChannel receives the answer channel.
func (er *EndRound) GetChannel(c chan string) {
	er.l.Lock()
	defer er.l.Unlock()

	er.i = c
}

// GetState gets the game state and computes an answer.
func (er *EndRound) GetState(g *Game) {
	er.l.Lock()
	defer er.l.Unlock()

	if er.i == nil {
		return
	}

	if g.Running {
		if g.Players[g.You].Active {
			select {
			case er.i <- ActionNOOP:
			default:
			}
		}
	}
}

// Name returns the name of the AI.
func (er *EndRound) Name() string {
	return "EndRound"
}
