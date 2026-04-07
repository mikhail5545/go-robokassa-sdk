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
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestNewClient_ValidationAndDefaults(t *testing.T) {
	_, err := NewClient("", "p1")
	if err == nil {
		t.Fatal("expected error when merchant login is missing")
	}

	_, err = NewClient("merchant", "")
	if err == nil {
		t.Fatal("expected error when password1 is missing")
	}

	client, err := NewClient("merchant", "password1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client.algorithm != SignatureAlgorithmMD5 {
		t.Fatalf("unexpected default algorithm: got=%q want=%q", client.algorithm, SignatureAlgorithmMD5)
	}
	if client.baseURL != defaultBaseURL {
		t.Fatalf("unexpected default base url: got=%q want=%q", client.baseURL, defaultBaseURL)
	}
}

func TestCreateToken_UsesConfiguredAlgorithm(t *testing.T) {
	client, err := NewClient("merchant", "password1", WithSignatureAlgorithm(SignatureAlgorithmHS256))
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	token, err := client.createToken(map[string]any{"MerchantLogin": "merchant", "OutSum": 1})
	if err != nil {
		t.Fatalf("create token: %v", err)
	}

	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		t.Fatalf("unexpected token format: %q", token)
	}

	headerBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		t.Fatalf("decode header: %v", err)
	}
	var header map[string]any
	if err := json.Unmarshal(headerBytes, &header); err != nil {
		t.Fatalf("unmarshal header: %v", err)
	}
	if header["typ"] != "JWT" {
		t.Fatalf("unexpected typ: got=%v want=JWT", header["typ"])
	}
	if header["alg"] != string(SignatureAlgorithmHS256) {
		t.Fatalf("unexpected alg: got=%v want=%q", header["alg"], SignatureAlgorithmHS256)
	}

	mac := hmac.New(sha256.New, []byte("merchant:password1"))
	_, _ = mac.Write([]byte(parts[0] + "." + parts[1]))
	expectedSignature := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	if parts[2] != expectedSignature {
		t.Fatalf("unexpected signature: got=%q want=%q", parts[2], expectedSignature)
	}
}

func TestCreateInvoice_AutoInjectsMerchantLoginAndParsesURLString(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("unexpected method: got=%s want=POST", r.Method)
		}
		if r.URL.Path != invoiceCreateEndpoint {
			t.Fatalf("unexpected path: got=%s want=%s", r.URL.Path, invoiceCreateEndpoint)
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read request body: %v", err)
		}

		var token string
		if err := json.Unmarshal(body, &token); err != nil {
			t.Fatalf("request body must be json string token: %v", err)
		}

		parts := strings.Split(token, ".")
		if len(parts) != 3 {
			t.Fatalf("unexpected token format: %q", token)
		}

		payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
		if err != nil {
			t.Fatalf("decode payload: %v", err)
		}
		var payload map[string]any
		if err := json.Unmarshal(payloadBytes, &payload); err != nil {
			t.Fatalf("unmarshal payload: %v", err)
		}
		if payload["MerchantLogin"] != "merchant" {
			t.Fatalf("merchant login was not auto-injected: got=%v", payload["MerchantLogin"])
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`"https://auth.robokassa.ru/merchant/Invoice/example"`))
	})
	server := httptest.NewServer(handler)
	defer server.Close()

	client := mustClient(t, "merchant", "password1", WithBaseURL(server.URL))

	resp, err := client.CreateInvoice(context.Background(), CreateInvoiceRequest{
		InvoiceType: InvoiceTypeOneTime,
		OutSum:      150,
	})
	if err != nil {
		t.Fatalf("create invoice: %v", err)
	}
	if resp.URL != "https://auth.robokassa.ru/merchant/Invoice/example" {
		t.Fatalf("unexpected payment URL: got=%q", resp.URL)
	}
}

func TestCreateInvoice_ParsesURLFromObjectResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"InvoiceUrl":"https://auth.robokassa.ru/merchant/Invoice/object","Id":"123"}`))
	}))
	defer server.Close()

	client := mustClient(t, "merchant", "password1", WithBaseURL(server.URL))

	resp, err := client.CreateInvoice(context.Background(), CreateInvoiceRequest{
		InvoiceType: InvoiceTypeOneTime,
		OutSum:      99,
	})
	if err != nil {
		t.Fatalf("create invoice: %v", err)
	}
	if resp.URL != "https://auth.robokassa.ru/merchant/Invoice/object" {
		t.Fatalf("unexpected payment URL: got=%q", resp.URL)
	}
}

func TestDeactivateInvoice_Validation(t *testing.T) {
	client := mustClient(t, "merchant", "password1")

	_, err := client.DeactivateInvoice(context.Background(), DeactivateInvoiceRequest{})
	if err == nil {
		t.Fatal("expected validation error when no identifiers are provided")
	}
}

func TestGetInvoiceInformationList_Validation(t *testing.T) {
	client := mustClient(t, "merchant", "password1")

	_, err := client.GetInvoiceInformationList(context.Background(), GetInvoiceInformationListRequest{})
	if err == nil {
		t.Fatal("expected validation error for empty request")
	}

	now := time.Now()
	_, err = client.GetInvoiceInformationList(context.Background(), GetInvoiceInformationListRequest{
		CurrentPage:     1,
		PageSize:        10,
		InvoiceStatuses: []InvoiceStatus{InvoiceStatusPaid},
		InvoiceTypes:    []InvoiceType{InvoiceTypeOneTime},
		DateFrom:        &now,
		DateTo:          &now,
	})
	if err != nil {
		t.Fatalf("expected valid request, got error: %v", err)
	}
}

func TestCreateInvoice_PropagatesHTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"bad request"}`))
	}))
	defer server.Close()

	client := mustClient(t, "merchant", "password1", WithBaseURL(server.URL))

	_, err := client.CreateInvoice(context.Background(), CreateInvoiceRequest{
		InvoiceType: InvoiceTypeOneTime,
		OutSum:      10,
	})
	if err == nil {
		t.Fatal("expected http error")
	}

	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected APIError, got=%T (%v)", err, err)
	}
	if apiErr.StatusCode != http.StatusBadRequest {
		t.Fatalf("unexpected status: got=%d want=%d", apiErr.StatusCode, http.StatusBadRequest)
	}
}

func mustClient(t *testing.T, merchantLogin, password1 string, opt ...ClientOption) *Client {
	t.Helper()
	client, err := NewClient(merchantLogin, password1, opt...)
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	return client
}
