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
	"net/url"
	"strconv"
	"strings"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

const (
	PaymentConfirmURL    = "https://auth.robokassa.ru/Merchant/Payment/Confirm"
	PaymentCancelURL     = "https://auth.robokassa.ru/Merchant/Payment/Cancel"
	PaymentRecurringURL  = "https://auth.robokassa.ru/Merchant/Recurring"
	PaymentCoFPaymentURL = "https://auth.robokassa.ru/Merchant/Payment/CoFPayment"
)

type paymentSignatureProfile string

const (
	paymentSignatureProfileIndex     paymentSignatureProfile = "index"
	paymentSignatureProfileConfirm   paymentSignatureProfile = "confirm"
	paymentSignatureProfileCancel    paymentSignatureProfile = "cancel"
	paymentSignatureProfileRecurring paymentSignatureProfile = "recurring"
	paymentSignatureProfileCoF       paymentSignatureProfile = "cof"
)

// ConfirmPaymentRequest describes params for hold confirmation.
type ConfirmPaymentRequest struct {
	MerchantLogin string
	InvoiceID     int64
	OutSum        float64
	OutSumText    string
	Receipt       *Receipt
	Shp           map[string]string
}

// CancelPaymentRequest describes params for hold cancellation.
type CancelPaymentRequest struct {
	MerchantLogin string
	InvoiceID     int64
	OutSum        *float64
	OutSumText    string
	Shp           map[string]string
}

// RecurringPaymentRequest describes params for a recurring child payment.
type RecurringPaymentRequest struct {
	MerchantLogin     string
	InvoiceID         int64
	PreviousInvoiceID int64
	OutSum            float64
	OutSumText        string
	Description       *string
	Email             *string
	Receipt           *Receipt
	ResultURL2        *string
	Shp               map[string]string
}

func (c *Client) paymentSignatureBaseStringForProfile(profile paymentSignatureProfile, n *normalizedPaymentRequest) string {
	parts := make([]string, 0, 16)
	parts = append(parts, n.merchantLogin)

	switch profile {
	case paymentSignatureProfileCancel:
		parts = append(parts, "", n.invID)
	default:
		parts = append(parts, n.outSum, n.invID)
	}

	appendIfNotEmpty := func(value string) {
		if value != "" {
			parts = append(parts, value)
		}
	}

	switch profile {
	case paymentSignatureProfileIndex:
		appendIfNotEmpty(n.receiptJSON)
		appendIfNotEmpty(n.stepByStep)
		appendIfNotEmpty(n.resultURL2)
		appendIfNotEmpty(n.successURL2)
		appendIfNotEmpty(n.successURL2Method)
		appendIfNotEmpty(n.failURL2)
		appendIfNotEmpty(n.failURL2Method)
		appendIfNotEmpty(n.token)
	case paymentSignatureProfileConfirm:
		appendIfNotEmpty(n.receiptJSON)
	case paymentSignatureProfileRecurring:
		appendIfNotEmpty(n.receiptJSON)
		appendIfNotEmpty(n.resultURL2)
	case paymentSignatureProfileCoF:
		appendIfNotEmpty(n.receiptJSON)
		appendIfNotEmpty(n.resultURL2)
		appendIfNotEmpty(n.token)
	case paymentSignatureProfileCancel:
		// no endpoint modifiers
	}

	parts = append(parts, c.password1)
	for _, key := range sortedKeys(n.shp) {
		parts = append(parts, fmt.Sprintf("%s=%s", key, n.shp[key]))
	}

	return strings.Join(parts, ":")
}

// BuildConfirmPaymentFormValues builds signed form fields for hold confirmation endpoint.
func (c *Client) BuildConfirmPaymentFormValues(req ConfirmPaymentRequest) (url.Values, error) {
	merchantLogin, err := c.normalizeMerchantLogin(req.MerchantLogin)
	if err != nil {
		return nil, err
	}
	invoiceID, err := normalizeInvoiceID(req.InvoiceID)
	if err != nil {
		return nil, err
	}
	outSum, err := normalizeOutSum(&req.OutSum, req.OutSumText)
	if err != nil {
		return nil, err
	}
	receiptJSON, err := marshalReceipt(req.Receipt)
	if err != nil {
		return nil, err
	}
	shp, err := normalizeShpParams(req.Shp)
	if err != nil {
		return nil, err
	}

	n := &normalizedPaymentRequest{
		merchantLogin: merchantLogin,
		outSum:        outSum,
		invID:         invoiceID,
		receiptJSON:   receiptJSON,
		shp:           shp,
	}
	signature, err := c.hashHex(c.paymentSignatureBaseStringForProfile(paymentSignatureProfileConfirm, n))
	if err != nil {
		return nil, err
	}

	values := make(url.Values)
	values.Set("MerchantLogin", merchantLogin)
	values.Set("InvoiceID", invoiceID)
	values.Set("OutSum", outSum)
	values.Set("SignatureValue", signature)
	setValueIfNotEmpty(values, "Receipt", receiptJSON)
	setShpValues(values, shp)
	return values, nil
}

// BuildConfirmPaymentURL builds signed redirect URL for hold confirmation endpoint.
func (c *Client) BuildConfirmPaymentURL(req ConfirmPaymentRequest) (string, error) {
	values, err := c.BuildConfirmPaymentFormValues(req)
	if err != nil {
		return "", err
	}
	return buildURLWithValues(PaymentConfirmURL, values)
}

// BuildCancelPaymentFormValues builds signed form fields for hold cancellation endpoint.
func (c *Client) BuildCancelPaymentFormValues(req CancelPaymentRequest) (url.Values, error) {
	merchantLogin, err := c.normalizeMerchantLogin(req.MerchantLogin)
	if err != nil {
		return nil, err
	}
	invoiceID, err := normalizeInvoiceID(req.InvoiceID)
	if err != nil {
		return nil, err
	}
	outSum, err := normalizeOutSum(req.OutSum, req.OutSumText)
	if err != nil {
		return nil, err
	}
	shp, err := normalizeShpParams(req.Shp)
	if err != nil {
		return nil, err
	}

	n := &normalizedPaymentRequest{
		merchantLogin: merchantLogin,
		invID:         invoiceID,
		shp:           shp,
	}
	signature, err := c.hashHex(c.paymentSignatureBaseStringForProfile(paymentSignatureProfileCancel, n))
	if err != nil {
		return nil, err
	}

	values := make(url.Values)
	values.Set("MerchantLogin", merchantLogin)
	values.Set("InvoiceID", invoiceID)
	values.Set("SignatureValue", signature)
	setValueIfNotEmpty(values, "OutSum", outSum)
	setShpValues(values, shp)
	return values, nil
}

// BuildCancelPaymentURL builds signed redirect URL for hold cancellation endpoint.
func (c *Client) BuildCancelPaymentURL(req CancelPaymentRequest) (string, error) {
	values, err := c.BuildCancelPaymentFormValues(req)
	if err != nil {
		return "", err
	}
	return buildURLWithValues(PaymentCancelURL, values)
}

// BuildRecurringPaymentFormValues builds signed form fields for recurring child payment.
func (c *Client) BuildRecurringPaymentFormValues(req RecurringPaymentRequest) (url.Values, error) {
	n, previousInvoiceID, err := c.normalizeRecurringPaymentRequest(req)
	if err != nil {
		return nil, err
	}
	signature, err := c.hashHex(c.paymentSignatureBaseStringForProfile(paymentSignatureProfileRecurring, n))
	if err != nil {
		return nil, err
	}

	values := make(url.Values)
	values.Set("MerchantLogin", n.merchantLogin)
	values.Set("InvoiceID", n.invID)
	values.Set("PreviousInvoiceID", previousInvoiceID)
	values.Set("OutSum", n.outSum)
	values.Set("SignatureValue", signature)
	setValueIfNotEmpty(values, "Description", n.description)
	setValueIfNotEmpty(values, "Email", n.email)
	setValueIfNotEmpty(values, "Receipt", n.receiptJSON)
	setValueIfNotEmpty(values, "ResultUrl2", n.resultURL2)
	setShpValues(values, n.shp)
	return values, nil
}

// BuildRecurringPaymentURL builds signed redirect URL for recurring child payment.
func (c *Client) BuildRecurringPaymentURL(req RecurringPaymentRequest) (string, error) {
	values, err := c.BuildRecurringPaymentFormValues(req)
	if err != nil {
		return "", err
	}
	return buildURLWithValues(PaymentRecurringURL, values)
}

// BuildCoFPaymentFormValues builds signed form fields for saved-card (CoF) payment.
func (c *Client) BuildCoFPaymentFormValues(req InitPaymentRequest) (url.Values, error) {
	n, err := c.normalizeInitPaymentRequest(req)
	if err != nil {
		return nil, err
	}
	if err := validation.Validate(n.token, validation.Required.Error("token is required for CoF payment")); err != nil {
		return nil, err
	}

	signature, err := c.hashHex(c.paymentSignatureBaseStringForProfile(paymentSignatureProfileCoF, n))
	if err != nil {
		return nil, err
	}

	values := make(url.Values)
	values.Set("MerchantLogin", n.merchantLogin)
	values.Set("OutSum", n.outSum)
	values.Set("SignatureValue", signature)
	values.Set("Token", n.token)
	setValueIfNotEmpty(values, "InvId", n.invID)
	setValueIfNotEmpty(values, "Description", n.description)
	setValueIfNotEmpty(values, "Email", n.email)
	setValueIfNotEmpty(values, "Culture", n.culture)
	setValueIfNotEmpty(values, "Encoding", n.encoding)
	setValueIfNotEmpty(values, "Receipt", n.receiptJSON)
	setValueIfNotEmpty(values, "ResultUrl2", n.resultURL2)
	setValueIfTrue(values, "IsTest", "1", n.isTest)
	setShpValues(values, n.shp)
	return values, nil
}

// BuildCoFPaymentURL builds signed redirect URL for saved-card (CoF) payment.
func (c *Client) BuildCoFPaymentURL(req InitPaymentRequest) (string, error) {
	values, err := c.BuildCoFPaymentFormValues(req)
	if err != nil {
		return "", err
	}
	return buildURLWithValues(PaymentCoFPaymentURL, values)
}

func (c *Client) normalizeRecurringPaymentRequest(
	req RecurringPaymentRequest,
) (*normalizedPaymentRequest, string, error) {
	merchantLogin, err := c.normalizeMerchantLogin(req.MerchantLogin)
	if err != nil {
		return nil, "", err
	}
	invoiceID, err := normalizeInvoiceID(req.InvoiceID)
	if err != nil {
		return nil, "", err
	}
	if err := validation.Validate(
		req.PreviousInvoiceID,
		greaterThanZeroInt64Rule("previous invoice id must be greater than zero"),
	); err != nil {
		return nil, "", err
	}
	outSum, err := normalizeOutSum(&req.OutSum, req.OutSumText)
	if err != nil {
		return nil, "", err
	}
	receiptJSON, err := marshalReceipt(req.Receipt)
	if err != nil {
		return nil, "", err
	}
	shp, err := normalizeShpParams(req.Shp)
	if err != nil {
		return nil, "", err
	}

	normalized := &normalizedPaymentRequest{
		merchantLogin: merchantLogin,
		outSum:        outSum,
		invID:         invoiceID,
		description:   trimPtr(req.Description),
		email:         trimPtr(req.Email),
		receiptJSON:   receiptJSON,
		resultURL2:    trimPtr(req.ResultURL2),
		shp:           shp,
	}
	return normalized, strconv.FormatInt(req.PreviousInvoiceID, 10), nil
}

func (c *Client) normalizeMerchantLogin(merchantLogin string) (string, error) {
	merchantLogin = strings.TrimSpace(merchantLogin)
	if merchantLogin == "" {
		merchantLogin = c.merchantLogin
	}
	if err := validation.Validate(merchantLogin, validation.Required.Error("merchant login is required")); err != nil {
		return "", err
	}
	return merchantLogin, nil
}
