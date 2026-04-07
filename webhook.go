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
	"crypto/rsa"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

// CallbackNotification represents a payment notification on ResultURL.
// The ResultURL is designed to [automatically notify] your website of a successful payment.
//
// [automatically notify]: https://docs.robokassa.ru/ru/notifications-and-redirects#resulturl
type CallbackNotification struct {
	// Payment amount.
	OutSum string
	// Store account number.
	InvID string
	// The checksum - hash, calculated using the method specified in the store's technical settings.
	SignatureValue string
	// Robokassa commission for completing a transaction. Changed according to the customer's rate.
	Fee string
	// Email specified by the buyer during payment.
	Email string
	// The payment method used by the user.
	PaymentMethod string
	// The currency in which the client paid.
	IncCurrLabel string
	Culture      string
	// Custom parameters are returned if they were passed when starting a payment.
	Shp map[string]string
}

// ResultURL2TokenHeader is JWS token header from ResultURL2 callback.
type ResultURL2TokenHeader struct {
	Typ string `json:"typ"`
	Alg string `json:"alg"`
}

// ResultURL2PayloadHeader contains ResultURL2Notification.Header payload
type ResultURL2PayloadHeader struct {
	Type      string `json:"type"`
	Version   string `json:"version"`
	Timestamp string `json:"timestamp"`
}

// ResultURL2PayloadData contains ResultURL2Notification.Data payload
type ResultURL2PayloadData struct {
	// Store ID.
	Shop string `json:"shop"`
	// Unique identifier of the operation.
	OpKey string `json:"opKey"`
	// Store account number.
	InvID string `json:"invId"`
	// The payment method used by user.
	PaymentMethod string `json:"paymentMethod"`
	// The amount paid by the client.
	IncSum string `json:"incSum"`
	// Current payment status.
	State string `json:"state"`
}

// ResultURL2Notification represents an additional notification of successful payment,
// allowing you to receive data to an alternative address (you need to provide it, e.g. [CreateInvoiceRequest.SuccessURL2Data]).
//
// For transactions with holds, a pre-authorization notification is sent to this address (the only way to receive it).
type ResultURL2Notification struct {
	Header ResultURL2PayloadHeader `json:"header"`
	Data   ResultURL2PayloadData   `json:"data"`
}

// ParsedResultURL2JWS contains decoded token header and payload.
type ParsedResultURL2JWS struct {
	TokenHeader ResultURL2TokenHeader
	Payload     ResultURL2Notification
}

// ResultAcknowledgement returns the exact response body expected by Robokassa
// after successful ResultURL validation.
func ResultAcknowledgement(invID string) (string, error) {
	invID = strings.TrimSpace(invID)
	if err := validateRequiredTrimmed(invID, "inv id is required"); err != nil {
		return "", err
	}
	return "OK" + invID, nil
}

// ParseCallbackNotification extracts callback payload fields from ResultURL/SuccessURL query values.
func ParseCallbackNotification(values url.Values) CallbackNotification {
	shp := make(map[string]string)
	for key, val := range values {
		if !strings.HasPrefix(strings.ToLower(key), "shp_") {
			continue
		}
		if len(val) > 0 {
			shp[key] = val[0]
		}
	}

	return CallbackNotification{
		OutSum:         callbackValue(values, "OutSum", "outSum"),
		InvID:          callbackValue(values, "InvId", "InvID", "invId", "invID"),
		SignatureValue: callbackValue(values, "SignatureValue", "signatureValue"),
		Fee:            callbackValue(values, "Fee", "fee"),
		Email:          callbackValue(values, "EMail", "Email", "email"),
		PaymentMethod:  callbackValue(values, "PaymentMethod", "paymentMethod"),
		IncCurrLabel:   callbackValue(values, "IncCurrLabel", "incCurrLabel"),
		Culture:        callbackValue(values, "Culture", "culture"),
		Shp:            shp,
	}
}

// ResultSignature calculates SignatureValue for ResultURL notifications.
//
// Base string format: OutSum:InvId:Password2[:Shp_*...]
func (c *Client) ResultSignature(outSum, invID string, shp map[string]string) (string, error) {
	if err := validateRequiredTrimmed(c.password2, "password2 is required to calculate ResultURL signature"); err != nil {
		return "", err
	}
	return c.callbackSignature(outSum, invID, c.password2, shp)
}

// SuccessSignature calculates SignatureValue for SuccessURL redirects.
//
// Base string format: OutSum:InvId:Password1[:Shp_*...]
func (c *Client) SuccessSignature(outSum, invID string, shp map[string]string) (string, error) {
	return c.callbackSignature(outSum, invID, c.password1, shp)
}

// VerifyCallbackSignature validates callback SignatureValue for the requested callback kind.
//
// On mismatch, it returns [CallbackSignatureMismatchError] and matches [ErrInvalidCallbackSignature]
// via errors.Is.
func (c *Client) VerifyCallbackSignature(
	kind CallbackSignatureKind,
	outSum,
	invID,
	signature string,
	shp map[string]string,
) error {
	if err := validateRequiredTrimmed(signature, "signature value is required"); err != nil {
		return err
	}
	expected, err := c.expectedCallbackSignature(kind, outSum, invID, shp)
	if err != nil {
		return err
	}
	if strings.EqualFold(strings.TrimSpace(signature), expected) {
		return nil
	}
	return &CallbackSignatureMismatchError{
		Kind:  kind,
		InvID: strings.TrimSpace(invID),
	}
}

// VerifyResultSignature validates ResultURL SignatureValue.
func (c *Client) VerifyResultSignature(outSum, invID, signature string, shp map[string]string) error {
	return c.VerifyCallbackSignature(CallbackSignatureKindResult, outSum, invID, signature, shp)
}

// VerifySuccessSignature validates SuccessURL SignatureValue.
func (c *Client) VerifySuccessSignature(outSum, invID, signature string, shp map[string]string) error {
	return c.VerifyCallbackSignature(CallbackSignatureKindSuccess, outSum, invID, signature, shp)
}

// VerifyResultNotification validates SignatureValue in parsed ResultURL callback payload.
func (c *Client) VerifyResultNotification(notification CallbackNotification) error {
	return c.VerifyResultSignature(notification.OutSum, notification.InvID, notification.SignatureValue, notification.Shp)
}

// VerifySuccessNotification validates SignatureValue in parsed SuccessURL callback payload.
func (c *Client) VerifySuccessNotification(notification CallbackNotification) error {
	return c.VerifySuccessSignature(notification.OutSum, notification.InvID, notification.SignatureValue, notification.Shp)
}

func (c *Client) expectedCallbackSignature(
	kind CallbackSignatureKind,
	outSum,
	invID string,
	shp map[string]string,
) (string, error) {
	switch kind {
	case CallbackSignatureKindResult:
		return c.ResultSignature(outSum, invID, shp)
	case CallbackSignatureKindSuccess:
		return c.SuccessSignature(outSum, invID, shp)
	default:
		return "", fmt.Errorf("%w: %q", ErrUnsupportedCallbackSignatureKind, kind)
	}
}

func (c *Client) callbackSignature(outSum, invID, password string, shp map[string]string) (string, error) {
	outSum = strings.TrimSpace(outSum)
	invID = strings.TrimSpace(invID)
	if err := validateRequiredTrimmed(outSum, "out sum is required"); err != nil {
		return "", err
	}
	if err := validateRequiredTrimmed(invID, "inv id is required"); err != nil {
		return "", err
	}

	normalizedShp, err := normalizeShpParams(shp)
	if err != nil {
		return "", err
	}

	parts := []string{
		outSum,
		invID,
		password,
	}

	shpKeys := make([]string, 0, len(normalizedShp))
	for key := range normalizedShp {
		shpKeys = append(shpKeys, key)
	}
	sort.Strings(shpKeys)

	for _, key := range shpKeys {
		parts = append(parts, fmt.Sprintf("%s=%s", key, normalizedShp[key]))
	}

	return c.hashHex(strings.Join(parts, ":"))
}

// ParseResultURL2JWS parses token header and payload from ResultURL2 JWS callback.
func ParseResultURL2JWS(token string) (*ParsedResultURL2JWS, error) {
	headerRaw, payloadRaw, _, _, err := splitJWS(token)
	if err != nil {
		return nil, wrapResultURL2Error(ErrResultURL2InvalidToken, err)
	}

	var tokenHeader ResultURL2TokenHeader
	if err := json.Unmarshal(headerRaw, &tokenHeader); err != nil {
		return nil, wrapResultURL2Error(ErrResultURL2InvalidHeader, fmt.Errorf("unmarshal jws header: %w", err))
	}

	var payload ResultURL2Notification
	if err := json.Unmarshal(payloadRaw, &payload); err != nil {
		return nil, wrapResultURL2Error(ErrResultURL2InvalidPayload, fmt.Errorf("unmarshal jws payload: %w", err))
	}

	return &ParsedResultURL2JWS{
		TokenHeader: tokenHeader,
		Payload:     payload,
	}, nil
}

// VerifyResultURL2JWS verifies RSA JWS signature for ResultURL2 payload.
func VerifyResultURL2JWS(token string, certificateData []byte) error {
	headerRaw, _, signingInput, signature, err := splitJWS(token)
	if err != nil {
		return wrapResultURL2Error(ErrResultURL2InvalidToken, err)
	}

	var tokenHeader ResultURL2TokenHeader
	if err := json.Unmarshal(headerRaw, &tokenHeader); err != nil {
		return wrapResultURL2Error(ErrResultURL2InvalidHeader, fmt.Errorf("unmarshal jws header: %w", err))
	}

	hashType, hasher, err := jwtRSHash(tokenHeader.Alg)
	if err != nil {
		return wrapResultURL2Error(ErrResultURL2UnsupportedAlgorithm, err)
	}

	hasher.Write([]byte(signingInput))
	digest := hasher.Sum(nil)

	cert, err := parseCertificate(certificateData)
	if err != nil {
		return wrapResultURL2Error(ErrResultURL2InvalidCertificate, err)
	}
	pub, ok := cert.PublicKey.(*rsa.PublicKey)
	if !ok {
		return wrapResultURL2Error(ErrResultURL2InvalidCertificateKey, errors.New("certificate does not contain RSA public key"))
	}

	if err := rsa.VerifyPKCS1v15(pub, hashType, digest, signature); err != nil {
		return wrapResultURL2Error(ErrResultURL2SignatureVerification, fmt.Errorf("verify jws signature: %w", err))
	}
	return nil
}

// ParsedTimestamp converts payload header timestamp (unix seconds) to UTC time.
func (r ResultURL2Notification) ParsedTimestamp() (time.Time, error) {
	if err := validateRequiredTrimmed(r.Header.Timestamp, "timestamp is empty"); err != nil {
		return time.Time{}, err
	}
	sec, err := strconv.ParseInt(r.Header.Timestamp, 10, 64)
	if err != nil {
		return time.Time{}, fmt.Errorf("parse timestamp: %w", err)
	}
	return time.Unix(sec, 0).UTC(), nil
}
