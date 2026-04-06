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

package robokassa

import (
	"fmt"
	"strconv"
)

const (
	quantity3Scale = int64(1_000)
	price8x2Scale  = int64(100)

	// Quantity3 supports up to 99999.999.
	maxQuantity3Raw = int64(99_999_999)
	// Price8x2 supports up to 99999999.99.
	maxPrice8x2Raw = int64(9_999_999_999)
)

// Quantity3 represents a decimal value with up to 5 integer digits and 3 fractional digits.
// The value is stored as fixed-point integer with scale 1000.
type Quantity3 int32

// Price8x2 represents a decimal value with up to 8 integer digits and 2 fractional digits.
// The value is stored as fixed-point integer with scale 100.
type Price8x2 int64

// Amount is an alias of Price8x2 used in split payments and top-level invoice sums.
type Amount = Price8x2

func (q Quantity3) MarshalJSON() ([]byte, error) {
	raw := int64(q)
	if raw < 0 {
		return nil, fmt.Errorf("quantity must be non-negative: %d", q)
	}
	if raw > maxQuantity3Raw {
		return nil, fmt.Errorf("quantity exceeds max 99999.999: %d", q)
	}
	return []byte(formatScaled(raw, quantity3Scale, 3)), nil
}

func (q Quantity3) IsValid() bool {
	raw := int64(q)
	return raw >= 0 && raw <= maxQuantity3Raw
}

func (q Quantity3) String() string {
	return formatScaled(int64(q), quantity3Scale, 3)
}

func (p Price8x2) MarshalJSON() ([]byte, error) {
	raw := int64(p)
	if raw < 0 {
		return nil, fmt.Errorf("price must be non-negative: %d", p)
	}
	if raw > maxPrice8x2Raw {
		return nil, fmt.Errorf("price exceeds max 99999999.99: %d", p)
	}
	return []byte(formatScaled(raw, price8x2Scale, 2)), nil
}

func (p Price8x2) IsValid() bool {
	raw := int64(p)
	return raw >= 0 && raw <= maxPrice8x2Raw
}

func (p Price8x2) String() string {
	return formatScaled(int64(p), price8x2Scale, 2)
}

func formatScaled(raw, scale int64, fracDigits int) string {
	sign := ""
	if raw < 0 {
		sign = "-"
		raw = -raw
	}
	intPart := raw / scale
	fracPart := raw % scale
	return sign + strconv.FormatInt(intPart, 10) + "." + leftPadInt(fracPart, fracDigits)
}

func leftPadInt(value int64, width int) string {
	s := strconv.FormatInt(value, 10)
	if len(s) >= width {
		return s
	}
	padding := make([]byte, width-len(s))
	for i := range padding {
		padding[i] = '0'
	}
	return string(padding) + s
}
