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
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestBuildConfirmAndCancelPaymentFormValues_SignatureProfiles(t *testing.T) {
	client := mustClient(t, Config{
		MerchantLogin: "merchant",
		Password1:     "password1",
	})

	confirmValues, err := client.BuildConfirmPaymentFormValues(ConfirmPaymentRequest{
		InvoiceID: 15,
		OutSum:    10,
		Shp: map[string]string{
			"order": "25",
		},
	})
	if err != nil {
		t.Fatalf("build confirm values: %v", err)
	}
	if confirmValues.Get("InvoiceID") != "15" {
		t.Fatalf("unexpected InvoiceID: %q", confirmValues.Get("InvoiceID"))
	}
	expectedConfirm := md5hex("merchant:10.00:15:password1:Shp_order=25")
	if confirmValues.Get("SignatureValue") != expectedConfirm {
		t.Fatalf("unexpected confirm signature: got=%q want=%q", confirmValues.Get("SignatureValue"), expectedConfirm)
	}

	cancelOutSum := 10.0
	cancelValues, err := client.BuildCancelPaymentFormValues(CancelPaymentRequest{
		InvoiceID: 15,
		OutSum:    &cancelOutSum,
		Shp: map[string]string{
			"order": "25",
		},
	})
	if err != nil {
		t.Fatalf("build cancel values: %v", err)
	}
	if cancelValues.Get("OutSum") != "10.00" {
		t.Fatalf("unexpected cancel OutSum: %q", cancelValues.Get("OutSum"))
	}
	expectedCancel := md5hex("merchant::15:password1:Shp_order=25")
	if cancelValues.Get("SignatureValue") != expectedCancel {
		t.Fatalf("unexpected cancel signature: got=%q want=%q", cancelValues.Get("SignatureValue"), expectedCancel)
	}
}

func TestBuildCoFPaymentFormValues_RequiresTokenAndUsesCoFProfile(t *testing.T) {
	client := mustClient(t, Config{
		MerchantLogin: "merchant",
		Password1:     "password1",
	})

	_, err := client.BuildCoFPaymentFormValues(InitPaymentRequest{
		OutSum: 10,
	})
	if err == nil {
		t.Fatal("expected error when token is missing")
	}

	resultURL2 := "https://merchant.test/result2"
	token := "token-1"
	values, err := client.BuildCoFPaymentFormValues(InitPaymentRequest{
		OutSum:     10,
		ResultURL2: &resultURL2,
		Token:      &token,
	})
	if err != nil {
		t.Fatalf("build cof values: %v", err)
	}

	expectedSignature := md5hex("merchant:10.00::https://merchant.test/result2:token-1:password1")
	if values.Get("SignatureValue") != expectedSignature {
		t.Fatalf("unexpected cof signature: got=%q want=%q", values.Get("SignatureValue"), expectedSignature)
	}
}

func TestResultAcknowledgement(t *testing.T) {
	value, err := ResultAcknowledgement("15")
	if err != nil {
		t.Fatalf("result acknowledgement: %v", err)
	}
	if value != "OK15" {
		t.Fatalf("unexpected result acknowledgement: %q", value)
	}

	if _, err := ResultAcknowledgement(""); err == nil {
		t.Fatal("expected validation error for empty inv id")
	}
}

func TestXMLClient_GetCurrenciesAndOpStateExt(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/GetCurrencies":
			if r.URL.Query().Get("MerchantLogin") != "merchant" {
				t.Fatalf("unexpected merchant login: %q", r.URL.Query().Get("MerchantLogin"))
			}
			_, _ = w.Write([]byte(`<?xml version="1.0" encoding="utf-8"?><CurrenciesList><Result><Code>0</Code></Result><Groups><Group Code="BankCard" Description="Cards"><Items><Currency Label="BankCardPSR" Alias="BankCard" Name="Bank card" MinValue="1" MaxValue="1000"/></Items></Group></Groups></CurrenciesList>`))
		case "/OpStateExt":
			if r.URL.Query().Get("InvoiceID") != "12" {
				t.Fatalf("unexpected invoice id: %q", r.URL.Query().Get("InvoiceID"))
			}
			expectedSignature := md5hex("merchant:12:password2")
			if !strings.EqualFold(r.URL.Query().Get("Signature"), expectedSignature) {
				t.Fatalf("unexpected signature: got=%q want=%q", r.URL.Query().Get("Signature"), expectedSignature)
			}
			_, _ = w.Write([]byte(`<?xml version="1.0" encoding="utf-8"?><OperationStateResponse><Result><Code>0</Code></Result><State><Code>100</Code><RequestDate>2026-01-01T00:00:00+03:00</RequestDate><StateDate>2026-01-01T00:00:00+03:00</StateDate></State><Info><OutSum>10.00</OutSum><OpKey>op-key</OpKey></Info></OperationStateResponse>`))
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	client := mustClient(t, Config{
		MerchantLogin: "merchant",
		Password1:     "password1",
		Password2:     "password2",
		XMLBaseURL:    server.URL,
	})

	lang := CultureRu
	currencies, err := client.GetCurrencies(context.Background(), &lang)
	if err != nil {
		t.Fatalf("get currencies: %v", err)
	}
	if len(currencies.Groups) != 1 || len(currencies.Groups[0].Items) != 1 {
		t.Fatalf("unexpected currencies payload: %+v", currencies)
	}

	state, err := client.OpStateExt(context.Background(), 12)
	if err != nil {
		t.Fatalf("op state ext: %v", err)
	}
	if state.State.Code != 100 {
		t.Fatalf("unexpected state code: %d", state.State.Code)
	}
}

func TestRefundAPI_CreateAndGetState(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/Refund/Create":
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
			headerJSON, err := base64.RawURLEncoding.DecodeString(parts[0])
			if err != nil {
				t.Fatalf("decode header: %v", err)
			}
			var header map[string]any
			if err := json.Unmarshal(headerJSON, &header); err != nil {
				t.Fatalf("unmarshal header: %v", err)
			}
			if header["alg"] != "HS256" {
				t.Fatalf("unexpected alg: %v", header["alg"])
			}

			mac := hmac.New(sha256.New, []byte("password3"))
			_, _ = mac.Write([]byte(parts[0] + "." + parts[1]))
			expectedSignature := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
			if parts[2] != expectedSignature {
				t.Fatalf("unexpected signature: got=%q want=%q", parts[2], expectedSignature)
			}

			_, _ = w.Write([]byte(`{"success":true,"requestId":"rid-1"}`))
		case "/Refund/GetState":
			if r.URL.Query().Get("id") != "rid-1" {
				t.Fatalf("unexpected request id: %q", r.URL.Query().Get("id"))
			}
			_, _ = w.Write([]byte(`{"requestId":"rid-1","amount":1.000000,"label":"finished"}`))
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	client := mustClient(t, Config{
		MerchantLogin: "merchant",
		Password1:     "password1",
		Password3:     "password3",
		RefundBaseURL: server.URL,
	})

	createResp, err := client.CreateRefund(context.Background(), CreateRefundRequest{
		OpKey: "op-key-1",
	})
	if err != nil {
		t.Fatalf("create refund: %v", err)
	}
	if !createResp.Success || createResp.RequestID != "rid-1" {
		t.Fatalf("unexpected create refund response: %+v", createResp)
	}

	stateResp, err := client.GetRefundState(context.Background(), "rid-1")
	if err != nil {
		t.Fatalf("get refund state: %v", err)
	}
	if stateResp.Label != RefundStatusFinished {
		t.Fatalf("unexpected refund label: %q", stateResp.Label)
	}
}

func TestBuildSplitPaymentFormValues_SignsInvoiceJSON(t *testing.T) {
	client := mustClient(t, Config{
		MerchantLogin: "merchant",
		Password1:     "password1",
	})

	values, err := client.BuildSplitPaymentFormValues(SplitPaymentInvoice{
		OutAmount: 700,
		Merchant: SplitMasterMerchant{
			ID: "master-shop",
		},
		SplitMerchants: []SplitMerchant{
			{
				ID:     "master-shop",
				Amount: Amount(50000),
			},
			{
				ID:     "partner-shop",
				Amount: Amount(20000),
			},
		},
	})
	if err != nil {
		t.Fatalf("build split payment values: %v", err)
	}

	invoiceJSON := values.Get("Invoice")
	if invoiceJSON == "" {
		t.Fatal("expected Invoice to be set")
	}
	expectedSignature := md5hex(invoiceJSON + ":password1")
	if values.Get("Signature") != expectedSignature {
		t.Fatalf("unexpected split signature: got=%q want=%q", values.Get("Signature"), expectedSignature)
	}
}

func TestBuildPaymentFormValues_ReceiptValidationParity(t *testing.T) {
	client := mustClient(t, Config{
		MerchantLogin: "merchant",
		Password1:     "password1",
	})

	fullPayment := PaymentMethodFullPayment
	lotteryPrize := PaymentObjectLotteryPrize
	_, err := client.BuildPaymentFormValues(InitPaymentRequest{
		OutSum: 10,
		Receipt: &Receipt{
			Items: []*ReceiptItem{
				{
					Name:          "Subscription",
					Quantity:      Quantity3(1000),
					Sum:           Price8x2(1000),
					Tax:           TaxRateVat22,
					PaymentMethod: &fullPayment,
					PaymentObject: &lotteryPrize,
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("expected supported receipt fields, got error: %v", err)
	}
}
