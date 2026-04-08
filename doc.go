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

// Package robokassa is a comprehensive SDK for integrating Robokassa API into your systems.
//
// # Basic usage
//
// To start using client starts with creating a new Client with your merchant login and passwords (up to 3), but password1 is required. To do so, call NewClient:
//
//	client, err := robokassa.NewClient(
//		"your-merchant-login",
//		"your-password-1",
//		robokassa.WithPassword2("your-password-2"), // required for ResultURL verification
//		robokassa.WithPassword3("your-password-3"), // required for Refund API
//		robokassa.WithSignatureAlgorithm(robokassa.SignatureAlgorithmSHA256),
//	)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	resp, err := client.CreateInvoice(context.Background(), robokassa.CreateInvoiceRequest{
//		InvoiceType: robokassa.InvoiceTypeOneTime,
//		OutSum:      100.50,
//	})
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	fmt.Println("Payment URL:", resp.URL)
//
// Build pay-interface link
//
//	invID := int64(1001)
//	link, err := client.BuildPaymentURL(robokassa.InitPaymentRequest{
//		OutSum: 99.90,
//		InvID:  &invID,
//		Shp: map[string]string{
//		"order": "1001",
//		},
//		IsTest: true,
//	})
//	if err != nil {
//		log.Fatal(err)
//	}
//	fmt.Println(link)
//
// Hold confirm/cancel helpers
//
//	confirmLink, err := client.BuildConfirmPaymentURL(robokassa.ConfirmPaymentRequest{
//		InvoiceID: 1001,
//		OutSum:    99.90,
//	})
//	if err != nil {
//		log.Fatal(err)
//	}
//	fmt.Println(confirmLink)
//
// # Error Handling
//
// When you will try to verify callback/notification signature, you can encounter different errors.
// Here is the quick breakdown:
//
//   - ErrInvalidCallbackSignature  - indicates that callback signature comparison failed.
//   - ErrUnsupportedCallbackSignatureKind - indicates unsupported callback verification mode.
//   - ErrResultURL2InvalidToken - indicates malformed JWS token structure.
//   - ErrResultURL2InvalidHeader - indicates malformed JWS header JSON.
//   - ErrResultURL2InvalidPayload - indicates malformed JWS payload JSON.
//   - ErrResultURL2UnsupportedAlgorithm - indicates unsupported JWS signature algorithm.
//   - ErrResultURL2InvalidCertificate - indicates invalid certificate format/data.
//   - ErrResultURL2InvalidCertificateKey - indicates that certificate does not contain RSA public key.
//   - ErrResultURL2SignatureVerification - indicates failed RSA signature verification.
package robokassa
