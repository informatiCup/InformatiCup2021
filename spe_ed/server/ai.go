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
	"errors"
	"fmt"
	"math/rand"
	"sort"
	"sync"
)

var aiMap = make(map[string]AINewFunc)
var aiLock sync.RWMutex
var aiArray = []func() (AI, string){
	func() (AI, string) { return new(EndRound), GlobalPseudonym.Get("AI-EndRound-1") },
	func() (AI, string) { return new(HeartAI), GlobalPseudonym.Get("AI-HeartAI-3") },
	func() (AI, string) { return new(ChristmasAI), GlobalPseudonym.Get("AI-ChristmasAI-24") },
	func() (AI, string) { return new(StupidAI), GlobalPseudonym.Get("AI-StupidAI-1") },
	func() (AI, string) { return new(StupidAI), GlobalPseudonym.Get("AI-StupidAI-2") },
	func() (AI, string) { return new(StupidAI), GlobalPseudonym.Get("AI-StupidAI-3") },
	func() (AI, string) { return new(StupidAI), GlobalPseudonym.Get("AI-StupidAI-4") },
	func() (AI, string) { return new(StupidAI), GlobalPseudonym.Get("AI-StupidAI-5") },
	func() (AI, string) { return new(SnailAI), GlobalPseudonym.Get("AI-SnailAI-1") },
	func() (AI, string) { return new(SnailAI), GlobalPseudonym.Get("AI-SnailAI-2") },
	func() (AI, string) { return new(SuperSnailAI), GlobalPseudonym.Get("AI-SuperSnailAI-1") },
	func() (AI, string) { return new(SuperSnailAI), GlobalPseudonym.Get("AI-SuperSnailAI-2") },
	func() (AI, string) { return new(SuperSnailAI), GlobalPseudonym.Get("AI-SuperSnailAI-3") },
	func() (AI, string) { return new(JumpingSnailAI), GlobalPseudonym.Get("AI-JumpingSnailAI-1") },
	func() (AI, string) { return new(JumpingSnailAI), GlobalPseudonym.Get("AI-JumpingSnailAI-2") },
	func() (AI, string) { return new(JumpingSnailAI), GlobalPseudonym.Get("AI-JumpingSnailAI-3") },
	func() (AI, string) { return new(JumpingSnailAI), GlobalPseudonym.Get("AI-JumpingSnailAI-4") },
	func() (AI, string) { return new(JumpingSnailAI), GlobalPseudonym.Get("AI-JumpingSnailAI-5") },
	func() (AI, string) { return new(LargestFreeAI), GlobalPseudonym.Get("AI-LargestFreeAI-1") },
	func() (AI, string) { return new(LargestFreeAI), GlobalPseudonym.Get("AI-LargestFreeAI-2") },
	func() (AI, string) { return new(LargestFreeAI), GlobalPseudonym.Get("AI-LargestFreeAI-3") },
	func() (AI, string) { return new(LargestFreeAI), GlobalPseudonym.Get("AI-LargestFreeAI-4") },
	func() (AI, string) { return new(LargestFreeAI), GlobalPseudonym.Get("AI-LargestFreeAI-5") },
	func() (AI, string) {
		return new(JumpingLargestFreeAI), GlobalPseudonym.Get("AI-JumpingLargestFreeAI-1")
	},
	func() (AI, string) {
		return new(JumpingLargestFreeAI), GlobalPseudonym.Get("AI-JumpingLargestFreeAI-2")
	},
	func() (AI, string) {
		return new(JumpingLargestFreeAI), GlobalPseudonym.Get("AI-JumpingLargestFreeAI-3")
	},
	func() (AI, string) {
		return new(JumpingLargestFreeAI), GlobalPseudonym.Get("AI-JumpingLargestFreeAI-4")
	},
	func() (AI, string) {
		return new(JumpingLargestFreeAI), GlobalPseudonym.Get("AI-JumpingLargestFreeAI-5")
	},
	func() (AI, string) { return new(RandomAI), GlobalPseudonym.Get("AI-RandomAI-1") },
	func() (AI, string) { return new(RandomAI), GlobalPseudonym.Get("AI-RandomAI-2") },
	func() (AI, string) { return new(BadRandomAI), GlobalPseudonym.Get("AI-BadRandomAI-1") },
	func() (AI, string) { return new(RandomAISlow), GlobalPseudonym.Get("AI-RandomAISlow-1") },
	func() (AI, string) { return new(RandomAISlow), GlobalPseudonym.Get("AI-RandomAISlow-2") },
	func() (AI, string) { return new(SuperRandomAI), GlobalPseudonym.Get("AI-SuperRandomAI-1") },
	func() (AI, string) { return new(SuperRandomAI), GlobalPseudonym.Get("AI-SuperRandomAI-2") },
	func() (AI, string) { return new(SuperRandomAI), GlobalPseudonym.Get("AI-SuperRandomAI-3") },
	func() (AI, string) { return new(SuperRandomAI), GlobalPseudonym.Get("AI-SuperRandomAI-4") },
	func() (AI, string) { return new(SuperRandomAI), GlobalPseudonym.Get("AI-SuperRandomAI-5") },
	func() (AI, string) { return new(MirrorAI), GlobalPseudonym.Get("AI-MirrorAI-1") },
	func() (AI, string) { return new(MirrorAI), GlobalPseudonym.Get("AI-MirrorAI-2") },
	func() (AI, string) { return new(MirrorAI), GlobalPseudonym.Get("AI-MirrorAI-3") },
	func() (AI, string) { return new(MirrorAI), GlobalPseudonym.Get("AI-MirrorAI-4") },
	func() (AI, string) { return new(JumpAI), GlobalPseudonym.Get("AI-JumpAI-1") },
	func() (AI, string) { return new(JumpAI), GlobalPseudonym.Get("AI-JumpAI-2") },
	func() (AI, string) { return new(JumpAI), GlobalPseudonym.Get("AI-JumpAI-3") },
	func() (AI, string) { return new(JumpAI), GlobalPseudonym.Get("AI-JumpAI-4") },
	func() (AI, string) { return new(JumpAI), GlobalPseudonym.Get("AI-JumpAI-5") },
	func() (AI, string) { return new(MetaAI), GlobalPseudonym.Get("AI-MetaAI-1") },
	func() (AI, string) { return new(MetaAI), GlobalPseudonym.Get("AI-MetaAI-2") },
	func() (AI, string) { return new(MetaAI), GlobalPseudonym.Get("AI-MetaAI-3") },
	func() (AI, string) { return new(MetaAI), GlobalPseudonym.Get("AI-MetaAI-4") },
	func() (AI, string) { return new(MetaAI), GlobalPseudonym.Get("AI-MetaAI-5") },
}

// The AI interface provides the interface for different AIs.
//
// All channel actions must be non-blocking.
//
// In GetState, AIs can only access public fields (and change them) plus stepCounter, privat fields are set to zero.
// Modification of the game is allowed. The Caller has to make sure that modifications to the provided game can be done without side effects (e.g. by using Game.PublicCopy )
type AI interface {
	GetChannel(c chan string)
	GetState(g *Game)
	Name() string
}

// NewAI provides a new AI with given Name.
type NewAI struct {
	AI  AI
	API string
}

// AINewFunc must return a new AI
type AINewFunc func() AI

// RegisterAI registers an AI. Name must be unique or else an error will occur.
func RegisterAI(name string, makeai AINewFunc) error {
	aiLock.Lock()
	defer aiLock.Unlock()
	if _, ok := aiMap[name]; ok {
		return fmt.Errorf("ai name %s already registered", name)
	}
	if makeai == nil {
		return errors.New("AINewFunc must not be nil")
	}
	aiMap[name] = makeai
	return nil
}

// UpdateAIPool sets the ai pool to the names provided.
// Names can be provided multiple times, which results in multiple additions (and thus higher chance of drawing) to the pool.
// Must have at least PlayersPerGame names.
// If it returns an error, the pool is guaranteed to be unchanged.
func UpdateAIPool(ais []string) error {
	aiLock.Lock()
	defer aiLock.Unlock()

	if len(ais) < PlayersPerGame {
		return fmt.Errorf("at least %d ai names must be included (including repetitions", PlayersPerGame)
	}

	counter := make(map[string]int, len(aiMap))

	a := make([]func() (AI, string), 0, len(ais))

	for i := range ais {
		f, ok := aiMap[ais[i]]
		if !ok {
			return fmt.Errorf("ai name %s not known", ais[i])
		}
		counter[ais[i]]++
		p := fmt.Sprintf("AI-%s-%d", ais[i], counter[ais[i]])
		a = append(a, func() (AI, string) { return f(), GlobalPseudonym.Get(p) })
	}
	aiArray = a
	return nil
}

// GetAINames returns a list of all known ais in alphabetical order.
func GetAINames() []string {
	aiLock.RLock()
	defer aiLock.RUnlock()
	s := make([]string, 0, len(aiMap))
	for k := range aiMap {
		s = append(s, k)
	}
	sort.Strings(s)
	return s
}

// GetAI returns a slice of AIs of specified number out of the current rotation.
// Function might panic if number is to large. This should only occur if the number is larger than 6.
func GetAI(num int) []NewAI {
	aiLock.RLock()
	defer aiLock.RUnlock()

	if num > len(aiArray) {
		panic("Not enough AI")
	}

	r := make([]NewAI, num)
	selectArray := make([]int, len(aiArray))
	for i := range selectArray {
		selectArray[i] = i
	}
	rand.Shuffle(len(selectArray), func(i, j int) { selectArray[i], selectArray[j] = selectArray[j], selectArray[i] })

	for i := range r {
		r[i].AI, r[i].API = aiArray[selectArray[i]]()
	}

	return r
}
