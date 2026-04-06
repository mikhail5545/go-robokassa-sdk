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
	"regexp"
	"strings"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

var decimalAmountRegex = regexp.MustCompile(`^\d+(\.\d{1,6})?$`)

func normalizeRequiredOutSum(outSum float64, outSumText string) (string, error) {
	if strings.TrimSpace(outSumText) != "" {
		normalized, err := normalizeDecimalStringAmount(outSumText)
		if err != nil {
			return "", err
		}
		if err := validation.Validate(
			normalized,
			positiveDecimalStringRule("out sum has invalid numeric value", "out sum must be greater than zero"),
		); err != nil {
			return "", err
		}
		return normalized, nil
	}
	if err := validation.Validate(outSum, greaterThanZeroFloatRule("out sum must be greater than zero")); err != nil {
		return "", err
	}
	return formatOutSum(outSum), nil
}

func normalizeOptionalOutSum(outSum *float64, outSumText string) (string, error) {
	if strings.TrimSpace(outSumText) != "" {
		normalized, err := normalizeDecimalStringAmount(outSumText)
		if err != nil {
			return "", err
		}
		if err := validation.Validate(
			normalized,
			positiveDecimalStringRule("out sum has invalid numeric value", "out sum must be greater than zero"),
		); err != nil {
			return "", err
		}
		return normalized, nil
	}
	if outSum == nil {
		return "", nil
	}
	if err := validation.Validate(*outSum, greaterThanZeroFloatRule("out sum must be greater than zero")); err != nil {
		return "", err
	}
	return formatOutSum(*outSum), nil
}

func normalizeDecimalStringAmount(raw string) (string, error) {
	amount := strings.TrimSpace(strings.ReplaceAll(raw, ",", "."))
	if err := validation.Validate(
		amount,
		validation.Required.Error("out sum is required"),
		validation.Match(decimalAmountRegex).Error(fmt.Sprintf("out sum has invalid format: %q", raw)),
	); err != nil {
		return "", err
	}
	if !strings.Contains(amount, ".") {
		amount += ".00"
	}
	return amount, nil
}
