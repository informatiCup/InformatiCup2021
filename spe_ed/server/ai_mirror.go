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
	"math/rand"
	"sync"
)

func init() {
	err := RegisterAI("MirrorAI", func() AI { return new(MirrorAI) })
	if err != nil {
		panic(err)
	}
}

// MirrorAI is an AI which mirrors the action of an other random (active) player.
type MirrorAI struct {
	l sync.Mutex

	target          int
	targetSpeed     int
	targetDirection string

	i chan string
}

// GetChannel receives the answer channel.
func (m *MirrorAI) GetChannel(c chan string) {
	m.l.Lock()
	defer m.l.Unlock()

	m.i = c
}

// GetState gets the game state and computes an answer.
func (m *MirrorAI) GetState(g *Game) {
	m.l.Lock()
	defer m.l.Unlock()

	if m.i == nil {
		return
	}

	if g.Running && g.Players[g.You].Active {
		// Is target still active?
		if m.target != 0 && !g.Players[m.target].Active {
			m.target = 0
		}

		// Do we need new target?
		if m.target == 0 {
			// Find target
			player := make([]int, 0, len(g.Players))
			for k := range g.Players {
				if k != g.You && g.Players[k].Active {
					player = append(player, k)
				}
			}
			m.target = player[rand.Intn(len(player))]

			// Save data
			m.targetDirection = g.Players[m.target].Direction
			m.targetSpeed = g.Players[m.target].Speed

			// Send action
			m.i <- ActionNOOP
			return
		}

		// Mirror target

		action := ActionNOOP
		// Find action taken
		switch {
		// Faster
		case g.Players[m.target].Speed > m.targetSpeed:
			m.targetSpeed = g.Players[m.target].Speed
			if g.Players[g.You].Speed < 10 {
				action = ActionFaster
			}

		// Slower
		case g.Players[m.target].Speed < m.targetSpeed:
			m.targetSpeed = g.Players[m.target].Speed
			if g.Players[g.You].Speed > 1 {
				action = ActionSlower
			}

		// Turning
		case g.Players[m.target].Direction != m.targetDirection:
			switch m.targetDirection {
			case DirectionUp:
				if g.Players[m.target].Direction == DirectionLeft {
					action = ActionTurnLeft
				} else if g.Players[m.target].Direction == DirectionRight {
					action = ActionTurnRight
				}
			case DirectionDown:
				if g.Players[m.target].Direction == DirectionRight {
					action = ActionTurnLeft
				} else if g.Players[m.target].Direction == DirectionLeft {
					action = ActionTurnRight
				}
			case DirectionLeft:
				if g.Players[m.target].Direction == DirectionDown {
					action = ActionTurnLeft
				} else if g.Players[m.target].Direction == DirectionUp {
					action = ActionTurnRight
				}
			case DirectionRight:
				if g.Players[m.target].Direction == DirectionUp {
					action = ActionTurnLeft
				} else if g.Players[m.target].Direction == DirectionDown {
					action = ActionTurnRight
				}
			}
			m.targetDirection = g.Players[m.target].Direction
		}

		// Send action
		m.i <- action
	}
}

// Name returns the name of the AI.
func (m *MirrorAI) Name() string {
	return "MirrorAI"
}
