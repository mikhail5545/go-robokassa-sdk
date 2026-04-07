# XML interfaces

## Purpose

XML interface helpers provide access to Robokassa XML methods for:

- available currencies/payment groups
- operation state details for an invoice

## Endpoints covered

- `.../Service.asmx/GetCurrencies`
- `.../Service.asmx/OpStateExt`

Base URL is configurable and defaults to:
`https://auth.robokassa.ru/Merchant/WebService/Service.asmx`

## Prerequisites

```go
client, err := robokassa.NewClient(
    "merchant-login",
    "password1",
    robokassa.WithPassword2("password2"), // required for OpStateExt
)
```

`GetCurrencies` requires merchant login.  
`OpStateExt` requires merchant login + `password2` (for signature).

## Main methods

## `GetCurrencies(ctx, language)`

Returns `GetCurrenciesResponse` with:

- `Result` (`Code`, `Description`)
- currency groups and items (`Groups > Group > Items`)

`language` supports `ru` and `en`.

## `OpStateExt(ctx, invoiceID)`

Returns `OpStateExtResponse` with:

- status/result metadata
- operation info (`OutSum`, `OpKey`, payment method, etc.)
- user fields

`invoiceID` must be greater than zero.

## Error behavior

- HTTP non-2xx returns `APIError`.
- XML `Result.Code != 0` returns `XMLAPIError`.
- malformed XML payloads return unmarshal errors.

## Typical use cases

1. Populate payment method/currency selectors in admin or storefront UIs.
2. Query operation state and obtain `OpKey` before refund operations.
3. Build reconciliation jobs that compare internal order state with Robokassa operation state.

## Example

```go
lang := robokassa.CultureRu
currencies, err := client.GetCurrencies(ctx, &lang)
if err != nil {
    return err
}

state, err := client.OpStateExt(ctx, 12345)
if err != nil {
    return err
}

_ = currencies.Groups
_ = state.Info.OpKey
```
