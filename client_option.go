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
	"net/http"
	"strings"
	"time"
)

// ClientOption mutates client configuration during NewClient initialization.
type ClientOption func(*Client) error

// WithPassword2 sets password #2 used for ResultURL signature verification and XML OpStateExt.
func WithPassword2(password2 string) ClientOption {
	return func(c *Client) error {
		c.password2 = password2
		return nil
	}
}

// WithPassword3 sets password #3 used for Refund API JWT signing.
func WithPassword3(password3 string) ClientOption {
	return func(c *Client) error {
		c.password3 = password3
		return nil
	}
}

// WithSignatureAlgorithm sets signature algorithm for hash/JWT operations supported by Robokassa.
func WithSignatureAlgorithm(alg SignatureAlgorithm) ClientOption {
	return func(c *Client) error {
		if _, err := signerForAlgorithm(alg); err != nil {
			return err
		}
		c.algorithm = alg
		return nil
	}
}

// WithBaseURL overrides Invoice API base URL.
func WithBaseURL(baseURL string) ClientOption {
	return func(c *Client) error {
		url := strings.TrimRight(strings.TrimSpace(baseURL), "/")
		if url == "" {
			return errors.New("base url is empty")
		}
		c.baseURL = url
		return nil
	}
}

// WithRefundBaseURL overrides Refund API base URL.
func WithRefundBaseURL(baseURL string) ClientOption {
	return func(c *Client) error {
		url := strings.TrimRight(strings.TrimSpace(baseURL), "/")
		if url == "" {
			return errors.New("refund base url is empty")
		}
		c.refundBaseURL = url
		return nil
	}
}

// WithXMLBaseURL overrides XML API base URL.
func WithXMLBaseURL(baseURL string) ClientOption {
	return func(c *Client) error {
		url := strings.TrimRight(strings.TrimSpace(baseURL), "/")
		if url == "" {
			return errors.New("xml base url is empty")
		}
		c.xmlBaseURL = url
		return nil
	}
}

// WithHTTPClient replaces the internal HTTP client used for all outbound requests.
func WithHTTPClient(httpClient *http.Client) ClientOption {
	return func(c *Client) error {
		if httpClient == nil {
			return errors.New("http client is nil")
		}
		c.httpClient = httpClient
		return nil
	}
}

// WithHTTPClientTimeout overrides timeout on current HTTP client.
func WithHTTPClientTimeout(timeout time.Duration) ClientOption {
	return func(c *Client) error {
		if c.httpClient == nil {
			return errors.New("http client is nil")
		}
		c.httpClient.Timeout = timeout
		return nil
	}
}
