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
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestParseInvoiceInformationListResponse_FlexibleKeys(t *testing.T) {
	raw := &RawResponse{
		Object: map[string]any{
			"data": map[string]any{
				"currentPage": 2,
				"pageSize":    10,
				"totalCount":  1,
				"totalPages":  5,
				"items": []any{
					map[string]any{
						"id":          "internal-id",
						"invId":       "123",
						"status":      "Paid",
						"invoiceType": "onetime",
						"outSum":      "10.50",
						"description": "Invoice",
						"createdAt":   "2025-03-28T08:36:02.651371+00:00",
					},
				},
			},
		},
	}

	parsed, err := ParseInvoiceInformationListResponse(raw)
	if err != nil {
		t.Fatalf("parse invoice list response: %v", err)
	}
	if parsed.CurrentPage != 2 || parsed.PageSize != 10 || parsed.TotalCount != 1 || parsed.TotalPages != 5 {
		t.Fatalf("unexpected page metadata: %+v", parsed)
	}
	if len(parsed.Invoices) != 1 {
		t.Fatalf("unexpected invoice count: got=%d want=1", len(parsed.Invoices))
	}

	invoice := parsed.Invoices[0]
	if invoice.ID != "internal-id" || invoice.InvID != "123" {
		t.Fatalf("unexpected invoice identifiers: %+v", invoice)
	}
	if invoice.Status != InvoiceStatusPaid {
		t.Fatalf("unexpected status: got=%q want=%q", invoice.Status, InvoiceStatusPaid)
	}
	if invoice.InvoiceType != InvoiceTypeOneTime {
		t.Fatalf("unexpected invoice type: got=%q want=%q", invoice.InvoiceType, InvoiceTypeOneTime)
	}
	if invoice.OutSum == nil || *invoice.OutSum != 10.50 {
		t.Fatalf("unexpected out sum: %+v", invoice.OutSum)
	}
	if invoice.CreatedAt == nil {
		t.Fatal("expected createdAt to be parsed")
	}
}

func TestGetInvoiceInformationListTyped_UsesTypedParser(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("unexpected method: got=%s want=POST", r.Method)
		}
		if r.URL.Path != getInvoiceInformationListPath {
			t.Fatalf("unexpected path: got=%s want=%s", r.URL.Path, getInvoiceInformationListPath)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
  "CurrentPage": 1,
  "PageSize": 10,
  "TotalCount": 1,
  "Invoices": [
    {
      "Id": "id-1",
      "InvId": "42",
      "Status": "expired",
      "InvoiceType": "Reusable",
      "OutSum": 77.35
    }
  ]
}`))
	}))
	defer server.Close()

	client := mustClient(t, Config{
		MerchantLogin: "merchant",
		Password1:     "password1",
		BaseURL:       server.URL,
	})

	from := time.Now().UTC().Add(-time.Hour)
	to := time.Now().UTC()
	typed, err := client.GetInvoiceInformationListTyped(context.Background(), GetInvoiceInformationListRequest{
		CurrentPage:     1,
		PageSize:        10,
		InvoiceStatuses: []InvoiceStatus{InvoiceStatusPaid, InvoiceStatusExpired, InvoiceStatusNotPaid},
		DateFrom:        &from,
		DateTo:          &to,
		InvoiceTypes:    []InvoiceType{InvoiceTypeOneTime, InvoiceTypeReusable},
	})
	if err != nil {
		t.Fatalf("get typed invoice information list: %v", err)
	}
	if len(typed.Invoices) != 1 {
		t.Fatalf("unexpected invoice count: got=%d want=1", len(typed.Invoices))
	}
	if typed.Invoices[0].Status != InvoiceStatusExpired {
		t.Fatalf("unexpected status: got=%q want=%q", typed.Invoices[0].Status, InvoiceStatusExpired)
	}
	if typed.Invoices[0].InvoiceType != InvoiceTypeReusable {
		t.Fatalf("unexpected type: got=%q want=%q", typed.Invoices[0].InvoiceType, InvoiceTypeReusable)
	}
}
