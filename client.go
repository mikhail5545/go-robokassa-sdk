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
	// SignatureAlgorithmMD5 uses MD5 hash (default in many Robokassa setups).
	SignatureAlgorithmMD5 SignatureAlgorithm = "MD5"
	// SignatureAlgorithmRIPEMD160 uses RIPEMD160 hash.
	SignatureAlgorithmRIPEMD160 SignatureAlgorithm = "RIPEMD160"
	// SignatureAlgorithmSHA1 uses SHA1 hash.
	SignatureAlgorithmSHA1 SignatureAlgorithm = "SHA1"
	// SignatureAlgorithmHS1 is Robokassa alias for SHA1.
	SignatureAlgorithmHS1 SignatureAlgorithm = "HS1"
	// SignatureAlgorithmSHA256 uses SHA256 hash.
	SignatureAlgorithmSHA256 SignatureAlgorithm = "SHA256"
	// SignatureAlgorithmHS256 is Robokassa alias for SHA256.
	SignatureAlgorithmHS256 SignatureAlgorithm = "HS256"
	// SignatureAlgorithmSHA384 uses SHA384 hash.
	SignatureAlgorithmSHA384 SignatureAlgorithm = "SHA384"
	// SignatureAlgorithmHS384 is Robokassa alias for SHA384.
	SignatureAlgorithmHS384 SignatureAlgorithm = "HS384"
	// SignatureAlgorithmSHA512 uses SHA512 hash.
	SignatureAlgorithmSHA512 SignatureAlgorithm = "SHA512"
	// SignatureAlgorithmHS512 is Robokassa alias for SHA512.
	SignatureAlgorithmHS512 SignatureAlgorithm = "HS512"
)

// Client is a typed SDK client for Robokassa Invoice, Payment Interface, XML and Refund APIs.
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

// APIError represents non-2xx HTTP response from Robokassa endpoints.
type APIError struct {
	StatusCode int
	Body       string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("robokassa api error: status=%d body=%q", e.StatusCode, e.Body)
}

// NewClient creates a new Robokassa client with required MerchantLogin and password #1.
//
// Additional credentials and transport settings can be configured via [ClientOption].
func NewClient(merchantLogin, password1 string, opt ...ClientOption) (*Client, error) {
	if err := validateRequiredTrimmed(merchantLogin, "merchant login is required"); err != nil {
		return nil, err
	}
	if err := validateRequiredTrimmed(password1, "password1 is required"); err != nil {
		return nil, err
	}

	client := &Client{
		merchantLogin: merchantLogin,
		password1:     password1,
		algorithm:     SignatureAlgorithmMD5,
		baseURL:       defaultBaseURL,
		refundBaseURL: defaultRefundBaseURL,
		xmlBaseURL:    defaultXMLBaseURL,
		httpClient:    &http.Client{Timeout: 15 * time.Second},
	}

	for _, o := range opt {
		if err := o(client); err != nil {
			return nil, err
		}
	}

	return client, nil
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
