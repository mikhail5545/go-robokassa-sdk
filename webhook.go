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

type CallbackNotification struct {
	OutSum         string
	InvID          string
	SignatureValue string
	Fee            string
	Email          string
	PaymentMethod  string
	IncCurrLabel   string
	Culture        string
	Shp            map[string]string
}

type ResultURL2TokenHeader struct {
	Typ string `json:"typ"`
	Alg string `json:"alg"`
}

type ResultURL2PayloadHeader struct {
	Type      string `json:"type"`
	Version   string `json:"version"`
	Timestamp string `json:"timestamp"`
}

type ResultURL2PayloadData struct {
	Shop          string `json:"shop"`
	OpKey         string `json:"opKey"`
	InvID         string `json:"invId"`
	PaymentMethod string `json:"paymentMethod"`
	IncSum        string `json:"incSum"`
	State         string `json:"state"`
}

type ResultURL2Notification struct {
	Header ResultURL2PayloadHeader `json:"header"`
	Data   ResultURL2PayloadData   `json:"data"`
}

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

func (c *Client) ResultSignature(outSum, invID string, shp map[string]string) (string, error) {
	if err := validateRequiredTrimmed(c.password2, "password2 is required to calculate ResultURL signature"); err != nil {
		return "", err
	}
	return c.callbackSignature(outSum, invID, c.password2, shp)
}

func (c *Client) SuccessSignature(outSum, invID string, shp map[string]string) (string, error) {
	return c.callbackSignature(outSum, invID, c.password1, shp)
}

func (c *Client) VerifyResultSignature(outSum, invID, signature string, shp map[string]string) (bool, error) {
	expected, err := c.ResultSignature(outSum, invID, shp)
	if err != nil {
		return false, err
	}
	return strings.EqualFold(strings.TrimSpace(signature), expected), nil
}

func (c *Client) VerifySuccessSignature(outSum, invID, signature string, shp map[string]string) (bool, error) {
	expected, err := c.SuccessSignature(outSum, invID, shp)
	if err != nil {
		return false, err
	}
	return strings.EqualFold(strings.TrimSpace(signature), expected), nil
}

func (c *Client) VerifyResultNotification(notification CallbackNotification) (bool, error) {
	return c.VerifyResultSignature(notification.OutSum, notification.InvID, notification.SignatureValue, notification.Shp)
}

func (c *Client) VerifySuccessNotification(notification CallbackNotification) (bool, error) {
	return c.VerifySuccessSignature(notification.OutSum, notification.InvID, notification.SignatureValue, notification.Shp)
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

func ParseResultURL2JWS(token string) (*ParsedResultURL2JWS, error) {
	headerRaw, payloadRaw, _, _, err := splitJWS(token)
	if err != nil {
		return nil, err
	}

	var tokenHeader ResultURL2TokenHeader
	if err := json.Unmarshal(headerRaw, &tokenHeader); err != nil {
		return nil, fmt.Errorf("unmarshal jws header: %w", err)
	}

	var payload ResultURL2Notification
	if err := json.Unmarshal(payloadRaw, &payload); err != nil {
		return nil, fmt.Errorf("unmarshal jws payload: %w", err)
	}

	return &ParsedResultURL2JWS{
		TokenHeader: tokenHeader,
		Payload:     payload,
	}, nil
}

func VerifyResultURL2JWS(token string, certificateData []byte) error {
	headerRaw, _, signingInput, signature, err := splitJWS(token)
	if err != nil {
		return err
	}

	var tokenHeader ResultURL2TokenHeader
	if err := json.Unmarshal(headerRaw, &tokenHeader); err != nil {
		return fmt.Errorf("unmarshal jws header: %w", err)
	}

	hashType, hasher, err := jwtRSHash(tokenHeader.Alg)
	if err != nil {
		return err
	}

	hasher.Write([]byte(signingInput))
	digest := hasher.Sum(nil)

	cert, err := parseCertificate(certificateData)
	if err != nil {
		return err
	}
	pub, ok := cert.PublicKey.(*rsa.PublicKey)
	if !ok {
		return errors.New("certificate does not contain RSA public key")
	}

	if err := rsa.VerifyPKCS1v15(pub, hashType, digest, signature); err != nil {
		return fmt.Errorf("verify jws signature: %w", err)
	}
	return nil
}

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
