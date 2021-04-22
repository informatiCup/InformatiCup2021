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
	err := RegisterAI("LargestFreeAI", func() AI { return new(LargestFreeAI) })
	if err != nil {
		panic(err)
	}
}

// LargestFreeAI is an AI which navigates the player in the direction of the largest free area (calculated as a line from the current position).
type LargestFreeAI struct {
	l sync.Mutex

	i chan string
}

// GetChannel receives the answer channel.
func (lf *LargestFreeAI) GetChannel(c chan string) {
	lf.l.Lock()
	defer lf.l.Unlock()

	lf.i = c
}

// GetState gets the game state and computes an answer.
func (lf *LargestFreeAI) GetState(g *Game) {
	lf.l.Lock()
	defer lf.l.Unlock()

	if lf.i == nil {
		return
	}

	if g.Running && g.Players[g.You].Active {
		action := ActionNOOP
		free := 0

		// Fill potential dead zones
		for k := range g.Players {
			if k == g.You {
				continue
			}

			if !g.Players[k].Active {
				continue
			}

			for i := 1; i <= g.Players[k].Speed+1; i++ {
				x, y := g.Players[k].X+i, g.Players[k].Y
				if x < 0 || x >= g.Width || y < 0 || y >= g.Height {
					// invalid - do nothing
				} else {
					g.Cells[y][x] = -100
				}

				x, y = g.Players[k].X-i, g.Players[k].Y
				if x < 0 || x >= g.Width || y < 0 || y >= g.Height {
					// invalid - do nothing
				} else {
					g.Cells[y][x] = -100
				}

				x, y = g.Players[k].X, g.Players[k].Y+i
				if x < 0 || x >= g.Width || y < 0 || y >= g.Height {
					// invalid - do nothing
				} else {
					g.Cells[y][x] = -100
				}

				x, y = g.Players[k].X, g.Players[k].Y-i
				if x < 0 || x >= g.Width || y < 0 || y >= g.Height {
					// invalid - do nothing
				} else {
					g.Cells[y][x] = -100
				}
			}
		}

		// Test direction
		switch g.Players[g.You].Direction {
		case DirectionUp:
			// Straight
			found := lf.GetFree(func(x, y int) (int, int) { return x, y - 1 }, g)
			if found > free {
				free = found
				action = ActionNOOP
			}
			// Left
			found = lf.GetFree(func(x, y int) (int, int) { return x - 1, y }, g)
			if found > free {
				free = found
				action = ActionTurnLeft
			}
			// Right
			found = lf.GetFree(func(x, y int) (int, int) { return x + 1, y }, g)
			if found > free {
				free = found
				action = ActionTurnRight
			}
		case DirectionRight:
			// Straight - right
			found := lf.GetFree(func(x, y int) (int, int) { return x + 1, y }, g)
			if found > free {
				free = found
				action = ActionNOOP
			}
			// Left - up
			found = lf.GetFree(func(x, y int) (int, int) { return x, y - 1 }, g)
			if found > free {
				free = found
				action = ActionTurnLeft
			}
			// Right - down
			found = lf.GetFree(func(x, y int) (int, int) { return x, y + 1 }, g)
			if found > free {
				free = found
				action = ActionTurnRight
			}
		case DirectionLeft:
			// Straight - left
			found := lf.GetFree(func(x, y int) (int, int) { return x - 1, y }, g)
			if found > free {
				free = found
				action = ActionNOOP
			}
			// Left - down
			found = lf.GetFree(func(x, y int) (int, int) { return x, y + 1 }, g)
			if found > free {
				free = found
				action = ActionTurnLeft
			}
			// Right - up
			found = lf.GetFree(func(x, y int) (int, int) { return x, y - 1 }, g)
			if found > free {
				free = found
				action = ActionTurnRight
			}
		case DirectionDown:
			// Straight
			found := lf.GetFree(func(x, y int) (int, int) { return x, y + 1 }, g)
			if found > free {
				free = found
				action = ActionNOOP
			}
			// Left
			found = lf.GetFree(func(x, y int) (int, int) { return x + 1, y }, g)
			if found > free {
				free = found
				action = ActionTurnLeft
			}
			// Right
			found = lf.GetFree(func(x, y int) (int, int) { return x - 1, y }, g)
			if found > free {
				free = found
				action = ActionTurnRight
			}
		}

		// Send action
		lf.i <- action
	}
}

// Name returns the name of the AI.
func (lf *LargestFreeAI) Name() string {
	return "LargestFreeAI"
}

// GetFree returns the number of free cells in the direction given by dostep.
// It is not safe for concurrent usage on the same game.
func (lf *LargestFreeAI) GetFree(dostep func(x, y int) (int, int), g *Game) int {
	x, y := g.Players[g.You].X, g.Players[g.You].Y
	free := 0
	for {
		x, y = dostep(x, y)
		if x < 0 || x >= g.Width || y < 0 || y >= g.Height {
			break
		}
		if g.Cells[y][x] != 0 {
			break
		}
		free++
	}
	return free
}
