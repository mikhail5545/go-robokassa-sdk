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

package validation

import (
	"errors"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	ozzo "github.com/go-ozzo/ozzo-validation/v4"
)

var (
	supportedTaxRates = map[string]struct{}{
		"none": {}, "vat0": {}, "vat10": {}, "vat110": {}, "vat20": {}, "vat22": {},
		"vat120": {}, "vat122": {}, "vat5": {}, "vat7": {}, "vat105": {}, "vat107": {},
	}
	supportedPaymentMethods = map[string]struct{}{
		"full_prepayment": {}, "prepayment": {}, "advance": {}, "full_payment": {},
		"partial_payment": {}, "credit": {}, "credit_payment": {},
	}
	supportedPaymentObjects = map[string]struct{}{
		"commodity": {}, "excise": {}, "job": {}, "service": {}, "gambling_bet": {},
		"gambling_prize": {}, "lottery": {}, "lottery_win": {}, "lottery_prize": {},
		"intellectual_activity": {}, "payment": {}, "agent_commission": {}, "composite": {},
		"resort_fee": {}, "another": {}, "property_right": {}, "non-operating_gain": {},
		"insurance_premium": {}, "sales_tax": {}, "tovar_mark": {},
	}
)

func RequiredTrimmed(value, message string) error {
	return ozzo.Validate(strings.TrimSpace(value), ozzo.Required.Error(message))
}

func PositiveInt64(value int64, message string) error {
	return ozzo.Validate(value, ozzo.Min(int64(1)).Error(message))
}

func StringIn(value, message string, allowed ...string) error {
	options := make([]any, 0, len(allowed))
	for _, option := range allowed {
		options = append(options, option)
	}
	return ozzo.Validate(value, ozzo.In(options...).Error(message))
}

func FirstError(err error, orderedFields ...string) error {
	if err == nil {
		return nil
	}

	var fieldErrors ozzo.Errors
	if !errors.As(err, &fieldErrors) {
		return err
	}

	for _, field := range orderedFields {
		if fieldErr, exists := fieldErrors[field]; exists && fieldErr != nil {
			return fieldErr
		}
	}

	for _, fieldErr := range fieldErrors {
		if fieldErr != nil {
			return fieldErr
		}
	}

	return err
}

func GreaterThanZeroFloatRule(message string) ozzo.Rule {
	return ozzo.By(func(value interface{}) error {
		number, ok := value.(float64)
		if !ok || number <= 0 {
			return errors.New(message)
		}
		return nil
	})
}

func GreaterThanZeroInt64Rule(message string) ozzo.Rule {
	return ozzo.By(func(value interface{}) error {
		number, ok := value.(int64)
		if !ok || number <= 0 {
			return errors.New(message)
		}
		return nil
	})
}

func PositiveDecimalStringRule(invalidNumericMessage, nonPositiveMessage string) ozzo.Rule {
	return ozzo.By(func(value interface{}) error {
		raw, ok := value.(string)
		if !ok {
			return errors.New(invalidNumericMessage)
		}
		parsed, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			return errors.New(invalidNumericMessage)
		}
		if parsed <= 0 {
			return errors.New(nonPositiveMessage)
		}
		return nil
	})
}

func RequiredTrimmedStringRule(message string) ozzo.Rule {
	return ozzo.By(func(value interface{}) error {
		text, ok := value.(string)
		if !ok || strings.TrimSpace(text) == "" {
			return errors.New(message)
		}
		return nil
	})
}

func MaxRuneCountRule(max int, message string) ozzo.Rule {
	return ozzo.By(func(value interface{}) error {
		text, ok := value.(string)
		if !ok || utf8.RuneCountInString(text) > max {
			return errors.New(message)
		}
		return nil
	})
}

func IsSupportedTaxRate(value string) bool {
	_, ok := supportedTaxRates[value]
	return ok
}

func IsSupportedPaymentMethod(value string) bool {
	_, ok := supportedPaymentMethods[value]
	return ok
}

func IsSupportedPaymentObject(value string) bool {
	_, ok := supportedPaymentObjects[value]
	return ok
}

func IsTimeBefore(t1, t2 *time.Time) bool {
	return t1 != nil && t2 != nil && t1.Before(*t2)
}

func IsNotNulAndGreater(f1, f2 *float64) bool {
	return f1 != nil && f2 != nil && *f1 > *f2
}
