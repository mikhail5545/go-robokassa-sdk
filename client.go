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
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	defaultBaseURL       = "https://services.robokassa.ru/InvoiceServiceWebApi"
	defaultRefundBaseURL = "https://services.robokassa.ru/RefundService"
	defaultXMLBaseURL    = "https://auth.robokassa.ru/Merchant/WebService/Service.asmx"
)

type SignatureAlgorithm string

const (
	SignatureAlgorithmMD5       SignatureAlgorithm = "MD5"
	SignatureAlgorithmRIPEMD160 SignatureAlgorithm = "RIPEMD160"
	SignatureAlgorithmSHA1      SignatureAlgorithm = "SHA1"
	SignatureAlgorithmHS1       SignatureAlgorithm = "HS1"
	SignatureAlgorithmSHA256    SignatureAlgorithm = "SHA256"
	SignatureAlgorithmHS256     SignatureAlgorithm = "HS256"
	SignatureAlgorithmSHA384    SignatureAlgorithm = "SHA384"
	SignatureAlgorithmHS384     SignatureAlgorithm = "HS384"
	SignatureAlgorithmSHA512    SignatureAlgorithm = "SHA512"
	SignatureAlgorithmHS512     SignatureAlgorithm = "HS512"
)

type Config struct {
	MerchantLogin      string
	Password1          string
	Password2          string
	Password3          string
	SignatureAlgorithm SignatureAlgorithm
	BaseURL            string
	RefundBaseURL      string
	XMLBaseURL         string
	HTTPClient         *http.Client
}

type Client struct {
	merchantLogin string
	password1     string
	password2     string
	password3     string
	algorithm     SignatureAlgorithm
	baseURL       string
	refundBaseURL string
	xmlBaseURL    string
	httpClient    *http.Client
}

type APIError struct {
	StatusCode int
	Body       string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("robokassa api error: status=%d body=%q", e.StatusCode, e.Body)
}

func NewClient(cfg Config) (*Client, error) {
	if strings.TrimSpace(cfg.MerchantLogin) == "" {
		return nil, errors.New("merchant login is required")
	}
	if strings.TrimSpace(cfg.Password1) == "" {
		return nil, errors.New("password1 is required")
	}

	algorithm := cfg.SignatureAlgorithm
	if algorithm == "" {
		algorithm = SignatureAlgorithmMD5
	}
	if _, err := signerForAlgorithm(algorithm); err != nil {
		return nil, err
	}

	baseURL := strings.TrimRight(strings.TrimSpace(cfg.BaseURL), "/")
	if baseURL == "" {
		baseURL = defaultBaseURL
	}
	refundBaseURL := strings.TrimRight(strings.TrimSpace(cfg.RefundBaseURL), "/")
	if refundBaseURL == "" {
		refundBaseURL = defaultRefundBaseURL
	}
	xmlBaseURL := strings.TrimRight(strings.TrimSpace(cfg.XMLBaseURL), "/")
	if xmlBaseURL == "" {
		xmlBaseURL = defaultXMLBaseURL
	}

	httpClient := cfg.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 15 * time.Second}
	}

	return &Client{
		merchantLogin: cfg.MerchantLogin,
		password1:     cfg.Password1,
		password2:     cfg.Password2,
		password3:     cfg.Password3,
		algorithm:     algorithm,
		baseURL:       baseURL,
		refundBaseURL: refundBaseURL,
		xmlBaseURL:    xmlBaseURL,
		httpClient:    httpClient,
	}, nil
}

func (c *Client) doJWTRequest(ctx context.Context, path string, payload any) (*RawResponse, error) {
	token, err := c.createToken(payload)
	if err != nil {
		return nil, err
	}

	body, err := json.Marshal(token)
	if err != nil {
		return nil, fmt.Errorf("marshal jwt body: %w", err)
	}

	url := c.baseURL + path
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, strings.NewReader(string(body)))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, &APIError{StatusCode: resp.StatusCode, Body: strings.TrimSpace(string(respBody))}
	}

	return parseRawResponse(respBody)
}
