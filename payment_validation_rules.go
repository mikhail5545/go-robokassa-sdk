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
	"errors"
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

var (
	supportedTaxRatesRule = validation.In(
		TaxRateNone, TaxRateVat0, TaxRateVat10, TaxRateVat110,
		TaxRateVat20, TaxRateVat22, TaxRateVat120, TaxRateVat122,
		TaxRateVat5, TaxRateVat7, TaxRateVat105, TaxRateVat107,
	)
	supportedPaymentMethodsRule = validation.In(
		PaymentMethodFullPrepayment, PaymentMethodPrepayment, PaymentMethodAdvance,
		PaymentMethodFullPayment, PaymentMethodPartialPayment, PaymentMethodCredit,
		PaymentMethodCreditPayment,
	)
	supportedPaymentObjectsRule = validation.In(
		PaymentObjectCommodity, PaymentObjectExcise, PaymentObjectJob,
		PaymentObjectService, PaymentObjectGamblingBet, PaymentObjectGamblingPrize,
		PaymentObjectLottery, PaymentObjectLotteryWin, PaymentObjectLotteryPrize,
		PaymentObjectIntellectualActivity, PaymentObjectPayment, PaymentObjectAgentCommission,
		PaymentObjectComposite, PaymentObjectResortFee, PaymentObjectAnother,
		PaymentObjectPropertyRight, PaymentObjectNonOperatingGain, PaymentObjectInsurancePremium,
		PaymentObjectSalesTax, PaymentObjectProductMark,
	)
)

func greaterThanZeroFloatRule(message string) validation.Rule {
	return validation.By(func(value interface{}) error {
		number, ok := value.(float64)
		if !ok || number <= 0 {
			return errors.New(message)
		}
		return nil
	})
}

func greaterThanZeroInt64Rule(message string) validation.Rule {
	return validation.By(func(value interface{}) error {
		number, ok := value.(int64)
		if !ok || number <= 0 {
			return errors.New(message)
		}
		return nil
	})
}

func positiveDecimalStringRule(invalidNumericMessage, nonPositiveMessage string) validation.Rule {
	return validation.By(func(value interface{}) error {
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

func requiredTrimmedStringRule(message string) validation.Rule {
	return validation.By(func(value interface{}) error {
		text, ok := value.(string)
		if !ok || strings.TrimSpace(text) == "" {
			return errors.New(message)
		}
		return nil
	})
}

func maxRuneCountRule(max int, message string) validation.Rule {
	return validation.By(func(value interface{}) error {
		text, ok := value.(string)
		if !ok || utf8.RuneCountInString(text) > max {
			return errors.New(message)
		}
		return nil
	})
}

func receiptTaxRateRule(index int) validation.Rule {
	return validation.By(func(value interface{}) error {
		taxRate, ok := value.(TaxRate)
		if !ok || supportedTaxRatesRule.Validate(taxRate) != nil {
			return fmt.Errorf("invalid receipt item at index %d: unsupported tax rate %q", index, value)
		}
		return nil
	})
}

func receiptPaymentMethodRule(index int) validation.Rule {
	return validation.By(func(value interface{}) error {
		paymentMethod, ok := value.(PaymentMethod)
		if !ok || supportedPaymentMethodsRule.Validate(paymentMethod) != nil {
			return fmt.Errorf("invalid receipt item at index %d: unsupported payment_method %q", index, value)
		}
		return nil
	})
}

func receiptPaymentObjectRule(index int) validation.Rule {
	return validation.By(func(value interface{}) error {
		paymentObject, ok := value.(PaymentObject)
		if !ok || supportedPaymentObjectsRule.Validate(paymentObject) != nil {
			return fmt.Errorf("invalid receipt item at index %d: unsupported payment_object %q", index, value)
		}
		return nil
	})
}
