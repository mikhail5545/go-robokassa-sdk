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

	validation "github.com/go-ozzo/ozzo-validation/v4"

	internalvalidation "github.com/mikhail5545/go-robokassa-sdk/internal/validation"
)

func greaterThanZeroFloatRule(message string) validation.Rule {
	return internalvalidation.GreaterThanZeroFloatRule(message)
}

func greaterThanZeroInt64Rule(message string) validation.Rule {
	return internalvalidation.GreaterThanZeroInt64Rule(message)
}

func positiveDecimalStringRule(invalidNumericMessage, nonPositiveMessage string) validation.Rule {
	return internalvalidation.PositiveDecimalStringRule(invalidNumericMessage, nonPositiveMessage)
}

func requiredTrimmedStringRule(message string) validation.Rule {
	return internalvalidation.RequiredTrimmedStringRule(message)
}

func maxRuneCountRule(max int, message string) validation.Rule {
	return internalvalidation.MaxRuneCountRule(max, message)
}

func receiptTaxRateRule(index int) validation.Rule {
	return validation.By(func(value interface{}) error {
		taxRate, ok := value.(TaxRate)
		if !ok || !internalvalidation.IsSupportedTaxRate(taxRate.String()) {
			return fmt.Errorf("invalid receipt item at index %d: unsupported tax rate %q", index, value)
		}
		return nil
	})
}

func receiptPaymentMethodRule(index int) validation.Rule {
	return validation.By(func(value interface{}) error {
		paymentMethod, ok := value.(PaymentMethod)
		if !ok || !internalvalidation.IsSupportedPaymentMethod(paymentMethod.String()) {
			return fmt.Errorf("invalid receipt item at index %d: unsupported payment_method %q", index, value)
		}
		return nil
	})
}

func receiptPaymentObjectRule(index int) validation.Rule {
	return validation.By(func(value interface{}) error {
		paymentObject, ok := value.(PaymentObject)
		if !ok || !internalvalidation.IsSupportedPaymentObject(paymentObject.String()) {
			return fmt.Errorf("invalid receipt item at index %d: unsupported payment_object %q", index, value)
		}
		return nil
	})
}
