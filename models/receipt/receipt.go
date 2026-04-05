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

package receipt

import (
	"github.com/mikhail5545/go-robokassa-sdk/models/items"
	"github.com/mikhail5545/go-robokassa-sdk/types"
)

type Receipt struct {
	// An array of receipt item data. At least one item must be provided, but single receipt can contain no more than 100 product items.
	Items []*items.ReceiptItem `json:"items,omitempty"`
	// This parameter is optional; if omitted, the value from your personal account is used. Pass only if you need
	// to differentiate tax systems for different products.
	Sno *types.TaxSystem `json:"sno,omitempty"`
}
