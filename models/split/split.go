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

package split

import (
	"github.com/mikhail5545/go-robokassa-sdk/models/receipt"
	"github.com/mikhail5545/go-robokassa-sdk/types"
)

type Merchant struct {
	ID        string           `json:"id"`        // Participating store identifier. Required.
	InvoiceID string           `json:"InvoiceId"` // A unique store account number. An integer from 1 to int64 max (9223372036854775807). If the value is equal to 0 or the field is missing, it is assigned automatically. Optional field
	Amount    types.Amount     `json:"amount"`    // The amount specific store will receive. Specified in rubles; a value is allowed `0.00` if no accrual is required. Required field.
	Receipt   *receipt.Receipt `json:"receipt"`   // An object with the item number for generating a receipt. Optional parameter
}
