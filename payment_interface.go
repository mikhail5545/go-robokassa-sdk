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
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

const (
	// PaymentGatewayURL is the standard Robokassa payment page URL.
	PaymentGatewayURL = "https://auth.robokassa.ru/Merchant/Index.aspx"
)

var shpKeySuffixRegex = regexp.MustCompile(`^[A-Za-z0-9_]+$`)

// InitPaymentRequest contains fields for pay-interface initialization
// (https://docs.robokassa.ru/ru/pay-interface).
type InitPaymentRequest struct {
	MerchantLogin string
	OutSum        float64
	// OutSumText can be used instead of OutSum to avoid float precision issues
	// in signature-sensitive scenarios. Example: "10.00" or "10.000000".
	OutSumText string

	InvID       *int64
	Description *string
	Email       *string

	IncCurrLabel *string
	Culture      *Culture
	Encoding     *string

	IsTest bool

	ExpirationDate *time.Time
	Receipt        *Receipt

	StepByStep bool

	ResultURL2 *string

	SuccessURL2       *string
	SuccessURL2Method *string
	FailURL2          *string
	FailURL2Method    *string

	Token *string

	PaymentMethods []string
	Recurring      *bool

	// User fields mapped to Shp_* params.
	// Keys can be provided either with or without "Shp_" prefix.
	Shp map[string]string
}

type normalizedPaymentRequest struct {
	merchantLogin string
	outSum        string
	invID         string

	description string
	email       string

	incCurrLabel string
	culture      string
	encoding     string

	isTest bool

	expirationDate string
	receiptJSON    string

	stepByStep string
	resultURL2 string

	successURL2       string
	successURL2Method string
	failURL2          string
	failURL2Method    string

	token string

	paymentMethods []string
	recurring      string

	shp map[string]string
}

// BuildPaymentURL returns a signed redirect URL to Robokassa payment page.
func (c *Client) BuildPaymentURL(req InitPaymentRequest) (string, error) {
	values, err := c.BuildPaymentFormValues(req)
	if err != nil {
		return "", err
	}

	u, err := url.Parse(PaymentGatewayURL)
	if err != nil {
		return "", fmt.Errorf("parse payment gateway url: %w", err)
	}
	u.RawQuery = values.Encode()
	return u.String(), nil
}

// BuildPaymentFormValues returns signed form fields for POSTing to payment page.
func (c *Client) BuildPaymentFormValues(req InitPaymentRequest) (url.Values, error) {
	n, err := c.normalizeInitPaymentRequest(req)
	if err != nil {
		return nil, err
	}

	signature, err := c.hashHex(c.paymentSignatureBaseString(n))
	if err != nil {
		return nil, err
	}

	values := make(url.Values)
	values.Set("MerchantLogin", n.merchantLogin)
	values.Set("OutSum", n.outSum)
	values.Set("SignatureValue", signature)

	if n.invID != "" {
		values.Set("InvId", n.invID)
	}
	if n.description != "" {
		values.Set("Description", n.description)
	}
	if n.email != "" {
		values.Set("Email", n.email)
	}
	if n.incCurrLabel != "" {
		values.Set("IncCurrLabel", n.incCurrLabel)
	}
	if n.culture != "" {
		values.Set("Culture", n.culture)
	}
	if n.encoding != "" {
		values.Set("Encoding", n.encoding)
	}
	if n.isTest {
		values.Set("IsTest", "1")
	}
	if n.expirationDate != "" {
		values.Set("ExpirationDate", n.expirationDate)
	}
	if n.receiptJSON != "" {
		values.Set("Receipt", n.receiptJSON)
	}
	if n.stepByStep != "" {
		values.Set("StepByStep", n.stepByStep)
	}
	if n.resultURL2 != "" {
		values.Set("ResultUrl2", n.resultURL2)
	}
	if n.successURL2 != "" {
		values.Set("SuccessUrl2", n.successURL2)
	}
	if n.successURL2Method != "" {
		values.Set("SuccessUrl2Method", n.successURL2Method)
	}
	if n.failURL2 != "" {
		values.Set("FailUrl2", n.failURL2)
	}
	if n.failURL2Method != "" {
		values.Set("FailUrl2Method", n.failURL2Method)
	}
	if n.token != "" {
		values.Set("Token", n.token)
	}
	for _, method := range n.paymentMethods {
		values.Add("PaymentMethods", method)
	}
	if n.recurring != "" {
		values.Set("Recurring", n.recurring)
	}

	shpKeys := sortedKeys(n.shp)
	for _, key := range shpKeys {
		values.Set(key, n.shp[key])
	}

	return values, nil
}

// CalculatePaymentSignature calculates SignatureValue for a payment request.
func (c *Client) CalculatePaymentSignature(req InitPaymentRequest) (string, error) {
	n, err := c.normalizeInitPaymentRequest(req)
	if err != nil {
		return "", err
	}
	return c.hashHex(c.paymentSignatureBaseString(n))
}

// PaymentSignatureBaseString returns raw signature base string used for SignatureValue.
func (c *Client) PaymentSignatureBaseString(req InitPaymentRequest) (string, error) {
	n, err := c.normalizeInitPaymentRequest(req)
	if err != nil {
		return "", err
	}
	return c.paymentSignatureBaseString(n), nil
}

func (c *Client) normalizeInitPaymentRequest(req InitPaymentRequest) (*normalizedPaymentRequest, error) {
	merchantLogin, err := c.normalizeMerchantLogin(req.MerchantLogin)
	if err != nil {
		return nil, err
	}
	outSum, err := normalizeRequiredOutSum(req.OutSum, req.OutSumText)
	if err != nil {
		return nil, err
	}

	receiptJSON, err := marshalReceipt(req.Receipt)
	if err != nil {
		return nil, err
	}

	successMethod, err := normalizeHTTPMethod(req.SuccessURL2Method)
	if err != nil {
		return nil, fmt.Errorf("invalid SuccessUrl2Method: %w", err)
	}
	failMethod, err := normalizeHTTPMethod(req.FailURL2Method)
	if err != nil {
		return nil, fmt.Errorf("invalid FailUrl2Method: %w", err)
	}

	shp, err := normalizeShpParams(req.Shp)
	if err != nil {
		return nil, err
	}

	paymentMethods := make([]string, 0, len(req.PaymentMethods))
	for _, method := range req.PaymentMethods {
		method = strings.TrimSpace(method)
		if method == "" {
			continue
		}
		paymentMethods = append(paymentMethods, method)
	}

	n := &normalizedPaymentRequest{
		merchantLogin:  merchantLogin,
		outSum:         outSum,
		description:    trimPtr(req.Description),
		email:          trimPtr(req.Email),
		incCurrLabel:   trimPtr(req.IncCurrLabel),
		encoding:       trimPtr(req.Encoding),
		isTest:         req.IsTest,
		receiptJSON:    receiptJSON,
		resultURL2:     trimPtr(req.ResultURL2),
		successURL2:    trimPtr(req.SuccessURL2),
		failURL2:       trimPtr(req.FailURL2),
		token:          trimPtr(req.Token),
		paymentMethods: paymentMethods,
		shp:            shp,
	}

	if req.InvID != nil {
		n.invID = strconv.FormatInt(*req.InvID, 10)
	}
	if req.Culture != nil {
		n.culture = strings.TrimSpace(req.Culture.String())
	}
	if req.ExpirationDate != nil {
		n.expirationDate = req.ExpirationDate.UTC().Format("2006-01-02T15:04")
	}
	if req.StepByStep {
		n.stepByStep = "true"
	}
	if req.Recurring != nil {
		n.recurring = strconv.FormatBool(*req.Recurring)
	}
	n.successURL2Method = successMethod
	n.failURL2Method = failMethod

	return n, nil
}

func (c *Client) paymentSignatureBaseString(n *normalizedPaymentRequest) string {
	return c.paymentSignatureBaseStringForProfile(paymentSignatureProfileIndex, n)
}

func (c *Client) hashHex(input string) (string, error) {
	hashFactory, err := signerForAlgorithm(c.algorithm)
	if err != nil {
		return "", err
	}
	hasher := hashFactory()
	_, _ = hasher.Write([]byte(input))
	return hex.EncodeToString(hasher.Sum(nil)), nil
}

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

func normalizeHTTPMethod(method *string) (string, error) {
	raw := trimPtr(method)
	if raw == "" {
		return "", nil
	}
	upper := strings.ToUpper(raw)
	if err := validation.Validate(upper, validation.In("GET", "POST").Error("method must be GET or POST")); err != nil {
		return "", err
	}
	return upper, nil
}

func normalizeShpParams(in map[string]string) (map[string]string, error) {
	if len(in) == 0 {
		return nil, nil
	}

	out := make(map[string]string, len(in))
	for key, value := range in {
		key = strings.TrimSpace(key)
		if err := validation.Validate(key, validation.Required.Error("shp key cannot be empty")); err != nil {
			return nil, err
		}

		keySuffix := key
		if strings.HasPrefix(strings.ToLower(key), "shp_") {
			keySuffix = key[4:]
		}
		keySuffix = strings.TrimSpace(keySuffix)
		if err := validation.Validate(
			keySuffix,
			validation.Required.Error(fmt.Sprintf("invalid shp key %q", key)),
			validation.Match(shpKeySuffixRegex).Error(
				fmt.Sprintf("invalid shp key %q: only latin letters, numbers and underscore are allowed", key),
			),
		); err != nil {
			return nil, err
		}

		canonicalKey := "Shp_" + keySuffix
		if _, exists := out[canonicalKey]; exists {
			return nil, fmt.Errorf("duplicate shp key after normalization: %q", canonicalKey)
		}
		out[canonicalKey] = value
	}

	return out, nil
}

func isSupportedTaxRate(t TaxRate) bool {
	return supportedTaxRatesRule.Validate(t) == nil
}

func isSupportedPaymentMethod(m PaymentMethod) bool {
	return supportedPaymentMethodsRule.Validate(m) == nil
}

func isSupportedPaymentObject(o PaymentObject) bool {
	return supportedPaymentObjectsRule.Validate(o) == nil
}

func formatOutSum(value float64) string {
	return strconv.FormatFloat(value, 'f', 2, 64)
}

func trimPtr(v *string) string {
	if v == nil {
		return ""
	}
	return strings.TrimSpace(*v)
}

func sortedKeys[K ~string, V any](m map[K]V) []K {
	if len(m) == 0 {
		return nil
	}
	keys := make([]K, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})
	return keys
}
