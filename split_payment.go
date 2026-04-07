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
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

const SplitPaymentGatewayURL = "https://auth.robokassa.ru/Merchant/Payment/CreateV2"

type SplitShopParam struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type SplitMasterMerchant struct {
	ID      string `json:"id"`
	Comment string `json:"comment,omitempty"`
}

type SplitPaymentInvoice struct {
	OutAmount      float64             `json:"outAmount"`
	ShopParams     []SplitShopParam    `json:"shop_params,omitempty"`
	Email          *string             `json:"email,omitempty"`
	IncCurr        *string             `json:"incCurr,omitempty"`
	Language       *Culture            `json:"language,omitempty"`
	ExpirationDate *time.Time          `json:"expirationDate,omitempty"`
	Merchant       SplitMasterMerchant `json:"merchant"`
	SplitMerchants []SplitMerchant     `json:"splitMerchants"`
}

func (c *Client) BuildSplitPaymentFormValues(invoice SplitPaymentInvoice) (url.Values, error) {
	if err := invoice.validate(); err != nil {
		return nil, err
	}

	invoiceJSONBytes, err := json.Marshal(invoice)
	if err != nil {
		return nil, fmt.Errorf("marshal split invoice: %w", err)
	}
	invoiceJSON := string(invoiceJSONBytes)

	signatureBase := invoiceJSON + ":" + c.password1
	signature, err := c.hashHex(signatureBase)
	if err != nil {
		return nil, err
	}

	values := make(url.Values)
	values.Set("Invoice", invoiceJSON)
	values.Set("Signature", signature)
	return values, nil
}

func (c *Client) BuildSplitPaymentURL(invoice SplitPaymentInvoice) (string, error) {
	values, err := c.BuildSplitPaymentFormValues(invoice)
	if err != nil {
		return "", err
	}
	u, err := url.Parse(SplitPaymentGatewayURL)
	if err != nil {
		return "", fmt.Errorf("parse split payment gateway url: %w", err)
	}
	u.RawQuery = values.Encode()
	return u.String(), nil
}

func (i SplitPaymentInvoice) validate() error {
	err := validation.ValidateStruct(&i,
		validation.Field(&i.OutAmount, validation.By(func(interface{}) error { return validateSplitOutAmount(i.OutAmount) })),
		validation.Field(&i.Merchant, validation.By(func(interface{}) error { return validateSplitMasterMerchant(i.Merchant) })),
		validation.Field(
			&i.SplitMerchants,
			validation.By(func(interface{}) error { return validateSplitMerchants(i.SplitMerchants) }),
		),
		validation.Field(&i.ShopParams, validation.By(func(interface{}) error { return validateSplitShopParams(i.ShopParams) })),
	)
	if err = firstValidationError(err, "OutAmount", "Merchant", "SplitMerchants", "ShopParams"); err != nil {
		return err
	}

	return validateSplitTotalAmount(i.OutAmount, i.SplitMerchants)
}

func validateSplitOutAmount(outAmount float64) error {
	if outAmount <= 0 {
		return errors.New("out amount must be greater than zero")
	}
	outAmountRaw, err := outAmountToRaw(outAmount)
	if err != nil {
		return err
	}
	if outAmountRaw > maxPrice8x2Raw {
		return errors.New("out amount must be within 0.01..99999999.99")
	}
	return nil
}

func validateSplitMasterMerchant(merchant SplitMasterMerchant) error {
	if strings.TrimSpace(merchant.ID) == "" {
		return errors.New("merchant.id is required")
	}
	return nil
}

func validateSplitMerchants(splitMerchants []SplitMerchant) error {
	if len(splitMerchants) == 0 {
		return errors.New("at least one split merchant is required")
	}
	for idx, merchant := range splitMerchants {
		if strings.TrimSpace(merchant.ID) == "" {
			return fmt.Errorf("splitMerchants[%d].id is required", idx)
		}
		if !merchant.Amount.IsValid() {
			return fmt.Errorf("splitMerchants[%d].amount must be within 0.00..99999999.99", idx)
		}
	}
	return nil
}

func validateSplitShopParams(shopParams []SplitShopParam) error {
	for idx, param := range shopParams {
		if strings.TrimSpace(param.Name) == "" {
			return fmt.Errorf("shop_params[%d].name is required", idx)
		}
	}
	return nil
}

func validateSplitTotalAmount(outAmount float64, splitMerchants []SplitMerchant) error {
	outAmountRaw, err := outAmountToRaw(outAmount)
	if err != nil {
		return err
	}
	totalSplitRaw := int64(0)
	for _, merchant := range splitMerchants {
		totalSplitRaw += int64(merchant.Amount)
	}
	if totalSplitRaw != outAmountRaw {
		return errors.New("sum of splitMerchants.amount must equal outAmount")
	}
	return nil
}
