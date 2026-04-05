package items

import "github.com/mikhail5545/go-robokassa-sdk/types"

type InvoiceItem struct {
	Name             string               `json:"name"`                        // Item name (max 128 characters)
	Quantity         types.Quantity3      `json:"quantity"`                    // Quantity or weight of th item.
	Cost             types.Price8x2       `json:"cost"`                        // Price per unit
	Tax              types.TaxRate        `json:"tax"`                         // TaxRate
	PaymentMethod    *types.PaymentMethod `json:"payment_method,omitempty"`    // Payment method (optional)
	PaymentObject    *types.PaymentObject `json:"payment_object,omitempty"`    // Payment object (optional)
	NomenclatureCode *string              `json:"nomenclature_code,omitempty"` // Product marking code (required for marked products)
}

type ReceiptItem struct {
	// Product name (up to 128 characters). Required
	Name string `json:"name"`
	// Quantity or weight of the item
	Quantity types.Quantity3 `json:"quantity"`
	// Required if Cost is not specified. The total price of the item, including discounts and bonuses.
	Sum types.Price8x2 `json:"sum,omitempty"`
	// Optional unit price. Can be passed in place of Sum; the total is calculated as Cost * Quantity
	Cost *types.Price8x2 `json:"cost,omitempty"`
	// Required. Tax rate in the cash register for the item.
	Tax types.TaxRate `json:"tax_rate"`
	// Optional types.PaymentMethod. Calculation method indicator. If not provided, the default value from your personal account is used.
	PaymentMethod *types.PaymentMethod `json:"payment_method,omitempty"`
	// Payment object (optional, uses default from merchant panel if not provided)
	PaymentObject *types.PaymentObject `json:"payment_object,omitempty"`
	// Mandatory for labeled products. The labeling code is from product packaging
	NomenclatureCode *string `json:"nomenclature_code,omitempty"`
}
