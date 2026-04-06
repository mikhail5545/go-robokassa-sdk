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

type InvoiceItem struct {
	Name             string         `json:"name"`                        // Item name (max 128 characters)
	Quantity         Quantity3      `json:"quantity"`                    // Quantity or weight of the item
	Cost             Price8x2       `json:"cost"`                        // Price per unit
	Tax              TaxRate        `json:"tax"`                         // Tax rate
	PaymentMethod    *PaymentMethod `json:"payment_method,omitempty"`    // Payment method (optional)
	PaymentObject    *PaymentObject `json:"payment_object,omitempty"`    // Payment object (optional)
	NomenclatureCode *string        `json:"nomenclature_code,omitempty"` // Product marking code (required for marked products)
}

type ReceiptItem struct {
	// Product name (up to 128 characters). Required.
	Name string `json:"name"`
	// Quantity or weight of the item.
	Quantity Quantity3 `json:"quantity"`
	// Required if Cost is not specified. The total price of the item, including discounts and bonuses.
	Sum Price8x2 `json:"sum,omitempty"`
	// Optional unit price. Can be passed in place of Sum; the total is calculated as Cost * Quantity.
	Cost *Price8x2 `json:"cost,omitempty"`
	// Required. Tax rate in the cash register for the item.
	Tax TaxRate `json:"tax"`
	// Optional PaymentMethod. Calculation method indicator. If not provided, the default value from merchant panel is used.
	PaymentMethod *PaymentMethod `json:"payment_method,omitempty"`
	// Payment object (optional, uses default from merchant panel if not provided).
	PaymentObject *PaymentObject `json:"payment_object,omitempty"`
	// Mandatory for labeled products. The labeling code is from product packaging.
	NomenclatureCode *string `json:"nomenclature_code,omitempty"`
}

type Receipt struct {
	// An array of receipt item data. At least one item must be provided, but single receipt can contain no more than 100 product items.
	Items []*ReceiptItem `json:"items,omitempty"`
	// This parameter is optional; if omitted, the value from your personal account is used.
	Sno *TaxSystem `json:"sno,omitempty"`
}

type SplitMerchant struct {
	ID        string   `json:"id"`        // Participating store identifier. Required.
	InvoiceID string   `json:"InvoiceId"` // A unique store account number.
	Amount    Amount   `json:"amount"`    // The amount specific store will receive.
	Receipt   *Receipt `json:"receipt"`   // Optional receipt data for this merchant.
}
