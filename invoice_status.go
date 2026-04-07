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
	"errors"
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

// InvoiceInformationListResponse is typed view of GetInvoiceInformationList payload.
type InvoiceInformationListResponse struct {
	CurrentPage int
	PageSize    int
	TotalCount  int
	TotalPages  int
	Invoices    []InvoiceInformation
	RawResponse RawResponse
}

// GetInvoiceInformationListTyped requests invoice list and parses it to typed response.
func (c *Client) GetInvoiceInformationListTyped(ctx context.Context, req GetInvoiceInformationListRequest) (*InvoiceInformationListResponse, error) {
	raw, err := c.GetInvoiceInformationList(ctx, req)
	if err != nil {
		return nil, err
	}
	return ParseInvoiceInformationListResponse(raw)
}

// ParseInvoiceInformationListResponse parses raw invoice list payload into typed fields.
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
