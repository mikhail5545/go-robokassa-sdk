/*
 * Copyright (c) 2026. Mikhail Kulik.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package types

import "strconv"

// Quantity3 represents a number with up to 5 integer digits and up to 3 fractional digits
// Stored as int32 with 3 decimal places (multiply by 1000)
// Max value: 99999.999
type Quantity3 int32

// MarshalJSON marshals Quantity3 as numeric JSON with 3 digits after decimal point
func (q Quantity3) MarshalJSON() ([]byte, error) {
	v := float64(q) / 1000.0
	return []byte(strconv.FormatFloat(v, 'f', 3, 64)), nil
}

// Price8x2 represents a number with up to 8 integer digits and up to 2 fractional digits
// Stored as int64 with 2 decimal places (multiply by 100)
// Max value: 99999999.99
type Price8x2 int64

// MarshalJSON marshals Price8x2 as numeric JSON with 2 digits after decimal point
func (p Price8x2) MarshalJSON() ([]byte, error) {
	v := float64(p) / 100.0
	return []byte(strconv.FormatFloat(v, 'f', 2, 64)), nil
}

// Amount is an alias of Price8x2 used in split payments and top-level invoice sums.
type Amount = Price8x2
