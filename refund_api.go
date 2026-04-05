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
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"hash"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/mikhail5545/go-robokassa-sdk/models/items"
)

const (
	refundCreatePath = "/Refund/Create"
	refundStatePath  = "/Refund/GetState"
)

type RefundStatus string

const (
	RefundStatusFinished   RefundStatus = "finished"
	RefundStatusProcessing RefundStatus = "processing"
	RefundStatusCanceled   RefundStatus = "canceled"
)

type CreateRefundRequest struct {
	OpKey        string               `json:"OpKey"`
	RefundSum    *float64             `json:"RefundSum,omitempty"`
	InvoiceItems []*items.InvoiceItem `json:"InvoiceItems,omitempty"`
}

type CreateRefundResponse struct {
	Success     bool
	Message     string
	RequestID   string
	RawResponse RawResponse
}

type RefundStateResponse struct {
	RequestID   string
	Amount      *float64
	Label       RefundStatus
	Message     string
	RawResponse RawResponse
}

func (c *Client) CreateRefund(ctx context.Context, req CreateRefundRequest) (*CreateRefundResponse, error) {
	if strings.TrimSpace(c.password3) == "" {
		return nil, errors.New("password3 is required for refund api")
	}
	if err := req.validate(); err != nil {
		return nil, err
	}

	raw, err := c.doRefundJWTRequest(ctx, refundCreatePath, req)
	if err != nil {
		return nil, err
	}

	response := &CreateRefundResponse{RawResponse: *raw}
	if raw.Object != nil {
		response.Success = firstBool(raw.Object, "success", "Success")
		response.Message = firstString(raw.Object, "message", "Message")
		response.RequestID = firstString(raw.Object, "requestId", "RequestId", "requestID")
	}
	return response, nil
}

func (c *Client) GetRefundState(ctx context.Context, requestID string) (*RefundStateResponse, error) {
	requestID = strings.TrimSpace(requestID)
	if requestID == "" {
		return nil, errors.New("request id is required")
	}

	params := make(url.Values)
	params.Set("id", requestID)
	endpoint := c.refundBaseURL + refundStatePath + "?" + params.Encode()

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("create refund state request: %w", err)
	}
	httpReq.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("send refund state request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read refund state response body: %w", err)
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, &APIError{StatusCode: resp.StatusCode, Body: strings.TrimSpace(string(body))}
	}

	raw, err := parseRawResponse(body)
	if err != nil {
		return nil, err
	}

	response := &RefundStateResponse{RawResponse: *raw}
	if raw.Object != nil {
		response.RequestID = firstString(raw.Object, "requestId", "RequestId", "requestID")
		if amount, ok := firstFloat(raw.Object, "amount", "Amount"); ok {
			response.Amount = &amount
		}
		label := strings.TrimSpace(firstString(raw.Object, "label", "Label"))
		if label != "" {
			response.Label = RefundStatus(strings.ToLower(label))
		}
		response.Message = firstString(raw.Object, "message", "Message")
	}

	return response, nil
}

func (r CreateRefundRequest) validate() error {
	if strings.TrimSpace(r.OpKey) == "" {
		return errors.New("op key is required")
	}
	if r.RefundSum != nil && *r.RefundSum <= 0 {
		return errors.New("refund sum must be greater than zero")
	}
	if len(r.InvoiceItems) > 100 {
		return errors.New("invoice items cannot contain more than 100 items")
	}
	return nil
}

func (c *Client) doRefundJWTRequest(ctx context.Context, path string, payload any) (*RawResponse, error) {
	token, err := c.createRefundToken(payload)
	if err != nil {
		return nil, err
	}
	body, err := json.Marshal(token)
	if err != nil {
		return nil, fmt.Errorf("marshal refund jwt body: %w", err)
	}

	endpoint := c.refundBaseURL + path
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(string(body)))
	if err != nil {
		return nil, fmt.Errorf("create refund request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send refund request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read refund response body: %w", err)
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, &APIError{StatusCode: resp.StatusCode, Body: strings.TrimSpace(string(respBody))}
	}
	return parseRawResponse(respBody)
}

func (c *Client) createRefundToken(payload any) (string, error) {
	algHeader, hashFactory, err := refundSignerForAlgorithm(c.algorithm)
	if err != nil {
		return "", err
	}

	header := jwtHeader{
		Typ: "JWT",
		Alg: algHeader,
	}
	headerBytes, err := json.Marshal(header)
	if err != nil {
		return "", fmt.Errorf("marshal refund jwt header: %w", err)
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshal refund jwt payload: %w", err)
	}

	headerEncoded := base64.RawURLEncoding.EncodeToString(headerBytes)
	payloadEncoded := base64.RawURLEncoding.EncodeToString(payloadBytes)
	signingInput := headerEncoded + "." + payloadEncoded

	mac := hmac.New(hashFactory, []byte(c.password3))
	_, _ = mac.Write([]byte(signingInput))
	signature := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))

	return signingInput + "." + signature, nil
}

func refundSignerForAlgorithm(algorithm SignatureAlgorithm) (headerAlgorithm string, factory func() hash.Hash, err error) {
	switch strings.ToUpper(string(algorithm)) {
	case "SHA512", "HS512":
		return "HS512", sha512.New, nil
	case "SHA384", "HS384":
		return "HS384", sha512.New384, nil
	case "SHA256", "HS256", "MD5", "RIPEMD160", "SHA1", "HS1", "":
		return "HS256", sha256.New, nil
	default:
		return "", nil, fmt.Errorf("unsupported refund signature algorithm: %q", algorithm)
	}
}

func firstBool(object map[string]any, keys ...string) bool {
	for _, key := range keys {
		value, ok := object[key]
		if !ok || value == nil {
			continue
		}
		switch typed := value.(type) {
		case bool:
			return typed
		case string:
			return strings.EqualFold(strings.TrimSpace(typed), "true")
		}
	}
	return false
}
