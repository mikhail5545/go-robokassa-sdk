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
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var decimalAmountRegex = regexp.MustCompile(`^\d+(\.\d{1,6})?$`)

func normalizeRequiredOutSum(outSum float64, outSumText string) (string, error) {
	if strings.TrimSpace(outSumText) != "" {
		normalized, err := normalizeDecimalStringAmount(outSumText)
		if err != nil {
			return "", err
		}
		parsed, err := strconv.ParseFloat(normalized, 64)
		if err != nil {
			return "", errors.New("out sum has invalid numeric value")
		}
		if parsed <= 0 {
			return "", errors.New("out sum must be greater than zero")
		}
		return normalized, nil
	}
	if outSum <= 0 {
		return "", errors.New("out sum must be greater than zero")
	}
	return formatOutSum(outSum), nil
}

func normalizeOptionalOutSum(outSum *float64, outSumText string) (string, error) {
	if strings.TrimSpace(outSumText) != "" {
		normalized, err := normalizeDecimalStringAmount(outSumText)
		if err != nil {
			return "", err
		}
		parsed, err := strconv.ParseFloat(normalized, 64)
		if err != nil {
			return "", errors.New("out sum has invalid numeric value")
		}
		if parsed <= 0 {
			return "", errors.New("out sum must be greater than zero")
		}
		return normalized, nil
	}
	if outSum == nil {
		return "", nil
	}
	if *outSum <= 0 {
		return "", errors.New("out sum must be greater than zero")
	}
	return formatOutSum(*outSum), nil
}

func normalizeDecimalStringAmount(raw string) (string, error) {
	amount := strings.TrimSpace(strings.ReplaceAll(raw, ",", "."))
	if amount == "" {
		return "", errors.New("out sum is required")
	}
	if !decimalAmountRegex.MatchString(amount) {
		return "", fmt.Errorf("out sum has invalid format: %q", raw)
	}
	if !strings.Contains(amount, ".") {
		amount += ".00"
	}
	return amount, nil
}
