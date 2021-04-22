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
	err := RegisterAI("MetaAI", func() AI { return new(MetaAI) })
	if err != nil {
		panic(err)
	}
}

// MetaAI is an AI which uses various different AIs.
type MetaAI struct {
	l sync.Mutex

	i  chan string
	ai AI
}

// GetChannel receives the answer channel.
func (meta *MetaAI) GetChannel(c chan string) {
	meta.l.Lock()
	defer meta.l.Unlock()

	meta.i = c
}

// GetState gets the game state and computes an answer.
func (meta *MetaAI) GetState(g *Game) {
	meta.l.Lock()
	defer meta.l.Unlock()

	if meta.i == nil {
		return
	}

	if g.Running {
		if rand.Float64() < 0.1 {
			meta.ai = nil
		}

		if meta.ai == nil {
			if g.Players[g.You].Speed > 1 {

				// Most AIs don't work with high speeds...
				var dostep func(x, y int) (int, int)
				possible := true
				x, y := g.Players[g.You].X, g.Players[g.You].Y

				switch g.Players[g.You].Direction {
				case DirectionUp:
					dostep = func(x, y int) (int, int) { return x, y - 1 }
				case DirectionDown:
					dostep = func(x, y int) (int, int) { return x, y + 1 }
				case DirectionLeft:
					dostep = func(x, y int) (int, int) { return x - 1, y }
				case DirectionRight:
					dostep = func(x, y int) (int, int) { return x + 1, y }
				}

				for i := 0; i < g.Players[g.You].Speed-1; i++ {
					x, y = dostep(x, y)
					if x < 0 || x >= g.Width || y < 0 || y >= g.Height || g.Cells[y][x] != 0 {
						possible = false
						break
					}
				}

				if possible {
					select {
					case meta.i <- ActionSlower:
					default:
					}
					return
				}

				// Check turn_left
				possible = true
				x, y = g.Players[g.You].X, g.Players[g.You].Y

				switch g.Players[g.You].Direction {
				case DirectionLeft:
					dostep = func(x, y int) (int, int) { return x, y - 1 }
				case DirectionRight:
					dostep = func(x, y int) (int, int) { return x, y + 1 }
				case DirectionDown:
					dostep = func(x, y int) (int, int) { return x - 1, y }
				case DirectionUp:
					dostep = func(x, y int) (int, int) { return x + 1, y }
				}

				for i := 0; i < g.Players[g.You].Speed; i++ {
					x, y = dostep(x, y)
					if x < 0 || x >= g.Width || y < 0 || y >= g.Height || g.Cells[y][x] != 0 {
						possible = false
						break
					}
				}

				if possible {
					select {
					case meta.i <- ActionTurnLeft:
					default:
					}
					return
				}

				// Check turn_right
				possible = true
				x, y = g.Players[g.You].X, g.Players[g.You].Y

				switch g.Players[g.You].Direction {
				case DirectionRight:
					dostep = func(x, y int) (int, int) { return x, y - 1 }
				case DirectionLeft:
					dostep = func(x, y int) (int, int) { return x, y + 1 }
				case DirectionUp:
					dostep = func(x, y int) (int, int) { return x - 1, y }
				case DirectionDown:
					dostep = func(x, y int) (int, int) { return x + 1, y }
				}

				for i := 0; i < g.Players[g.You].Speed; i++ {
					x, y = dostep(x, y)
					if x < 0 || x >= g.Width || y < 0 || y >= g.Height || g.Cells[y][x] != 0 {
						possible = false
						break
					}
				}

				if possible {
					select {
					case meta.i <- ActionTurnRight:
					default:
					}
					return
				}

				// nothing to survive, so....
				select {
				case meta.i <- ActionSlower:
				default:
				}
				return
			}
			ais := []AI{&LargestFreeAI{}, &SuperSnailAI{}, &StupidAI{}, &RandomAISlow{}}
			meta.ai = ais[rand.Intn(len(ais))]
			meta.ai.GetChannel(meta.i)
		}

		meta.ai.GetState(g)
	}
}

// Name returns the name of the AI.
func (meta *MetaAI) Name() string {
	return "MetaAI"
}
