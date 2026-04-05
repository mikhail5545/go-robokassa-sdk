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
	robokassa "github.com/mikhail5545/go-robokassa-sdk"
	"github.com/mikhail5545/go-robokassa-sdk/models/items"
)

// ToCreateRefundRequest converts legacy RefundRequest into modern root request type.
func (r RefundRequest) ToCreateRefundRequest() robokassa.CreateRefundRequest {
	return robokassa.CreateRefundRequest{
		OpKey:        r.OpKey,
		RefundSum:    r.RefundSum,
		InvoiceItems: cloneInvoiceItems(r.InvoiceItems),
	}
}

// RefundRequestFromCreateRefundRequest converts modern root request type into legacy shape.
func RefundRequestFromCreateRefundRequest(req robokassa.CreateRefundRequest) RefundRequest {
	return RefundRequest{
		OpKey:        req.OpKey,
		RefundSum:    req.RefundSum,
		InvoiceItems: cloneInvoiceItems(req.InvoiceItems),
	}
}

func cloneInvoiceItems(in []*items.InvoiceItem) []*items.InvoiceItem {
	if len(in) == 0 {
		return nil
	}
	out := make([]*items.InvoiceItem, len(in))
	copy(out, in)
	return out
}
