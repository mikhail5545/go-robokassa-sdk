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
	"encoding/json"
	"fmt"

	validation "github.com/go-ozzo/ozzo-validation/v4"

	internalvalidation "github.com/mikhail5545/go-robokassa-sdk/internal/validation"
)

func marshalReceipt(r *Receipt) (string, error) {
	if r == nil {
		return "", nil
	}
	if err := validation.Validate(
		len(r.Items),
		validation.Min(1).Error("invalid receipt: must contain at least one item"),
		validation.Max(100).Error("invalid receipt: cannot contain more than 100 items"),
	); err != nil {
		return "", err
	}
	for i, item := range r.Items {
		if err := validateReceiptItem(i, item); err != nil {
			return "", err
		}
	}
	b, err := json.Marshal(r)
	if err != nil {
		return "", fmt.Errorf("marshal receipt: %w", err)
	}
	return string(b), nil
}

func validateReceiptItem(index int, item *ReceiptItem) error {
	if err := validation.Validate(
		item,
		validation.Required.Error(fmt.Sprintf("invalid receipt: item at index %d is nil", index)),
	); err != nil {
		return err
	}

	if err := validation.Validate(
		item.Name,
		requiredTrimmedStringRule(fmt.Sprintf("invalid receipt item at index %d: name is required", index)),
		maxRuneCountRule(128, fmt.Sprintf("invalid receipt item at index %d: name must not exceed 128 characters", index)),
	); err != nil {
		return err
	}

	if err := validation.Validate(
		item.Quantity,
		validation.By(func(value interface{}) error {
			quantity, _ := value.(Quantity3)
			if !quantity.IsValid() {
				return fmt.Errorf("invalid receipt item at index %d: quantity must be within 0..99999.999", index)
			}
			return nil
		}),
		validation.By(func(value interface{}) error {
			quantity, _ := value.(Quantity3)
			if quantity <= 0 {
				return fmt.Errorf("invalid receipt item at index %d: quantity must be > 0", index)
			}
			return nil
		}),
	); err != nil {
		return err
	}

	if err := validation.Validate(
		item.Sum,
		validation.By(func(value interface{}) error {
			sum, _ := value.(Price8x2)
			if !sum.IsValid() {
				return fmt.Errorf("invalid receipt item at index %d: sum must be within 0..99999999.99", index)
			}
			return nil
		}),
	); err != nil {
		return err
	}

	if item.Cost != nil {
		if err := validation.Validate(
			*item.Cost,
			validation.By(func(value interface{}) error {
				cost, _ := value.(Price8x2)
				if !cost.IsValid() {
					return fmt.Errorf("invalid receipt item at index %d: cost must be within 0..99999999.99", index)
				}
				return nil
			}),
		); err != nil {
			return err
		}
	}

	if err := validation.Validate(item, validation.By(func(_ interface{}) error {
		if item.Sum <= 0 && (item.Cost == nil || *item.Cost <= 0) {
			return fmt.Errorf("invalid receipt item at index %d: sum or cost must be > 0", index)
		}
		return nil
	})); err != nil {
		return err
	}

	if err := validation.Validate(item.Tax, receiptTaxRateRule(index)); err != nil {
		return err
	}
	if item.PaymentMethod != nil {
		if err := validation.Validate(*item.PaymentMethod, receiptPaymentMethodRule(index)); err != nil {
			return err
		}
	}
	if item.PaymentObject != nil {
		if err := validation.Validate(*item.PaymentObject, receiptPaymentObjectRule(index)); err != nil {
			return err
		}
	}

	return nil
}

func isSupportedTaxRate(t TaxRate) bool {
	return internalvalidation.IsSupportedTaxRate(t.String())
}

func isSupportedPaymentMethod(m PaymentMethod) bool {
	return internalvalidation.IsSupportedPaymentMethod(m.String())
}

func isSupportedPaymentObject(o PaymentObject) bool {
	return internalvalidation.IsSupportedPaymentObject(o.String())
}
