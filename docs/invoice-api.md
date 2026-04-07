# Invoice API

## Purpose

Invoice API helpers cover the invoice lifecycle in Robokassa:

- create payment invoice links
- deactivate existing invoice links
- query invoice list/history
- parse list responses into typed structures

## Endpoints covered

- `POST /api/CreateInvoice`
- `POST /api/DeactivateInvoice`
- `POST /api/GetInvoiceInformationList`

These are called through the configured `baseURL` (default: `https://services.robokassa.ru/InvoiceServiceWebApi`).

## Prerequisites

Create a client with merchant credentials:

```go
client, err := robokassa.NewClient(
    "merchant-login",
    "password1",
    robokassa.WithSignatureAlgorithm(robokassa.SignatureAlgorithmSHA256),
)
```

Before production usage, align with Robokassa quick-start flow:

1. Configure `ResultURL`, `SuccessURL`, `FailURL` in technical settings.
2. Run test payments with `IsTest=1` on payment interface side.
3. Confirm your `ResultURL` handler returns `OK{InvId}` after signature verification.

## Main methods

## `CreateInvoice(ctx, CreateInvoiceRequest)`

Creates an invoice and returns `CreateInvoiceResponse`.

Important request fields:

- `InvoiceType` (`one-time` or `reusable`)
- `OutSum`
- optional receipt (`InvoiceItems`)
- optional UI/redirect customization (`Culture`, `SuccessURL2Data`, `FailURL2Data`)

Notes:

- If `MerchantLogin` is empty, the client login is injected automatically.
- Response URL is resolved from both string and object-shaped payloads.

## `DeactivateInvoice(ctx, DeactivateInvoiceRequest)`

Deactivates invoice by one of:

- `EncodedId`
- `Id`
- `InvId`

At least one identifier is required.

## `GetInvoiceInformationList(ctx, GetInvoiceInformationListRequest)`

Fetches raw invoice list payload with filtering and pagination:

- page controls (`CurrentPage`, `PageSize`)
- required status and type filters
- required date range (`DateFrom`, `DateTo`)
- optional sum and alias filters

## `GetInvoiceInformationListTyped(ctx, GetInvoiceInformationListRequest)`

Calls raw list endpoint and parses payload into:

- `InvoiceInformationListResponse`
- `[]InvoiceInformation`

Use this when application code needs normalized typed fields (`Status`, `InvoiceType`, dates, amounts) instead of manual map parsing.

## Typical use cases

1. Create one-time invoice and redirect user to returned payment URL.
2. Build dashboard/reporting pages from typed invoice history.
3. Disable stale or canceled invoice links in admin/backoffice workflows.

## Example

```go
resp, err := client.CreateInvoice(ctx, robokassa.CreateInvoiceRequest{
    InvoiceType: robokassa.InvoiceTypeOneTime,
    OutSum:      199.99,
})
if err != nil {
    return err
}

typedList, err := client.GetInvoiceInformationListTyped(ctx, robokassa.GetInvoiceInformationListRequest{
    CurrentPage:     1,
    PageSize:        20,
    InvoiceStatuses: []robokassa.InvoiceStatus{robokassa.InvoiceStatusPaid},
    InvoiceTypes:    []robokassa.InvoiceType{robokassa.InvoiceTypeOneTime},
    DateFrom:        &from,
    DateTo:          &to,
})
if err != nil {
    return err
}

_ = resp.URL
_ = typedList.Invoices
```
