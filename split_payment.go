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
	"strconv"
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
		validation.Field(&i.OutAmount, validation.By(func(interface{}) error {
			if i.OutAmount <= 0 {
				return errors.New("out amount must be greater than zero")
			}
			outAmountRaw, err := outAmountToRaw(i.OutAmount)
			if err != nil {
				return err
			}
			if outAmountRaw > maxPrice8x2Raw {
				return errors.New("out amount must be within 0.01..99999999.99")
			}
			return nil
		})),
		validation.Field(&i.Merchant, validation.By(func(interface{}) error {
			if strings.TrimSpace(i.Merchant.ID) == "" {
				return errors.New("merchant.id is required")
			}
			return nil
		})),
		validation.Field(&i.SplitMerchants, validation.By(func(interface{}) error {
			if len(i.SplitMerchants) == 0 {
				return errors.New("at least one split merchant is required")
			}
			for idx, merchant := range i.SplitMerchants {
				if strings.TrimSpace(merchant.ID) == "" {
					return fmt.Errorf("splitMerchants[%d].id is required", idx)
				}
				if !merchant.Amount.IsValid() {
					return fmt.Errorf("splitMerchants[%d].amount must be within 0.00..99999999.99", idx)
				}
			}
			return nil
		})),
		validation.Field(&i.ShopParams, validation.By(func(interface{}) error {
			for idx, param := range i.ShopParams {
				if strings.TrimSpace(param.Name) == "" {
					return fmt.Errorf("shop_params[%d].name is required", idx)
				}
			}
			return nil
		})),
	)
	if err = firstValidationError(err, "OutAmount", "Merchant", "SplitMerchants", "ShopParams"); err != nil {
		return err
	}

	return validation.Validate(i, validation.By(func(interface{}) error {
		outAmountRaw, err := outAmountToRaw(i.OutAmount)
		if err != nil {
			return err
		}
		totalSplitRaw := int64(0)
		for _, merchant := range i.SplitMerchants {
			totalSplitRaw += int64(merchant.Amount)
		}
		if totalSplitRaw != outAmountRaw {
			return errors.New("sum of splitMerchants.amount must equal outAmount")
		}
		return nil
	}))
}

func outAmountToRaw(value float64) (int64, error) {
	normalized := formatOutSum(value)
	parts := strings.SplitN(normalized, ".", 2)
	if len(parts) != 2 {
		return 0, errors.New("out amount has invalid format")
	}
	whole, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return 0, errors.New("out amount has invalid integer part")
	}
	frac, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return 0, errors.New("out amount has invalid fractional part")
	}
	return whole*100 + frac, nil
}
