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
	"strings"
	"testing"
)

func TestValidateInvoiceItems_TableDriven(t *testing.T) {
	tests := []struct {
		name      string
		buildItem func() *InvoiceItem
		wantErr   string
	}{
		{
			name: "valid item",
			buildItem: func() *InvoiceItem {
				return validInvoiceItemForValidationTests()
			},
		},
		{
			name: "nil item",
			buildItem: func() *InvoiceItem {
				return nil
			},
			wantErr: "invoice items[0] is nil",
		},
		{
			name: "name is required",
			buildItem: func() *InvoiceItem {
				item := validInvoiceItemForValidationTests()
				item.Name = "  "
				return item
			},
			wantErr: "invoice items[0].name is required",
		},
		{
			name: "quantity must be positive and in range",
			buildItem: func() *InvoiceItem {
				item := validInvoiceItemForValidationTests()
				item.Quantity = 0
				return item
			},
			wantErr: "invoice items[0].quantity must be within 0.001..99999.999",
		},
		{
			name: "cost must be positive and in range",
			buildItem: func() *InvoiceItem {
				item := validInvoiceItemForValidationTests()
				item.Cost = 0
				return item
			},
			wantErr: "invoice items[0].cost must be within 0.01..99999999.99",
		},
		{
			name: "tax must be supported",
			buildItem: func() *InvoiceItem {
				item := validInvoiceItemForValidationTests()
				item.Tax = TaxRate("unsupported")
				return item
			},
			wantErr: `invoice items[0].tax has unsupported value "unsupported"`,
		},
		{
			name: "payment method must be supported",
			buildItem: func() *InvoiceItem {
				item := validInvoiceItemForValidationTests()
				method := PaymentMethod("unsupported")
				item.PaymentMethod = &method
				return item
			},
			wantErr: `invoice items[0].payment_method has unsupported value "unsupported"`,
		},
		{
			name: "payment object must be supported",
			buildItem: func() *InvoiceItem {
				item := validInvoiceItemForValidationTests()
				object := PaymentObject("unsupported")
				item.PaymentObject = &object
				return item
			},
			wantErr: `invoice items[0].payment_object has unsupported value "unsupported"`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := validateInvoiceItems([]*InvoiceItem{tc.buildItem()}, "invoice items")
			if tc.wantErr == "" {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("expected error containing %q", tc.wantErr)
			}
			if !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("unexpected error: got=%q wantContains=%q", err.Error(), tc.wantErr)
			}
		})
	}
}

func TestCreateInvoiceAndRefundValidation_UsesInvoiceItemRules(t *testing.T) {
	invalidTax := TaxRate("bad-tax")

	createInvoiceErr := CreateInvoiceRequest{
		MerchantLogin: "merchant",
		InvoiceType:   InvoiceTypeOneTime,
		OutSum:        10,
		InvoiceItems: []*InvoiceItem{
			{
				Name:     "Subscription",
				Quantity: 1000,
				Cost:     100,
				Tax:      invalidTax,
			},
		},
	}.validate()
	if createInvoiceErr == nil || !strings.Contains(createInvoiceErr.Error(), "invoice items[0].tax has unsupported value") {
		t.Fatalf("expected invoice-item tax validation error, got: %v", createInvoiceErr)
	}

	createRefundErr := CreateRefundRequest{
		OpKey: "op-key-1",
		InvoiceItems: []*InvoiceItem{
			{
				Name:     "Subscription",
				Quantity: 1000,
				Cost:     100,
				Tax:      invalidTax,
			},
		},
	}.validate()
	if createRefundErr == nil || !strings.Contains(createRefundErr.Error(), "invoice items[0].tax has unsupported value") {
		t.Fatalf("expected invoice-item tax validation error, got: %v", createRefundErr)
	}
}

func TestCreateInvoiceRequestValidate_PreservesValidationPriority(t *testing.T) {
	err := (CreateInvoiceRequest{
		InvoiceType: InvoiceType("bad-type"),
		OutSum:      0,
	}).validate()
	if err == nil {
		t.Fatal("expected validation error")
	}
	if err.Error() != "merchant login is required" {
		t.Fatalf("unexpected first validation error: got=%q want=%q", err.Error(), "merchant login is required")
	}
}

func validInvoiceItemForValidationTests() *InvoiceItem {
	return &InvoiceItem{
		Name:     "Subscription",
		Quantity: 1000,
		Cost:     100,
		Tax:      TaxRateVat20,
	}
}
