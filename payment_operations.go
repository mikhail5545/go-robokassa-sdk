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
	"net/url"
	"strconv"
	"strings"

	"github.com/mikhail5545/go-robokassa-sdk/models/receipt"
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
	Receipt       *receipt.Receipt
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
	Receipt           *receipt.Receipt
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

func (c *Client) BuildConfirmPaymentFormValues(req ConfirmPaymentRequest) (url.Values, error) {
	merchantLogin, err := c.normalizeMerchantLogin(req.MerchantLogin)
	if err != nil {
		return nil, err
	}
	invoiceID, err := normalizeInvoiceID(req.InvoiceID)
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
	if receiptJSON != "" {
		values.Set("Receipt", receiptJSON)
	}
	for _, key := range sortedKeys(shp) {
		values.Set(key, shp[key])
	}
	return values, nil
}

func (c *Client) BuildConfirmPaymentURL(req ConfirmPaymentRequest) (string, error) {
	values, err := c.BuildConfirmPaymentFormValues(req)
	if err != nil {
		return "", err
	}
	return buildURLWithValues(PaymentConfirmURL, values)
}

func (c *Client) BuildCancelPaymentFormValues(req CancelPaymentRequest) (url.Values, error) {
	merchantLogin, err := c.normalizeMerchantLogin(req.MerchantLogin)
	if err != nil {
		return nil, err
	}
	invoiceID, err := normalizeInvoiceID(req.InvoiceID)
	if err != nil {
		return nil, err
	}
	outSum, err := normalizeOptionalOutSum(req.OutSum, req.OutSumText)
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
	if outSum != "" {
		values.Set("OutSum", outSum)
	}
	for _, key := range sortedKeys(shp) {
		values.Set(key, shp[key])
	}
	return values, nil
}

func (c *Client) BuildCancelPaymentURL(req CancelPaymentRequest) (string, error) {
	values, err := c.BuildCancelPaymentFormValues(req)
	if err != nil {
		return "", err
	}
	return buildURLWithValues(PaymentCancelURL, values)
}

func (c *Client) BuildRecurringPaymentFormValues(req RecurringPaymentRequest) (url.Values, error) {
	merchantLogin, err := c.normalizeMerchantLogin(req.MerchantLogin)
	if err != nil {
		return nil, err
	}
	invoiceID, err := normalizeInvoiceID(req.InvoiceID)
	if err != nil {
		return nil, err
	}
	previousInvoiceID, err := normalizeInvoiceID(req.PreviousInvoiceID)
	if err != nil {
		return nil, errors.New("previous invoice id must be greater than zero")
	}
	outSum, err := normalizeRequiredOutSum(req.OutSum, req.OutSumText)
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
		description:   trimPtr(req.Description),
		email:         trimPtr(req.Email),
		receiptJSON:   receiptJSON,
		resultURL2:    trimPtr(req.ResultURL2),
		shp:           shp,
	}
	signature, err := c.hashHex(c.paymentSignatureBaseStringForProfile(paymentSignatureProfileRecurring, n))
	if err != nil {
		return nil, err
	}

	values := make(url.Values)
	values.Set("MerchantLogin", merchantLogin)
	values.Set("InvoiceID", invoiceID)
	values.Set("PreviousInvoiceID", previousInvoiceID)
	values.Set("OutSum", outSum)
	values.Set("SignatureValue", signature)
	if n.description != "" {
		values.Set("Description", n.description)
	}
	if n.email != "" {
		values.Set("Email", n.email)
	}
	if n.receiptJSON != "" {
		values.Set("Receipt", n.receiptJSON)
	}
	if n.resultURL2 != "" {
		values.Set("ResultUrl2", n.resultURL2)
	}
	for _, key := range sortedKeys(shp) {
		values.Set(key, shp[key])
	}
	return values, nil
}

func (c *Client) BuildRecurringPaymentURL(req RecurringPaymentRequest) (string, error) {
	values, err := c.BuildRecurringPaymentFormValues(req)
	if err != nil {
		return "", err
	}
	return buildURLWithValues(PaymentRecurringURL, values)
}

func (c *Client) BuildCoFPaymentFormValues(req InitPaymentRequest) (url.Values, error) {
	n, err := c.normalizeInitPaymentRequest(req)
	if err != nil {
		return nil, err
	}
	if n.token == "" {
		return nil, errors.New("token is required for CoF payment")
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
	if n.invID != "" {
		values.Set("InvId", n.invID)
	}
	if n.description != "" {
		values.Set("Description", n.description)
	}
	if n.email != "" {
		values.Set("Email", n.email)
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
	if n.receiptJSON != "" {
		values.Set("Receipt", n.receiptJSON)
	}
	if n.resultURL2 != "" {
		values.Set("ResultUrl2", n.resultURL2)
	}
	for _, key := range sortedKeys(n.shp) {
		values.Set(key, n.shp[key])
	}
	return values, nil
}

func (c *Client) BuildCoFPaymentURL(req InitPaymentRequest) (string, error) {
	values, err := c.BuildCoFPaymentFormValues(req)
	if err != nil {
		return "", err
	}
	return buildURLWithValues(PaymentCoFPaymentURL, values)
}

func (c *Client) normalizeMerchantLogin(merchantLogin string) (string, error) {
	merchantLogin = strings.TrimSpace(merchantLogin)
	if merchantLogin == "" {
		merchantLogin = c.merchantLogin
	}
	if merchantLogin == "" {
		return "", errors.New("merchant login is required")
	}
	return merchantLogin, nil
}

func normalizeInvoiceID(invoiceID int64) (string, error) {
	if invoiceID <= 0 {
		return "", errors.New("invoice id must be greater than zero")
	}
	return strconv.FormatInt(invoiceID, 10), nil
}

func buildURLWithValues(rawURL string, values url.Values) (string, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("parse endpoint url: %w", err)
	}
	u.RawQuery = values.Encode()
	return u.String(), nil
}
