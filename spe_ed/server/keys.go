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
	"bufio"
	"os"
	"strings"
	"sync"
)

const (
	// KeyOK is the value representing that the key is ok.
	KeyOK = iota
	// KeyRateLimit is the value representing that the key is currently used.
	KeyRateLimit
	// KeyInvalid is the value representing that the key is invalid.
	KeyInvalid
)

// NumberAllowedGames contains the number of games a key can participate in.
const NumberAllowedGames = 1

var keymapLock sync.Mutex
var keymap = make(map[string]int)

// InitKeys initialises all API keys from a file.
// Not safe to be used in parallel with other key functions.
func InitKeys(filename string) {
	f, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	s := bufio.NewScanner(f)
	for s.Scan() {
		text := s.Text()
		if len(text) == 0 {
			continue
		}
		if strings.HasPrefix(text, "#") {
			continue
		}
		keymap[text] = NumberAllowedGames
	}
	err = s.Err()
	if err != nil {
		panic(err)
	}
}

// ClaimKey tries to claim an API key.
// If it returns KeyOK, the number of usage of that key is internally increased, for all other values nothing changes.
// Keys are loaded from "./keys"
func ClaimKey(key string) int {
	if key == "" {
		log.Println("keys:", "invalid", key)
		return KeyInvalid
	}

	keymapLock.Lock()
	defer keymapLock.Unlock()

	available, ok := keymap[key]
	if !ok {
		log.Println("keys:", "invalid", key)
		return KeyInvalid
	}
	if available == 0 {
		log.Println("keys:", "ratelimit", key)
		return KeyRateLimit
	}
	keymap[key] = available - 1
	log.Println("keys:", "ok", key)
	return KeyOK
}

// ReleaseKey releases a key thus making it claimable again.
// Each call releases it for exactly one claim.
func ReleaseKey(key string) {
	if key == "" {
		return
	}

	keymapLock.Lock()
	defer keymapLock.Unlock()

	available, ok := keymap[key]
	if !ok {
		return
	}
	keymap[key] = available + 1
}
