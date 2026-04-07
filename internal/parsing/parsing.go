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

package parsing

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
)

var supportedTimeLayouts = []string{
	time.RFC3339Nano,
	time.RFC3339,
	"2006-01-02T15:04:05",
	"2006-01-02T15:04",
	"2006-01-02 15:04:05",
	"2006-01-02 15:04",
	"2006-01-02",
}

func ParseRawResponse(body []byte) (json.RawMessage, string, map[string]any) {
	raw := make([]byte, len(body))
	copy(raw, body)

	trimmed := bytes.TrimSpace(body)
	if len(trimmed) == 0 {
		return raw, "", nil
	}

	var asString string
	if err := json.Unmarshal(trimmed, &asString); err == nil {
		return raw, asString, nil
	}

	var asObject map[string]any
	if err := json.Unmarshal(trimmed, &asObject); err == nil {
		return raw, "", asObject
	}

	return raw, string(trimmed), nil
}

func RawResponseObject(object map[string]any, rawJSON json.RawMessage) (map[string]any, error) {
	if object != nil {
		return object, nil
	}
	if len(rawJSON) == 0 {
		return nil, errors.New("response body is empty")
	}

	var parsed map[string]any
	if err := json.Unmarshal(rawJSON, &parsed); err != nil {
		return nil, fmt.Errorf("response is not a JSON object: %w", err)
	}
	return parsed, nil
}

func NestedDataMap(object map[string]any) map[string]any {
	if object == nil {
		return nil
	}
	if m, ok := object["Data"].(map[string]any); ok {
		return m
	}
	if m, ok := object["data"].(map[string]any); ok {
		return m
	}
	return nil
}

func InvoiceRowsFrom(object map[string]any) []any {
	if object == nil {
		return nil
	}
	for _, key := range []string{"Invoices", "invoices", "Items", "items", "InvoiceInformationList", "invoiceInformationList"} {
		if rows, ok := object[key].([]any); ok {
			return rows
		}
	}
	if rows, ok := object["Data"].([]any); ok {
		return rows
	}
	if rows, ok := object["data"].([]any); ok {
		return rows
	}
	return nil
}

func FindPaymentURL(obj map[string]any) string {
	keys := []string{
		"InvoiceUrl", "invoiceUrl",
		"PaymentUrl", "paymentUrl",
		"ShortUrl", "shortUrl",
		"Url", "url",
		"Link", "link",
	}
	for _, key := range keys {
		if value, ok := obj[key]; ok {
			if s, ok := value.(string); ok && strings.TrimSpace(s) != "" {
				return s
			}
		}
	}

	if data, ok := obj["Data"]; ok {
		if nested, ok := data.(map[string]any); ok {
			return FindPaymentURL(nested)
		}
	}
	if data, ok := obj["data"]; ok {
		if nested, ok := data.(map[string]any); ok {
			return FindPaymentURL(nested)
		}
	}

	return ""
}

func FirstBool(object map[string]any, keys ...string) bool {
	for _, key := range keys {
		value, ok := object[key]
		if !ok || value == nil {
			continue
		}
		switch typed := value.(type) {
		case bool:
			return typed
		case string:
			return strings.EqualFold(strings.TrimSpace(typed), "true")
		}
	}
	return false
}

func FirstString(object map[string]any, keys ...string) string {
	for _, key := range keys {
		value, ok := object[key]
		if !ok || value == nil {
			continue
		}
		switch typed := value.(type) {
		case string:
			if strings.TrimSpace(typed) != "" {
				return typed
			}
		case float64:
			return strconv.FormatFloat(typed, 'f', -1, 64)
		case int:
			return strconv.Itoa(typed)
		case int64:
			return strconv.FormatInt(typed, 10)
		case json.Number:
			return typed.String()
		}
	}
	return ""
}

func FirstInt(primary, fallback map[string]any, keys ...string) int {
	if value, ok := findInt(primary, keys...); ok {
		return value
	}
	if fallback != nil {
		if value, ok := findInt(fallback, keys...); ok {
			return value
		}
	}
	return 0
}

func FirstFloat(object map[string]any, keys ...string) (float64, bool) {
	for _, key := range keys {
		value, ok := object[key]
		if !ok || value == nil {
			continue
		}
		switch typed := value.(type) {
		case float64:
			return typed, true
		case int:
			return float64(typed), true
		case int64:
			return float64(typed), true
		case string:
			parsed, err := strconv.ParseFloat(strings.ReplaceAll(strings.TrimSpace(typed), ",", "."), 64)
			if err == nil {
				return parsed, true
			}
		case json.Number:
			parsed, err := typed.Float64()
			if err == nil {
				return parsed, true
			}
		}
	}
	return 0, false
}

func FirstTime(object map[string]any, keys ...string) *time.Time {
	for _, key := range keys {
		value, ok := object[key]
		if !ok || value == nil {
			continue
		}
		if parsed, ok := ParseTime(value); ok {
			return &parsed
		}
	}
	return nil
}

func ParseTime(value any) (time.Time, bool) {
	switch typed := value.(type) {
	case time.Time:
		return typed.UTC(), true
	case string:
		return parseTimeFromString(typed)
	case float64:
		return parseTimeFromUnix(int64(typed))
	case int64:
		return parseTimeFromUnix(typed)
	case json.Number:
		return parseTimeFromJSONNumber(typed)
	}
	return time.Time{}, false
}

func parseTimeFromString(value string) (time.Time, bool) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return time.Time{}, false
	}
	for _, layout := range supportedTimeLayouts {
		if parsed, err := time.Parse(layout, trimmed); err == nil {
			return parsed.UTC(), true
		}
	}
	return time.Time{}, false
}

func parseTimeFromUnix(unix int64) (time.Time, bool) {
	if unix > 1e12 {
		return time.UnixMilli(unix).UTC(), true
	}
	return time.Unix(unix, 0).UTC(), true
}

func parseTimeFromJSONNumber(value json.Number) (time.Time, bool) {
	unix, err := value.Int64()
	if err != nil {
		return time.Time{}, false
	}
	return parseTimeFromUnix(unix)
}

func CallbackValue(values url.Values, keys ...string) string {
	for _, key := range keys {
		if value := strings.TrimSpace(values.Get(key)); value != "" {
			return value
		}
	}
	return ""
}

func findInt(object map[string]any, keys ...string) (int, bool) {
	for _, key := range keys {
		value, ok := object[key]
		if !ok || value == nil {
			continue
		}
		switch typed := value.(type) {
		case int:
			return typed, true
		case int64:
			return int(typed), true
		case float64:
			return int(typed), true
		case string:
			parsed, err := strconv.Atoi(strings.TrimSpace(typed))
			if err == nil {
				return parsed, true
			}
		case json.Number:
			parsed, err := typed.Int64()
			if err == nil {
				return int(parsed), true
			}
		}
	}
	return 0, false
}
