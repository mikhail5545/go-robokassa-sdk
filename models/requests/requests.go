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

package requests

import (
	"github.com/mikhail5545/go-robokassa-sdk/models/items"
	"github.com/mikhail5545/go-robokassa-sdk/models/receipt"
	"github.com/mikhail5545/go-robokassa-sdk/types"
)

// PaymentRequest is a legacy request shape.
// Use ToInitPaymentRequest to work with robokassa.InitPaymentRequest API.
type PaymentRequest struct {
	OutSum         float64           `json:"out_sum"`                   // Payment amount
	Description    string            `json:"description"`               // Payment description
	InvID          *string           `json:"inv_id,omitempty"`          // Invoice ID (optional)
	Email          *string           `json:"email,omitempty"`           // Customer email address (optional)
	Culture        *types.Culture    `json:"culture,omitempty"`         // Language (ru, en)
	Encoding       *string           `json:"encoding,omitempty"`        // Encoding (default: utf-8)
	IsTest         *int              `json:"is_test,omitempty"`         // Test mode flag (1 for test)
	ExpirationDate *string           `json:"expiration_date,omitempty"` // Payment expiration date (optional)
	UserParameters map[string]string `json:"user_parameters,omitempty"` // Additional user parameters (optional)
	Receipt        *receipt.Receipt  `json:"receipt,omitempty"`         // Receipt data for fiscalization
}

// RefundRequest is a legacy request shape.
// Use ToCreateRefundRequest to work with robokassa.CreateRefundRequest API.
type RefundRequest struct {
	OpKey        string               `json:"op_key"`                  // Operation key
	RefundSum    *float64             `json:"refund_sum,omitempty"`    // Partial refund amount (omit for full refund)
	InvoiceItems []*items.InvoiceItem `json:"invoice_items,omitempty"` // Invoice items to refund (optional)
}
