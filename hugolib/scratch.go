// Copyright 2015 The Hugo Authors. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package hugolib

import (
	"github.com/spf13/hugo/helpers"
	"reflect"
	"sort"
	"sync"
)

// Scratch is a writable context used for stateful operations in Page/Node rendering.
type Scratch struct {
	values map[string]interface{}
	mu     sync.RWMutex
}

// For single values, Add will add (using the + operator) the addend to the existing addend (if found).
// Supports numeric values and strings.
//
// If the first add for a key is an array or slice, then the next value(s) will be appended.
func (c *Scratch) Add(key string, newAddend interface{}) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	var newVal interface{}
	existingAddend, found := c.values[key]
	if found {
		var err error

		addendV := reflect.ValueOf(existingAddend)

		if addendV.Kind() == reflect.Slice || addendV.Kind() == reflect.Array {
			nav := reflect.ValueOf(newAddend)
			if nav.Kind() == reflect.Slice || nav.Kind() == reflect.Array {
				newVal = reflect.AppendSlice(addendV, nav).Interface()
			} else {
				newVal = reflect.Append(addendV, nav).Interface()
			}
		} else {
			newVal, err = helpers.DoArithmetic(existingAddend, newAddend, '+')
			if err != nil {
				return "", err
			}
		}
	} else {
		newVal = newAddend
	}
	c.values[key] = newVal
	return "", nil // have to return something to make it work with the Go templates
}

// Set stores a value with the given key in the Node context.
// This value can later be retrieved with Get.
func (c *Scratch) Set(key string, value interface{}) string {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.values[key] = value
	return ""
}

// Get returns a value previously set by Add or Set
func (c *Scratch) Get(key string) interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.values[key]
}

// SetInMap stores a value to a map with the given key in the Node context.
// This map can later be retrieved with GetSortedMapValues.
func (c *Scratch) SetInMap(key string, mapKey string, value interface{}) string {
	c.mu.Lock()
	defer c.mu.Unlock()

	_, found := c.values[key]
	if !found {
		c.values[key] = make(map[string]interface{})
	}

	c.values[key].(map[string]interface{})[mapKey] = value
	return ""
}

// GetSortedMapValues returns a sorted map previously filled with SetInMap
func (c *Scratch) GetSortedMapValues(key string) interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.values[key] == nil {
		return nil
	}

	unsortedMap := c.values[key].(map[string]interface{})

	var keys []string
	for mapKey := range unsortedMap {
		keys = append(keys, mapKey)
	}

	sort.Strings(keys)

	sortedArray := make([]interface{}, len(unsortedMap))
	for i, mapKey := range keys {
		sortedArray[i] = unsortedMap[mapKey]
	}

	return sortedArray
}

func newScratch() *Scratch {
	return &Scratch{values: make(map[string]interface{})}
}