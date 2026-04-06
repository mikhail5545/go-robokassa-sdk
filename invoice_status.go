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
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type InvoiceInformation struct {
	ID          string
	EncodedID   string
	InvID       string
	Status      InvoiceStatus
	InvoiceType InvoiceType
	OutSum      *float64
	Description string

	CreatedAt *time.Time
	PaidAt    *time.Time
	ExpiredAt *time.Time

	Raw map[string]any
}

type InvoiceInformationListResponse struct {
	CurrentPage int
	PageSize    int
	TotalCount  int
	TotalPages  int
	Invoices    []InvoiceInformation
	RawResponse RawResponse
}

func (c *Client) GetInvoiceInformationListTyped(ctx context.Context, req GetInvoiceInformationListRequest) (*InvoiceInformationListResponse, error) {
	raw, err := c.GetInvoiceInformationList(ctx, req)
	if err != nil {
		return nil, err
	}
	return ParseInvoiceInformationListResponse(raw)
}

func ParseInvoiceInformationListResponse(raw *RawResponse) (*InvoiceInformationListResponse, error) {
	if raw == nil {
		return nil, errors.New("raw response is nil")
	}

	root, err := rawResponseObject(raw)
	if err != nil {
		return nil, err
	}

	payload := root
	hasNestedPayload := false
	if nested := nestedDataMap(root); nested != nil {
		payload = nested
		hasNestedPayload = true
	}

	invoiceRows := invoiceRowsFrom(payload)
	if len(invoiceRows) == 0 && hasNestedPayload {
		invoiceRows = invoiceRowsFrom(root)
	}

	invoices := make([]InvoiceInformation, 0, len(invoiceRows))
	for _, row := range invoiceRows {
		rowMap, ok := row.(map[string]any)
		if !ok {
			continue
		}
		invoices = append(invoices, parseInvoiceInformation(rowMap))
	}

	response := &InvoiceInformationListResponse{
		CurrentPage: firstInt(payload, root, "CurrentPage", "currentPage", "Page", "page"),
		PageSize:    firstInt(payload, root, "PageSize", "pageSize", "PerPage", "perPage"),
		TotalCount:  firstInt(payload, root, "TotalCount", "totalCount", "Count", "count", "Total", "total"),
		TotalPages:  firstInt(payload, root, "TotalPages", "totalPages", "Pages", "pages"),
		Invoices:    invoices,
		RawResponse: *raw,
	}

	if response.TotalCount == 0 && len(invoices) > 0 {
		response.TotalCount = len(invoices)
	}

	return response, nil
}

func parseInvoiceInformation(in map[string]any) InvoiceInformation {
	info := InvoiceInformation{
		ID:          firstString(in, "ID", "Id", "id"),
		EncodedID:   firstString(in, "EncodedID", "EncodedId", "encodedID", "encodedId"),
		InvID:       firstString(in, "InvID", "InvId", "invID", "invId", "InvoiceID", "InvoiceId", "invoiceId"),
		Status:      normalizeInvoiceStatus(firstString(in, "InvoiceStatus", "invoiceStatus", "Status", "status", "State", "state")),
		InvoiceType: normalizeInvoiceType(firstString(in, "InvoiceType", "invoiceType", "Type", "type")),
		Description: firstString(in, "Description", "description", "Comment", "comment"),
		CreatedAt:   firstTime(in, "CreatedAt", "createdAt", "Date", "date", "CreationDate", "creationDate"),
		PaidAt:      firstTime(in, "PaidAt", "paidAt", "PaymentDate", "paymentDate"),
		ExpiredAt:   firstTime(in, "ExpiredAt", "expiredAt", "ExpirationDate", "expirationDate"),
		Raw:         in,
	}

	if sum, ok := firstFloat(in, "OutSum", "outSum", "Amount", "amount", "Sum", "sum"); ok {
		info.OutSum = &sum
	}

	return info
}

func normalizeInvoiceStatus(value string) InvoiceStatus {
	raw := strings.TrimSpace(value)
	normalized := strings.ToLower(strings.ReplaceAll(strings.ReplaceAll(raw, "-", ""), "_", ""))
	switch normalized {
	case "paid", "ok", "success":
		return InvoiceStatusPaid
	case "expired":
		return InvoiceStatusExpired
	case "notpaid", "unpaid", "new", "created":
		return InvoiceStatusNotPaid
	default:
		if raw == "" {
			return ""
		}
		return InvoiceStatus(raw)
	}
}

func normalizeInvoiceType(value string) InvoiceType {
	raw := strings.TrimSpace(value)
	normalized := strings.ToLower(strings.ReplaceAll(strings.ReplaceAll(raw, "-", ""), "_", ""))
	switch normalized {
	case "onetime":
		return InvoiceTypeOneTime
	case "reusable":
		return InvoiceTypeReusable
	default:
		if raw == "" {
			return ""
		}
		return InvoiceType(raw)
	}
}

func rawResponseObject(raw *RawResponse) (map[string]any, error) {
	if raw.Object != nil {
		return raw.Object, nil
	}
	if len(raw.RawJSON) == 0 {
		return nil, errors.New("response body is empty")
	}

	var object map[string]any
	if err := json.Unmarshal(raw.RawJSON, &object); err != nil {
		return nil, fmt.Errorf("response is not a JSON object: %w", err)
	}
	return object, nil
}

func nestedDataMap(object map[string]any) map[string]any {
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

func invoiceRowsFrom(object map[string]any) []any {
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

func firstString(object map[string]any, keys ...string) string {
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

func firstInt(primary, fallback map[string]any, keys ...string) int {
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

func firstFloat(object map[string]any, keys ...string) (float64, bool) {
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

func firstTime(object map[string]any, keys ...string) *time.Time {
	for _, key := range keys {
		value, ok := object[key]
		if !ok || value == nil {
			continue
		}
		if parsed, ok := parseTime(value); ok {
			return &parsed
		}
	}
	return nil
}

func parseTime(value any) (time.Time, bool) {
	switch typed := value.(type) {
	case time.Time:
		return typed.UTC(), true
	case string:
		s := strings.TrimSpace(typed)
		if s == "" {
			return time.Time{}, false
		}
		layouts := []string{
			time.RFC3339Nano,
			time.RFC3339,
			"2006-01-02T15:04:05",
			"2006-01-02T15:04",
			"2006-01-02 15:04:05",
			"2006-01-02 15:04",
			"2006-01-02",
		}
		for _, layout := range layouts {
			if parsed, err := time.Parse(layout, s); err == nil {
				return parsed.UTC(), true
			}
		}
	case float64:
		if typed > 1e12 {
			return time.UnixMilli(int64(typed)).UTC(), true
		}
		return time.Unix(int64(typed), 0).UTC(), true
	case int64:
		if typed > 1e12 {
			return time.UnixMilli(typed).UTC(), true
		}
		return time.Unix(typed, 0).UTC(), true
	case json.Number:
		if unix, err := typed.Int64(); err == nil {
			if unix > 1e12 {
				return time.UnixMilli(unix).UTC(), true
			}
			return time.Unix(unix, 0).UTC(), true
		}
	}
	return time.Time{}, false
}
