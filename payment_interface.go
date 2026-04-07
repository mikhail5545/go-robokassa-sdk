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
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	// PaymentGatewayURL is the standard Robokassa payment page URL.
	PaymentGatewayURL = "https://auth.robokassa.ru/Merchant/Index.aspx"
)

// InitPaymentRequest contains fields for [pay-interface] initialization
//
// [pay-interface]: https://docs.robokassa.ru/ru/pay-interface
type InitPaymentRequest struct {
	// Store login specified in technical settings
	MerchantLogin string
	// OutSum is the amount to be paid.
	OutSum float64
	// OutSumText can be used instead of OutSum to avoid float precision issues
	// in signature-sensitive scenarios. Example: "10.00" or "10.000000".
	OutSumText string

	// InvID is the store invoice number. This parameter is optional, but it's strongly recommended using id by [official documentation].
	// The number must be unique for each payment. Valid values are from 1 to int64 max. If the value is empty, equals 0 or is not specified,
	// a unique value will be automatically assigned to it when the payment transaction is created.
	//
	// [official documentation]:  https://docs.robokassa.ru/ru/pay-interface#optional-parameters
	InvID *int64
	// Name of product or service (up to 100 characters, without special characters).
	Description *string
	// Buyer's email address. Used for receipts and notifications.
	Email *string

	// Suggested payment method. The payment method you recommend to your customers.
	IncCurrLabel *string
	// Interface language
	Culture *Culture
	// Encoding of transmitted data (default UTF-8)
	Encoding *string

	// Enabling test mode
	IsTest bool

	// Payment deadline
	ExpirationDate *time.Time
	// Fascial data
	Receipt *Receipt

	// A hold flag which means payment is processed in two stages.
	// For details, see [Hold and Pre-Authorization]
	//
	// [Hold and Pre-Authorization]: https://docs.robokassa.ru/ru/holding
	StepByStep bool

	// Additional server callback. For hold ResultURL2, specify to receive a notification Result2 and include
	// it in the signature along with StepByStep. For details, see [Additional payment notification in Result2].
	//
	// [Additional payment notification in Result2]: https://docs.robokassa.ru/ru/notifications-and-redirects#resulturl2
	ResultURL2 *string

	// An additional return address for successful payment. For details, see [Additional Redirect].
	//
	// [Additional Redirect]: https://docs.robokassa.ru/ru/notifications-and-redirects#returnurl
	SuccessURL2 *string
	// Query method to SuccessURL2 (GET or POST).
	SuccessURL2Method *string
	// An additional return address for errors. For details, see [AdditionalRedirect (ReturnURL: FailUrl2)].
	//
	// [AdditionalRedirect (ReturnURL: FailUrl2)]: https://docs.robokassa.ru/ru/notifications-and-redirects#returnurl
	FailURL2 *string
	// Query method to FailURL2 (GET or POST).
	FailURL2Method *string

	// Saved card token. For details, see [Paying with a saved card].
	//
	// [Paying with a saved card]: https://docs.robokassa.ru/ru/saving
	Token *string

	// Specifies a list of available payment methods.
	PaymentMethods []string
	// Recurring payment flag.
	Recurring *bool

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

	setValueIfNotEmpty(values, "InvId", n.invID)
	setValueIfNotEmpty(values, "Description", n.description)
	setValueIfNotEmpty(values, "Email", n.email)
	setValueIfNotEmpty(values, "IncCurrLabel", n.incCurrLabel)
	setValueIfNotEmpty(values, "Culture", n.culture)
	setValueIfNotEmpty(values, "Encoding", n.encoding)
	setValueIfNotEmpty(values, "ExpirationDate", n.expirationDate)
	setValueIfNotEmpty(values, "Receipt", n.receiptJSON)
	setValueIfNotEmpty(values, "StepByStep", n.stepByStep)
	setValueIfNotEmpty(values, "ResultUrl2", n.resultURL2)
	setValueIfNotEmpty(values, "SuccessUrl2", n.successURL2)
	setValueIfNotEmpty(values, "SuccessUrl2Method", n.successURL2Method)
	setValueIfNotEmpty(values, "FailUrl2", n.failURL2)
	setValueIfNotEmpty(values, "FailUrl2Method", n.failURL2Method)
	setValueIfNotEmpty(values, "Token", n.token)
	setValueIfNotEmpty(values, "Recurring", n.recurring)
	setValueIfTrue(values, "IsTest", "1", n.isTest)
	addPaymentMethods(values, n.paymentMethods)
	setShpValues(values, n.shp)

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
		paymentMethods: normalizePaymentMethods(req.PaymentMethods),
		shp:            shp,
	}
	applyOptionalInitPaymentFields(n, req, successMethod, failMethod)

	return n, nil
}

func normalizePaymentMethods(methods []string) []string {
	normalized := make([]string, 0, len(methods))
	for _, method := range methods {
		method = strings.TrimSpace(method)
		if method == "" {
			continue
		}
		normalized = append(normalized, method)
	}
	return normalized
}

func applyOptionalInitPaymentFields(
	n *normalizedPaymentRequest,
	req InitPaymentRequest,
	successMethod string,
	failMethod string,
) {
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
