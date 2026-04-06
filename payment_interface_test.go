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
	"crypto"
	"crypto/md5"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"math/big"
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestBuildPaymentFormValues_SignsRequestWithModifiersAndShp(t *testing.T) {
	client := mustClient(t, Config{
		MerchantLogin:      "merchant",
		Password1:          "password1",
		SignatureAlgorithm: SignatureAlgorithmMD5,
	})

	invID := int64(12)
	desc := "Order #12"
	culture := CultureRu
	successMethod := "get"
	failMethod := "post"
	successURL := "https://merchant.test/success"
	failURL := "https://merchant.test/fail"
	resultURL2 := "https://merchant.test/result2"
	token := "token_123"

	req := InitPaymentRequest{
		OutSum:      990,
		InvID:       &invID,
		Description: &desc,
		Culture:     &culture,
		IsTest:      true,
		Receipt: &Receipt{
			Items: []*ReceiptItem{
				{
					Name:     "Product",
					Quantity: Quantity3(1000),
					Sum:      Price8x2(99000),
					Tax:      TaxRateVat20,
				},
			},
		},
		StepByStep:        true,
		ResultURL2:        &resultURL2,
		SuccessURL2:       &successURL,
		SuccessURL2Method: &successMethod,
		FailURL2:          &failURL,
		FailURL2Method:    &failMethod,
		Token:             &token,
		PaymentMethods:    []string{"BankCard", "SBP"},
		Shp: map[string]string{
			"order":     "25",
			"Shp_login": "Vasya",
		},
	}

	values, err := client.BuildPaymentFormValues(req)
	if err != nil {
		t.Fatalf("build payment form values: %v", err)
	}

	if values.Get("MerchantLogin") != "merchant" {
		t.Fatalf("unexpected MerchantLogin: %q", values.Get("MerchantLogin"))
	}
	if values.Get("OutSum") != "990.00" {
		t.Fatalf("unexpected OutSum: %q", values.Get("OutSum"))
	}
	if values.Get("IsTest") != "1" {
		t.Fatalf("unexpected IsTest: %q", values.Get("IsTest"))
	}
	if values.Get("SuccessUrl2Method") != "GET" {
		t.Fatalf("unexpected SuccessUrl2Method: %q", values.Get("SuccessUrl2Method"))
	}
	if values.Get("FailUrl2Method") != "POST" {
		t.Fatalf("unexpected FailUrl2Method: %q", values.Get("FailUrl2Method"))
	}
	if values.Get("Shp_order") != "25" || values.Get("Shp_login") != "Vasya" {
		t.Fatalf("unexpected shp values: %+v", values)
	}

	baseString, err := client.PaymentSignatureBaseString(req)
	if err != nil {
		t.Fatalf("signature base string: %v", err)
	}
	expectedSignature := md5hex(baseString)
	if values.Get("SignatureValue") != expectedSignature {
		t.Fatalf("unexpected signature: got=%q want=%q", values.Get("SignatureValue"), expectedSignature)
	}

	link, err := client.BuildPaymentURL(req)
	if err != nil {
		t.Fatalf("build payment url: %v", err)
	}
	if !strings.HasPrefix(link, PaymentGatewayURL+"?") {
		t.Fatalf("unexpected payment url prefix: %q", link)
	}
}

func TestResultAndSuccessSignature_Verification(t *testing.T) {
	client := mustClient(t, Config{
		MerchantLogin:      "merchant",
		Password1:          "password1",
		Password2:          "password2",
		SignatureAlgorithm: SignatureAlgorithmMD5,
	})

	shp := map[string]string{
		"Shp_oplata": "1",
		"login":      "Vasya",
	}
	outSum := "100.000000"
	invID := "450009"

	expectedResult := md5hex("100.000000:450009:password2:Shp_login=Vasya:Shp_oplata=1")
	resultSignature, err := client.ResultSignature(outSum, invID, shp)
	if err != nil {
		t.Fatalf("result signature: %v", err)
	}
	if resultSignature != expectedResult {
		t.Fatalf("unexpected result signature: got=%q want=%q", resultSignature, expectedResult)
	}

	ok, err := client.VerifyResultSignature(outSum, invID, strings.ToUpper(resultSignature), shp)
	if err != nil {
		t.Fatalf("verify result signature: %v", err)
	}
	if !ok {
		t.Fatal("expected valid result signature")
	}

	successSignature, err := client.SuccessSignature(outSum, invID, shp)
	if err != nil {
		t.Fatalf("success signature: %v", err)
	}
	ok, err = client.VerifySuccessSignature(outSum, invID, successSignature, shp)
	if err != nil {
		t.Fatalf("verify success signature: %v", err)
	}
	if !ok {
		t.Fatal("expected valid success signature")
	}

	ok, err = client.VerifySuccessSignature(outSum, invID, "bad-signature", shp)
	if err != nil {
		t.Fatalf("verify success signature with bad value: %v", err)
	}
	if ok {
		t.Fatal("expected invalid success signature")
	}
}

func TestParseCallbackNotification(t *testing.T) {
	values := url.Values{
		"OutSum":         {"100.000000"},
		"InvID":          {"5"},
		"SignatureValue": {"abc123"},
		"EMail":          {"buyer@example.com"},
		"Shp_order":      {"25"},
		"Shp_user":       {"42"},
	}

	callback := ParseCallbackNotification(values)
	if callback.OutSum != "100.000000" {
		t.Fatalf("unexpected OutSum: %q", callback.OutSum)
	}
	if callback.InvID != "5" {
		t.Fatalf("unexpected InvID: %q", callback.InvID)
	}
	if callback.Email != "buyer@example.com" {
		t.Fatalf("unexpected email: %q", callback.Email)
	}
	if callback.Shp["Shp_order"] != "25" || callback.Shp["Shp_user"] != "42" {
		t.Fatalf("unexpected shp map: %+v", callback.Shp)
	}
}

func TestParseAndVerifyResultURL2JWS(t *testing.T) {
	privateKey, certificateDER := newTestCertificate(t)
	certificatePEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certificateDER,
	})

	payload := ResultURL2Notification{
		Header: ResultURL2PayloadHeader{
			Type:      "PaymentStateNotification",
			Version:   "1.0.0",
			Timestamp: "1691186412",
		},
		Data: ResultURL2PayloadData{
			Shop:          "merchant",
			OpKey:         "op-key",
			InvID:         "123",
			PaymentMethod: "BankCard",
			IncSum:        "10.00",
			State:         "OK",
		},
	}
	token := signRS256JWS(t, privateKey, payload)

	parsed, err := ParseResultURL2JWS(token)
	if err != nil {
		t.Fatalf("parse jws: %v", err)
	}
	if parsed.TokenHeader.Alg != "RS256" {
		t.Fatalf("unexpected alg: %q", parsed.TokenHeader.Alg)
	}
	if parsed.Payload.Data.InvID != "123" {
		t.Fatalf("unexpected inv id: %q", parsed.Payload.Data.InvID)
	}

	if err := VerifyResultURL2JWS(token, certificateDER); err != nil {
		t.Fatalf("verify jws by der: %v", err)
	}
	if err := VerifyResultURL2JWS(token, certificatePEM); err != nil {
		t.Fatalf("verify jws by pem: %v", err)
	}

	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		t.Fatalf("unexpected jws format: %q", token)
	}
	if len(parts[2]) == 0 {
		t.Fatalf("empty jws signature in token: %q", token)
	}

	replacement := "A"
	if strings.HasPrefix(parts[2], "A") {
		replacement = "B"
	}
	tampered := parts[0] + "." + parts[1] + "." + replacement + parts[2][1:]
	if err := VerifyResultURL2JWS(tampered, certificatePEM); err == nil {
		t.Fatal("expected tampered jws verification error")
	}
}

func TestResultSignature_RequiresPassword2(t *testing.T) {
	client := mustClient(t, Config{
		MerchantLogin: "merchant",
		Password1:     "password1",
	})

	_, err := client.ResultSignature("10.00", "1", nil)
	if err == nil {
		t.Fatal("expected error when password2 is not configured")
	}
}

func md5hex(input string) string {
	sum := md5.Sum([]byte(input))
	return strings.ToLower(hexEncode(sum[:]))
}

func hexEncode(data []byte) string {
	const hex = "0123456789abcdef"
	out := make([]byte, len(data)*2)
	for i, b := range data {
		out[i*2] = hex[b>>4]
		out[i*2+1] = hex[b&0x0f]
	}
	return string(out)
}

func newTestCertificate(t *testing.T) (*rsa.PrivateKey, []byte) {
	t.Helper()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate rsa key: %v", err)
	}

	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}

	certificateDER, err := x509.CreateCertificate(rand.Reader, template, template, &privateKey.PublicKey, privateKey)
	if err != nil {
		t.Fatalf("create certificate: %v", err)
	}

	return privateKey, certificateDER
}

func signRS256JWS(t *testing.T, privateKey *rsa.PrivateKey, payload ResultURL2Notification) string {
	t.Helper()

	header := ResultURL2TokenHeader{
		Typ: "JWT",
		Alg: "RS256",
	}
	headerJSON, err := json.Marshal(header)
	if err != nil {
		t.Fatalf("marshal jws header: %v", err)
	}
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal jws payload: %v", err)
	}

	encodedHeader := base64.RawURLEncoding.EncodeToString(headerJSON)
	encodedPayload := base64.RawURLEncoding.EncodeToString(payloadJSON)
	signingInput := encodedHeader + "." + encodedPayload

	digest := sha256.Sum256([]byte(signingInput))
	signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, digest[:])
	if err != nil {
		t.Fatalf("sign jws: %v", err)
	}

	return signingInput + "." + base64.RawURLEncoding.EncodeToString(signature)
}
