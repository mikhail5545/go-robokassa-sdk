# Refund API

## Purpose

Refund API helpers allow you to:

- create a refund request for an operation
- query refund request state

## Endpoints covered

- `POST /Refund/Create`
- `GET /Refund/GetState?id=<requestId>`

Base URL is configurable and defaults to:
`https://services.robokassa.ru/RefundService`

## Prerequisites

Create client with refund secret:

```go
client, err := robokassa.NewClient(
    "merchant-login",
    "password1",
    robokassa.WithPassword3("password3"),
    robokassa.WithSignatureAlgorithm(robokassa.SignatureAlgorithmHS256),
)
```

`password3` is required for refund JWT signing.

## Main methods

## `CreateRefund(ctx, CreateRefundRequest)`

Request fields:

- `OpKey` (required)
- optional `RefundSum`
- optional `InvoiceItems`

Response (`CreateRefundResponse`) includes:

- `Success`
- `Message`
- `RequestID`
- raw payload (`RawResponse`)

## `GetRefundState(ctx, requestID)`

Returns `RefundStateResponse`:

- `RequestID`
- optional `Amount`
- `Label` (`finished`, `processing`, `canceled`)
- `Message`
- raw payload (`RawResponse`)

## Security/signing notes

- Refund create uses JWT with HMAC signature.
- Signature algorithm follows client configuration (`MD5`/`SHA*` families mapped in SDK).
- Secret for refund token is `password3` (not `password1/password2`).

## Typical use cases

1. Start full/partial refund by `OpKey` after order cancellation.
2. Poll refund state in backoffice until operation reaches terminal status.
3. Reconcile finance records with refund request id and resulting amount label.

## Example

```go
createResp, err := client.CreateRefund(ctx, robokassa.CreateRefundRequest{
    OpKey: "operation-key",
})
if err != nil {
    return err
}

stateResp, err := client.GetRefundState(ctx, createResp.RequestID)
if err != nil {
    return err
}

_ = stateResp.Label
```
