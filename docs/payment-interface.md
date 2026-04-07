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
- `VerifyCallbackSignature`
- `VerifyResultSignature`, `VerifySuccessSignature`
- `ParseCallbackNotification`
- `VerifyResultNotification`, `VerifySuccessNotification`
- `ResultAcknowledgement`
- `ParseResultURL2JWS`
- `VerifyResultURL2JWS`

`ResultSignature` requires `password2` in client options.

Verification methods return `error` (not `bool`). For signature mismatch, use
`errors.Is(err, robokassa.ErrInvalidCallbackSignature)`.

ResultURL2 JWS verification errors are also classified (`ErrResultURL2InvalidToken`,
`ErrResultURL2UnsupportedAlgorithm`, `ErrResultURL2InvalidCertificate`,
`ErrResultURL2SignatureVerification`).

## Quick-start checklist alignment

1. Set `ResultURL`, `SuccessURL`, `FailURL` in Robokassa technical settings.
2. Verify callback signature before changing order state.
3. Return exact acknowledgement `OK{InvId}` from `ResultURL`.
4. Use test mode (`IsTest=1`) before production launch.

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

err = client.VerifyResultSignature(
    "149.500000",
    "1001",
    signatureValueFromCallback,
    map[string]string{"Shp_order": "1001"},
)
if err != nil {
    return fmt.Errorf("invalid result signature")
}

_ = values
```
