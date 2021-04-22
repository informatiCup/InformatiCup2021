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
	err := RegisterAI("HeartAI", func() AI { return new(HeartAI) })
	if err != nil {
		panic(err)
	}
}

// HeartAIActions contains the actions needed to draw a heart onto the game board. The last action will do a crash.
var HeartAIActions = []string{ActionNOOP, ActionTurnLeft, ActionTurnRight, ActionTurnLeft, ActionTurnRight, ActionTurnLeft, ActionTurnRight, ActionNOOP, ActionTurnRight, ActionTurnLeft, ActionTurnRight, ActionTurnRight, ActionTurnLeft, ActionNOOP, ActionTurnLeft, ActionTurnRight, ActionTurnRight, ActionTurnLeft, ActionTurnRight, ActionNOOP, ActionTurnRight, ActionTurnLeft, ActionTurnRight, ActionTurnLeft, ActionTurnRight}

// HeartAI is an AI that draws a heart.
type HeartAI struct {
	l sync.Mutex

	i       chan string
	counter int
}

// GetChannel receives the answer channel.
func (h *HeartAI) GetChannel(c chan string) {
	h.l.Lock()
	defer h.l.Unlock()

	h.i = c
}

// GetState gets the game state and computes an answer.
func (h *HeartAI) GetState(g *Game) {
	h.l.Lock()
	defer h.l.Unlock()

	if h.i == nil {
		return
	}

	if g.Running {
		if g.Players[g.You].Active {
			if h.counter >= len(HeartAIActions) {
				h.i <- ActionNOOP
				return
			}
			h.i <- HeartAIActions[h.counter]
			h.counter++
		}
	}
}

// Name returns the name of the AI.
func (h *HeartAI) Name() string {
	return "HeartAI"
}
