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
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/mikhail5545/go-robokassa-sdk/models/receipt"
	"github.com/mikhail5545/go-robokassa-sdk/types"
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
	Culture      *types.Culture
	Encoding     *string

	IsTest bool

	ExpirationDate *time.Time
	Receipt        *receipt.Receipt

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
	merchantLogin := strings.TrimSpace(req.MerchantLogin)
	if merchantLogin == "" {
		merchantLogin = c.merchantLogin
	}
	if merchantLogin == "" {
		return nil, errors.New("merchant login is required")
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

func marshalReceipt(r *receipt.Receipt) (string, error) {
	if r == nil {
		return "", nil
	}
	if len(r.Items) == 0 {
		return "", errors.New("invalid receipt: must contain at least one item")
	}
	if len(r.Items) > 100 {
		return "", errors.New("invalid receipt: cannot contain more than 100 items")
	}
	for i, item := range r.Items {
		if item == nil {
			return "", fmt.Errorf("invalid receipt: item at index %d is nil", i)
		}
		if strings.TrimSpace(item.Name) == "" {
			return "", fmt.Errorf("invalid receipt item at index %d: name is required", i)
		}
		if item.Quantity <= 0 {
			return "", fmt.Errorf("invalid receipt item at index %d: quantity must be > 0", i)
		}
		if item.Sum <= 0 && (item.Cost == nil || *item.Cost <= 0) {
			return "", fmt.Errorf("invalid receipt item at index %d: sum or cost must be > 0", i)
		}
		if !isSupportedTaxRate(item.Tax) {
			return "", fmt.Errorf("invalid receipt item at index %d: unsupported tax rate %q", i, item.Tax)
		}
		if item.PaymentMethod != nil && !isSupportedPaymentMethod(*item.PaymentMethod) {
			return "", fmt.Errorf("invalid receipt item at index %d: unsupported payment_method %q", i, *item.PaymentMethod)
		}
		if item.PaymentObject != nil && !isSupportedPaymentObject(*item.PaymentObject) {
			return "", fmt.Errorf("invalid receipt item at index %d: unsupported payment_object %q", i, *item.PaymentObject)
		}
	}
	b, err := json.Marshal(r)
	if err != nil {
		return "", fmt.Errorf("marshal receipt: %w", err)
	}
	return string(b), nil
}

func normalizeHTTPMethod(method *string) (string, error) {
	raw := trimPtr(method)
	if raw == "" {
		return "", nil
	}
	upper := strings.ToUpper(raw)
	if upper != "GET" && upper != "POST" {
		return "", errors.New("method must be GET or POST")
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
		if key == "" {
			return nil, errors.New("shp key cannot be empty")
		}

		keySuffix := key
		if strings.HasPrefix(strings.ToLower(key), "shp_") {
			keySuffix = key[4:]
		}
		keySuffix = strings.TrimSpace(keySuffix)
		if keySuffix == "" {
			return nil, fmt.Errorf("invalid shp key %q", key)
		}
		if !shpKeySuffixRegex.MatchString(keySuffix) {
			return nil, fmt.Errorf("invalid shp key %q: only latin letters, numbers and underscore are allowed", key)
		}

		canonicalKey := "Shp_" + keySuffix
		if _, exists := out[canonicalKey]; exists {
			return nil, fmt.Errorf("duplicate shp key after normalization: %q", canonicalKey)
		}
		out[canonicalKey] = value
	}

	return out, nil
}

func isSupportedTaxRate(t types.TaxRate) bool {
	switch t {
	case types.TaxRateNone, types.TaxRateVat0, types.TaxRateVat10, types.TaxRateVat110,
		types.TaxRateVat20, types.TaxRateVat22, types.TaxRateVat120, types.TaxRateVat122,
		types.TaxRateVat5, types.TaxRateVat7, types.TaxRateVat105, types.TaxRateVat107:
		return true
	default:
		return false
	}
}

func isSupportedPaymentMethod(m types.PaymentMethod) bool {
	switch m {
	case types.PaymentMethodFullPrepayment, types.PaymentMethodPrepayment, types.PaymentMethodAdvance,
		types.PaymentMethodFullPayment, types.PaymentMethodPartialPayment, types.PaymentMethodCredit,
		types.PaymentMethodCreditPayment:
		return true
	default:
		return false
	}
}

func isSupportedPaymentObject(o types.PaymentObject) bool {
	switch o {
	case types.PaymentObjectCommodity, types.PaymentObjectExcise, types.PaymentObjectJob,
		types.PaymentObjectService, types.PaymentObjectGamblingBet, types.PaymentObjectGamblingPrize,
		types.PaymentObjectLottery, types.PaymentObjectLotteryWin, types.PaymentObjectLotteryPrize,
		types.PaymentObjectIntellectualActivity, types.PaymentObjectPayment, types.PaymentObjectAgentCommission,
		types.PaymentObjectComposite, types.PaymentObjectResortFee, types.PaymentObjectAnother,
		types.PaymentObjectPropertyRight, types.PaymentObjectNonOperatingGain, types.PaymentObjectInsurancePremium,
		types.PaymentObjectSalesTax, types.PaymentObjectProductMark:
		return true
	default:
		return false
	}
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
