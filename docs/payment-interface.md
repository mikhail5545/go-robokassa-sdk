# Payment Interface

## Purpose

Payment Interface helpers generate signed Robokassa payment URLs/form fields and provide callback signature verification utilities.

Use this API when your backend wants to:

- create redirect links/forms for payment
- support hold/confirm/cancel flows
- start recurring or saved-card (CoF) payments
- verify `ResultURL`/`SuccessURL` signatures
- parse and verify `ResultUrl2` JWS notifications

## Endpoints covered

- `https://auth.robokassa.ru/Merchant/Index.aspx`
- `https://auth.robokassa.ru/Merchant/Payment/Confirm`
- `https://auth.robokassa.ru/Merchant/Payment/Cancel`
- `https://auth.robokassa.ru/Merchant/Recurring`
- `https://auth.robokassa.ru/Merchant/Payment/CoFPayment`

## Main methods

## Payment initialization

- `BuildPaymentFormValues(InitPaymentRequest)`
- `BuildPaymentURL(InitPaymentRequest)`
- `CalculatePaymentSignature(InitPaymentRequest)`
- `PaymentSignatureBaseString(InitPaymentRequest)`

`InitPaymentRequest` supports:

- amount (`OutSum` / `OutSumText`)
- invoice id and description
- receipt
- test mode
- extra callbacks (`ResultUrl2`, `SuccessUrl2`, `FailUrl2`)
- custom user params (`Shp_*`)
- optional payment method filtering

## Hold and recurring flows

- `BuildConfirmPaymentFormValues` / `BuildConfirmPaymentURL`
- `BuildCancelPaymentFormValues` / `BuildCancelPaymentURL`
- `BuildRecurringPaymentFormValues` / `BuildRecurringPaymentURL`
- `BuildCoFPaymentFormValues` / `BuildCoFPaymentURL`

Use these when you need two-stage payments, child recurring payments, or charging a saved card token.

## Callback and signature helpers

- `ResultSignature`, `SuccessSignature`
- `VerifyResultSignature`, `VerifySuccessSignature`
- `ParseCallbackNotification`
- `VerifyResultNotification`, `VerifySuccessNotification`
- `ResultAcknowledgement`
- `ParseResultURL2JWS`
- `VerifyResultURL2JWS`

`ResultSignature` requires `password2` in client options.

## Typical use cases

1. Build signed checkout URL and redirect customer.
2. Verify server callback signatures before marking order as paid.
3. Handle two-stage payment with explicit confirm/cancel step.
4. Validate cryptographically signed `ResultUrl2` notifications.

## Example

```go
invID := int64(1001)
values, err := client.BuildPaymentFormValues(robokassa.InitPaymentRequest{
    OutSum: 149.50,
    InvID:  &invID,
    Shp: map[string]string{
        "order": "1001",
    },
})
if err != nil {
    return err
}

ok, err := client.VerifyResultSignature(
    "149.500000",
    "1001",
    signatureValueFromCallback,
    map[string]string{"Shp_order": "1001"},
)
if err != nil || !ok {
    return fmt.Errorf("invalid result signature")
}

_ = values
```
