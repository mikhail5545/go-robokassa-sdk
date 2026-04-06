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

package normalization

import (
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"

	ozzo "github.com/go-ozzo/ozzo-validation/v4"

	internalvalidation "github.com/mikhail5545/go-robokassa-sdk/internal/validation"
)

var (
	decimalAmountRegex = regexp.MustCompile(`^\d+(\.\d{1,6})?$`)
	shpKeySuffixRegex  = regexp.MustCompile(`^[A-Za-z0-9_]+$`)
)

func NormalizeRequiredOutSum(outSum float64, outSumText string) (string, error) {
	if strings.TrimSpace(outSumText) != "" {
		normalized, err := NormalizeDecimalStringAmount(outSumText)
		if err != nil {
			return "", err
		}
		if err := ozzo.Validate(
			normalized,
			internalvalidation.PositiveDecimalStringRule("out sum has invalid numeric value", "out sum must be greater than zero"),
		); err != nil {
			return "", err
		}
		return normalized, nil
	}

	if err := ozzo.Validate(outSum, internalvalidation.GreaterThanZeroFloatRule("out sum must be greater than zero")); err != nil {
		return "", err
	}
	return FormatOutSum(outSum), nil
}

func NormalizeOptionalOutSum(outSum *float64, outSumText string) (string, error) {
	if strings.TrimSpace(outSumText) != "" {
		normalized, err := NormalizeDecimalStringAmount(outSumText)
		if err != nil {
			return "", err
		}
		if err := ozzo.Validate(
			normalized,
			internalvalidation.PositiveDecimalStringRule("out sum has invalid numeric value", "out sum must be greater than zero"),
		); err != nil {
			return "", err
		}
		return normalized, nil
	}

	if outSum == nil {
		return "", nil
	}
	if err := ozzo.Validate(*outSum, internalvalidation.GreaterThanZeroFloatRule("out sum must be greater than zero")); err != nil {
		return "", err
	}
	return FormatOutSum(*outSum), nil
}

func NormalizeDecimalStringAmount(raw string) (string, error) {
	amount := strings.TrimSpace(strings.ReplaceAll(raw, ",", "."))
	if err := ozzo.Validate(
		amount,
		ozzo.Required.Error("out sum is required"),
		ozzo.Match(decimalAmountRegex).Error(fmt.Sprintf("out sum has invalid format: %q", raw)),
	); err != nil {
		return "", err
	}
	if !strings.Contains(amount, ".") {
		amount += ".00"
	}
	return amount, nil
}

func NormalizeHTTPMethod(method *string) (string, error) {
	raw := TrimPtr(method)
	if raw == "" {
		return "", nil
	}
	upper := strings.ToUpper(raw)
	if err := ozzo.Validate(upper, ozzo.In("GET", "POST").Error("method must be GET or POST")); err != nil {
		return "", err
	}
	return upper, nil
}

func NormalizeShpParams(in map[string]string) (map[string]string, error) {
	if len(in) == 0 {
		return nil, nil
	}

	out := make(map[string]string, len(in))
	for key, value := range in {
		key = strings.TrimSpace(key)
		if err := ozzo.Validate(key, ozzo.Required.Error("shp key cannot be empty")); err != nil {
			return nil, err
		}

		keySuffix := key
		if strings.HasPrefix(strings.ToLower(key), "shp_") {
			keySuffix = key[4:]
		}
		keySuffix = strings.TrimSpace(keySuffix)
		if err := ozzo.Validate(
			keySuffix,
			ozzo.Required.Error(fmt.Sprintf("invalid shp key %q", key)),
			ozzo.Match(shpKeySuffixRegex).Error(
				fmt.Sprintf("invalid shp key %q: only latin letters, numbers and underscore are allowed", key),
			),
		); err != nil {
			return nil, err
		}

		canonicalKey := "Shp_" + keySuffix
		if _, exists := out[canonicalKey]; exists {
			return nil, fmt.Errorf("duplicate shp key after normalization: %q", canonicalKey)
		}
		out[canonicalKey] = value
	}

	return out, nil
}

func NormalizeInvoiceID(invoiceID int64) (string, error) {
	if err := ozzo.Validate(invoiceID, internalvalidation.GreaterThanZeroInt64Rule("invoice id must be greater than zero")); err != nil {
		return "", err
	}
	return strconv.FormatInt(invoiceID, 10), nil
}

func BuildURLWithValues(rawURL string, values url.Values) (string, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("parse endpoint url: %w", err)
	}
	u.RawQuery = values.Encode()
	return u.String(), nil
}

func OutAmountToRaw(value float64) (int64, error) {
	normalized := FormatOutSum(value)
	parts := strings.SplitN(normalized, ".", 2)
	if len(parts) != 2 {
		return 0, errors.New("out amount has invalid format")
	}
	whole, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return 0, errors.New("out amount has invalid integer part")
	}
	frac, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return 0, errors.New("out amount has invalid fractional part")
	}
	return whole*100 + frac, nil
}

func FormatOutSum(value float64) string {
	return strconv.FormatFloat(value, 'f', 2, 64)
}

func TrimPtr(v *string) string {
	if v == nil {
		return ""
	}
	return strings.TrimSpace(*v)
}

func SortedKeys[K ~string, V any](m map[K]V) []K {
	if len(m) == 0 {
		return nil
	}
	keys := make([]K, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})
	return keys
}
