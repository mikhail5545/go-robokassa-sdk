# go-robokassa-sdk

Go SDK for Robokassa payments that covers invoice lifecycle management, payment URL/signature building, XML service methods, refund operations, and webhook/JWS verification helpers.

The library is designed for backend services that need a typed, testable integration layer around Robokassa endpoints. It keeps endpoint-specific request validation and signature/JWT creation inside the SDK so application code can focus on business logic.

## Features

- JWT request signing (`MD5`, `RIPEMD160`, `SHA1/HS1`, `SHA256/HS256`, `SHA384/HS384`, `SHA512/HS512`)
- Create invoice links (`CreateInvoice`)
- Deactivate invoice links (`DeactivateInvoice`)
- Fetch invoice information list (`GetInvoiceInformationList`)
- Fetch typed invoice information list (`GetInvoiceInformationListTyped`)
- Build signed pay-interface URLs (`BuildPaymentURL`, `BuildPaymentFormValues`)
- Build hold/recurring/saved-card URLs (`BuildConfirmPaymentURL`, `BuildCancelPaymentURL`, `BuildRecurringPaymentURL`, `BuildCoFPaymentURL`)
- Verify webhook signatures (`ResultURL`, `SuccessURL`)
- Parse and verify `ResultUrl2` JWS callbacks
- Build `OK{InvId}` ResultURL acknowledgement (`ResultAcknowledgement`)
- XML interface support (`GetCurrencies`, `OpStateExt`)
- Refund API support (`CreateRefund`, `GetRefundState`)
- Split payments URL/form helpers (`BuildSplitPaymentURL`, `BuildSplitPaymentFormValues`)
- Merchant login auto-injection from client config

## Install

```bash
go get github.com/mikhail5545/go-robokassa-sdk
```

## Versioning policy

- Current major version `v1` uses the base module path: `github.com/mikhail5545/go-robokassa-sdk`.
- Semantic import suffixes are used only for future major versions (`/v2`, `/v3`, ...).

## Documentation

- [Invoice API guide](docs/invoice-api.md)
- [Payment Interface guide](docs/payment-interface.md)
- [XML interfaces guide](docs/xml-interfaces.md)
- [Refund API guide](docs/refund-api.md)
- [Contributing guide](CONTRIBUTING.md)

## Quick start

```go
package main

import (
	"context"
	"fmt"
	"log"

	robokassa "github.com/mikhail5545/go-robokassa-sdk"
)

func main() {
	client, err := robokassa.NewClient(robokassa.Config{
		MerchantLogin:      "your-merchant-login",
		Password1:          "your-password-1",
		Password2:          "your-password-2", // required for ResultURL verification
		Password3:          "your-password-3", // required for Refund API
		SignatureAlgorithm: robokassa.SignatureAlgorithmSHA256,
	})
	if err != nil {
		log.Fatal(err)
	}

	resp, err := client.CreateInvoice(context.Background(), robokassa.CreateInvoiceRequest{
		InvoiceType: robokassa.InvoiceTypeOneTime,
		OutSum:      100.50,
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Payment URL:", resp.URL)
}
```

### Build pay-interface link (Quick Start)

```go
invID := int64(1001)
link, err := client.BuildPaymentURL(robokassa.InitPaymentRequest{
	OutSum: 99.90,
	InvID:  &invID,
	Shp: map[string]string{
		"order": "1001",
	},
	IsTest: true,
})
if err != nil {
	log.Fatal(err)
}
fmt.Println(link)
```

### Hold confirm/cancel helpers

```go
confirmLink, err := client.BuildConfirmPaymentURL(robokassa.ConfirmPaymentRequest{
	InvoiceID: 1001,
	OutSum:    99.90,
})
if err != nil {
	log.Fatal(err)
}
fmt.Println(confirmLink)
```

### Refund API

```go
refund, err := client.CreateRefund(context.Background(), robokassa.CreateRefundRequest{
	OpKey: "operation-key-from-opstateext-or-result2",
})
if err != nil {
	log.Fatal(err)
}
fmt.Println("refund request id:", refund.RequestID)
```

### Verify ResultURL webhook

```go
ok, err := client.VerifyResultSignature(
	"100.000000",
	"1001",
	"<SignatureValue from callback>",
	map[string]string{"Shp_order": "1001"},
)
if err != nil {
	log.Fatal(err)
}
if !ok {
	log.Fatal("invalid result signature")
}
ack, err := robokassa.ResultAcknowledgement("1001")
if err != nil {
	log.Fatal(err)
}
fmt.Println(ack) // OK1001
```

### Parse and verify ResultUrl2 JWS callback

```go
parsed, err := robokassa.ParseResultURL2JWS(token)
if err != nil {
	log.Fatal(err)
}
fmt.Println(parsed.Payload.Data.State)

if err := robokassa.VerifyResultURL2JWS(token, certBytes); err != nil {
	log.Fatal(err)
}
```

### XML interfaces

```go
currencies, err := client.GetCurrencies(context.Background(), nil)
if err != nil {
	log.Fatal(err)
}
fmt.Println("currency groups:", len(currencies.Groups))
```

### Split payment helper

```go
splitURL, err := client.BuildSplitPaymentURL(robokassa.SplitPaymentInvoice{
	OutAmount: 700,
	Merchant:  robokassa.SplitMasterMerchant{ID: "master-shop"},
	SplitMerchants: []robokassa.SplitMerchant{
		{ID: "master-shop", Amount: robokassa.Amount(50000)},
		{ID: "partner-shop", Amount: robokassa.Amount(20000)},
	},
})
if err != nil {
	log.Fatal(err)
}
fmt.Println(splitURL)
```

## Endpoints covered

1. Invoice API:
    - `POST /api/CreateInvoice`
    - `POST /api/DeactivateInvoice`
    - `POST /api/GetInvoiceInformationList`
2. Payment interface:
    - `https://auth.robokassa.ru/Merchant/Index.aspx` parameter/signature helpers
    - `https://auth.robokassa.ru/Merchant/Payment/Confirm`
    - `https://auth.robokassa.ru/Merchant/Payment/Cancel`
    - `https://auth.robokassa.ru/Merchant/Recurring`
    - `https://auth.robokassa.ru/Merchant/Payment/CoFPayment`
    - `https://auth.robokassa.ru/Merchant/Payment/CreateV2` (split)
3. XML interfaces:
    - `.../Service.asmx/GetCurrencies`
    - `.../Service.asmx/OpStateExt`
4. Refund API:
    - `POST https://services.robokassa.ru/RefundService/Refund/Create`
    - `GET https://services.robokassa.ru/RefundService/Refund/GetState?id=<requestId>`

JWT for Invoice API is sent in request body as a JSON string, according to Robokassa docs.
