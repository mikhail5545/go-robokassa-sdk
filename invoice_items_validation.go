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
	"fmt"
	"unicode/utf8"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

func validateInvoiceItems(items []*InvoiceItem, fieldName string) error {
	return validation.Validate(items, validation.By(func(interface{}) error {
		if len(items) == 0 {
			return nil
		}
		if len(items) > 100 {
			return fmt.Errorf("%s cannot contain more than 100 items", fieldName)
		}
		for index, item := range items {
			if err := validateInvoiceItem(item, fieldName, index); err != nil {
				return err
			}
		}
		return nil
	}))
}

func validateInvoiceItem(item *InvoiceItem, fieldName string, index int) error {
	if item == nil {
		return fmt.Errorf("%s[%d] is nil", fieldName, index)
	}

	err := validation.ValidateStruct(item,
		validation.Field(&item.Name,
			requiredTrimmedStringRule(fmt.Sprintf("%s[%d].name is required", fieldName, index)),
			validation.By(func(interface{}) error {
				if utf8.RuneCountInString(item.Name) > 128 {
					return fmt.Errorf("%s[%d].name must not exceed 128 characters", fieldName, index)
				}
				return nil
			}),
		),
		validation.Field(&item.Quantity, validation.By(func(interface{}) error {
			if !item.Quantity.IsValid() || item.Quantity <= 0 {
				return fmt.Errorf("%s[%d].quantity must be within 0.001..99999.999", fieldName, index)
			}
			return nil
		})),
		validation.Field(&item.Cost, validation.By(func(interface{}) error {
			if !item.Cost.IsValid() || item.Cost <= 0 {
				return fmt.Errorf("%s[%d].cost must be within 0.01..99999999.99", fieldName, index)
			}
			return nil
		})),
		validation.Field(&item.Tax, validation.By(func(interface{}) error {
			if !isSupportedTaxRate(item.Tax) {
				return fmt.Errorf("%s[%d].tax has unsupported value %q", fieldName, index, item.Tax)
			}
			return nil
		})),
		validation.Field(&item.PaymentMethod, validation.By(func(interface{}) error {
			if item.PaymentMethod != nil && !isSupportedPaymentMethod(*item.PaymentMethod) {
				return fmt.Errorf("%s[%d].payment_method has unsupported value %q", fieldName, index, *item.PaymentMethod)
			}
			return nil
		})),
		validation.Field(&item.PaymentObject, validation.By(func(interface{}) error {
			if item.PaymentObject != nil && !isSupportedPaymentObject(*item.PaymentObject) {
				return fmt.Errorf("%s[%d].payment_object has unsupported value %q", fieldName, index, *item.PaymentObject)
			}
			return nil
		})),
	)
	return firstValidationError(err, "Name", "Quantity", "Cost", "Tax", "PaymentMethod", "PaymentObject")
}
