/*
Copyright (c) 2015-Present CloudFoundry.org Foundation, Inc. All Rights Reserved.

This project contains software that is Copyright (c) 2013-2015 Pivotal Software, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

This project may include a number of subcomponents with separate
copyright notices and license terms. Your use of these subcomponents
is subject to the terms and conditions of each subcomponent's license,
as noted in the LICENSE file.
*/

package strings

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/google/cel-go/common/types"
)

const (
	BYTE = 1 << (10 * iota)
	KILOBYTE
	MEGABYTE
	GIGABYTE
	TERABYTE
	PETABYTE
	EXABYTE
)

// ByteSize returns a human-readable byte string of the form 10M, 12.5K, and so forth.  The following units are available:
//
//	E: Exabyte
//	P: Petabyte
//	T: Terabyte
//	G: Gigabyte
//	M: Megabyte
//	K: Kilobyte
//	B: Byte
//
// The unit that results in the smallest number greater than or equal to 1 is always chosen.
// Input is the size in bytes.
func HumanBytes(size interface{}) string {
	unit := ""
	var bytes uint64
	var err error
	switch t := size.(type) {
	case types.Int:
		bytes = uint64(t)
	case uint:
		bytes = uint64(t)
	case uint64:
		bytes = t
	case int:
		bytes = uint64(t)
	case int64:
		bytes = uint64(t)
	case float64:
		bytes = uint64(t)
	case float32:
		bytes = uint64(t)
	case string:
		bytes, err = strconv.ParseUint(t, 10, 64)
		if err != nil {
			return "NaN"
		}
	default:
		return fmt.Sprintf("unknown type: %v = %t", size, size)
	}

	value := float64(bytes)

	switch {
	case bytes >= EXABYTE:
		unit = "E"
		value = value / EXABYTE
	case bytes >= PETABYTE:
		unit = "P"
		value = value / PETABYTE
	case bytes >= TERABYTE:
		unit = "T"
		value = value / TERABYTE
	case bytes >= GIGABYTE:
		unit = "G"
		value = value / GIGABYTE
	case bytes >= MEGABYTE:
		unit = "M"
		value = value / MEGABYTE
	case bytes >= KILOBYTE:
		unit = "K"
		value = value / KILOBYTE
	case bytes >= BYTE:
		unit = "B"
	case bytes == 0:
		return "0B"
	}

	result := strconv.FormatFloat(value, 'f', 1, 64)
	result = strings.TrimSuffix(result, ".0")
	return result + unit
}
