// Copyright (C) 2024 stencil contributors
// Copyright (C) 2022-2023 Outreach Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//        http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package dotnotation implements a dotnotation (hello.world) for
// accessing fields within a map[string]any
package dotnotation

import (
	"fmt"
	"reflect"
	"strings"
)

// Get looks up an entry in data by parsing the "key" into deeply nested keys, traversing it by "dots" in the key name.
func Get(data any, key string) (any, error) {
	return get(data, key)
}

// getFieldOnMap returns a field on a given map
func getFieldOnMap(data any, key string) (any, error) {
	dataVal := reflect.ValueOf(data)
	dataTyp := dataVal.Type()
	if dataTyp.Kind() != reflect.Map {
		return nil, fmt.Errorf("data is not a map")
	}

	// iterate over the keys of the map
	// converting them to the type of the key, when we find the key
	// we return the value
	iter := dataVal.MapRange()
	for iter.Next() {
		k := iter.Key()
		v := iter.Value()

		if k.Kind() == reflect.Interface {
			// convert any keys into their actual
			// values
			k = reflect.ValueOf(k.Interface())
		}

		// quick hack to convert all types to string :(
		strK := fmt.Sprintf("%v", k.Interface())
		if strK == key {
			return v.Interface(), nil
		}
	}

	return nil, fmt.Errorf("key %q not found", key)
}

// get is a recursive function to get a field from a map[any]any
// this is done by splitting the key on "." and using the first part of the
// split, if there is anymore parts of the key then get() is called with
// the non processed part
func get(data any, key string) (any, error) {
	spl := strings.Split(key, ".")

	v, err := getFieldOnMap(data, spl[0])
	if err != nil {
		return nil, err
	}

	// check if we have more keys to iterate over
	if len(spl) > 1 {
		// pop the first key, and get the next value as the next key to
		// process
		nextKey := spl[1:][0]
		nextDataTyp := reflect.TypeOf(v)
		if nextDataTyp == nil || nextDataTyp.Kind() != reflect.Map {
			return nil, fmt.Errorf("key %q is not a map, got %v on %q", nextKey, nextDataTyp, reflect.TypeOf(data))
		}

		// pop the first key, and call get() again
		return get(v, strings.Join(spl[1:], "."))
	}

	// otherwise return the data
	return v, nil
}
