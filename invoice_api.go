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
	"errors"
	"fmt"
	"strings"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	internalvalidation "github.com/mikhail5545/go-robokassa-sdk/internal/validation"
)

const (
	invoiceCreateEndpoint          = "/api/CreateInvoice"
	invoiceDeactivateEndpoint      = "/api/DeactivateInvoice"
	invoiceInformationListEndpoint = "/api/GetInvoiceInformationList"
)

type URLData struct {
	URL    string `json:"Url"`
	Method string `json:"Method,omitempty"`
}

type CreateInvoiceRequest struct {
	// Store login, required for invoice creation.
	MerchantLogin string `json:"MerchantLogin,omitempty"`
	// Type of the invoice, required for invoice creation.
	InvoiceType InvoiceType `json:"InvoiceType"`
	// Optional interface language.
	Culture *Culture `json:"Culture,omitempty"`
	// Optional store account number.
	InvID *int64 `json:"InvId,omitempty"`
	// Required invoice amount.
	OutSum float64 `json:"OutSum"`
	// Optional name of the product or service.
	Description *string `json:"Description,omitempty"`
	// Optional internal comment for employees. Displayed in the "invoicing" section of your personal account.
	MerchantComments *string `json:"MerchantComments,omitempty"`
	// User parameters
	UserFields map[string]string `json:"UserFields,omitempty"`
	// Nomenclature for fiscalization (structure is similar to Receipt). If this parameter is not passed,
	// the receipt will display the value "free sale", which is not compliant with tax legislation and may result in fines.
	// In some fiscalization cases, the receipt will not be generated. Optional.
	InvoiceItems []*InvoiceItem `json:"InvoiceItems,omitempty"`
	// Optional additional redirect URL with method specification for redirect after success.
	SuccessURL2Data *URLData `json:"SuccessUrl2Data,omitempty"`
	// Optional additional redirect URL with method specification for redirect after fail.
	FailURL2Data *URLData `json:"FailUrl2Data,omitempty"`
}

type DeactivateInvoiceRequest struct {
	MerchantLogin string `json:"MerchantLogin,omitempty"`
	// The last part of the invoice link, e.g. 6hucaX7-BkKNi4lyi-Iu2g in auth.robokassa.ru/merchant/Invoice/6hucaX7-BkKNi4lyi-Iu2g
	EncodedID *string `json:"EncodedId,omitempty"`
	// Account ID, returned in the account creation response.
	ID *string `json:"Id,omitempty"`
	// The invoice number specified by the seller when creating the link. If not provided, it is generated
	// automatically and is available in the invoice creation response and in the "invoicing" section.
	InvID *int64 `json:"InvId,omitempty"`
}

type GetInvoiceInformationListRequest struct {
	MerchantLogin string `json:"MerchantLogin,omitempty"`
	// Current page number (from 1), required
	CurrentPage int `json:"CurrentPage"`
	// Number of records in the response, required
	PageSize int `json:"PageSize"`
	// Invoice statuses, required
	InvoiceStatuses []InvoiceStatus `json:"InvoiceStatuses"`
	// A string of keywords to search by amount, ID, description, or email
	Keywords *string `json:"Keywords,omitempty"`
	// The lower limit of the invoice creation date filter, required
	DateFrom *time.Time `json:"DateFrom"`
	// Upper limit of the invoice creation date filter, required
	DateTo *time.Time `json:"DateTo"`
	// Ascending sort flag
	IsAscending *bool `json:"IsAscending,omitempty"`
	// Link (Invoice) types, required
	InvoiceTypes []InvoiceType `json:"InvoiceTypes"`
	// List of payment aliases
	PaymentAliases []string `json:"PaymentAliases,omitempty"`
	// Minimum bill amount
	SumFrom *float64 `json:"SumFrom,omitempty"`
	// Maximum bill amount
	SumTo *float64 `json:"SumTo,omitempty"`
}

type CreateInvoiceResponse struct {
	URL string
	RawResponse
}

// CreateInvoice creates a new billing link.
func (c *Client) CreateInvoice(ctx context.Context, req CreateInvoiceRequest) (*CreateInvoiceResponse, error) {
	if strings.TrimSpace(req.MerchantLogin) == "" {
		req.MerchantLogin = c.merchantLogin
	}

	if err := req.validate(); err != nil {
		return nil, err
	}

	raw, err := c.doJWTRequest(ctx, invoiceCreateEndpoint, req)
	if err != nil {
		return nil, err
	}

	res := &CreateInvoiceResponse{RawResponse: *raw}
	if raw.String != "" {
		res.URL = raw.String
	}
	if raw.Object != nil {
		if url := findPaymentURL(raw.Object); url != "" {
			res.URL = url
		}
	}
	return res, nil
}

// DeactivateInvoice deactivates invoice link.
func (c *Client) DeactivateInvoice(ctx context.Context, req DeactivateInvoiceRequest) (*RawResponse, error) {
	if strings.TrimSpace(req.MerchantLogin) == "" {
		req.MerchantLogin = c.merchantLogin
	}
	if err := req.validate(); err != nil {
		return nil, err
	}
	return c.doJWTRequest(ctx, invoiceDeactivateEndpoint, req)
}

// GetInvoiceInformationList retrieves list of invoices or links according to filters in GetInvoiceInformationListRequest
func (c *Client) GetInvoiceInformationList(ctx context.Context, req GetInvoiceInformationListRequest) (*RawResponse, error) {
	if strings.TrimSpace(req.MerchantLogin) == "" {
		req.MerchantLogin = c.merchantLogin
	}
	if err := req.validate(); err != nil {
		return nil, err
	}
	return c.doJWTRequest(ctx, invoiceInformationListEndpoint, req)
}

func (r CreateInvoiceRequest) validate() error {
	err := validation.ValidateStruct(&r,
		validation.Field(&r.MerchantLogin, requiredTrimmedStringRule("merchant login is required")),
		validation.Field(&r.InvoiceType, validation.Required, validation.In(InvoiceTypeOneTime, InvoiceTypeReusable).Error("must be InvoiceTypeOneTime or InvoiceTypeReusable")),
		validation.Field(&r.OutSum, validation.Required.Error("must be greater than zero")),
		validation.Field(&r.UserFields, validation.By(func(interface{}) error {
			for key := range r.UserFields {
				if strings.TrimSpace(key) == "" {
					return errors.New("user fields cannot contain empty keys")
				}
			}
			return nil
		})),
		validation.Field(&r.Culture, validation.NilOrNotEmpty, validation.In(CultureEn, CultureRu).Error("must be CultureEn or CultureRu")),
		validation.Field(&r.InvoiceItems, validation.By(func(interface{}) error {
			return validateInvoiceItems(r.InvoiceItems, "invoice items")
		})),
		validation.Field(&r.SuccessURL2Data, validation.By(func(interface{}) error {
			if err := validateURLData(r.SuccessURL2Data); err != nil {
				return fmt.Errorf("invalid SuccessUrl2Data: %w", err)
			}
			return nil
		})),
		validation.Field(&r.FailURL2Data, validation.By(func(interface{}) error {
			if err := validateURLData(r.FailURL2Data); err != nil {
				return fmt.Errorf("invalid FailUrl2Data: %w", err)
			}
			return nil
		})),
	)
	return firstValidationError(err,
		"MerchantLogin",
		"InvoiceType",
		"OutSum",
		"UserFields",
		"InvoiceItems",
		"SuccessURL2Data",
		"FailURL2Data",
	)
}

func (r DeactivateInvoiceRequest) validate() error {
	err := validation.ValidateStruct(&r,
		validation.Field(&r.MerchantLogin, requiredTrimmedStringRule("merchant login is required")),
		validation.Field(&r.EncodedID, validation.By(func(interface{}) error {
			identifiers := 0
			if r.EncodedID != nil && strings.TrimSpace(*r.EncodedID) != "" {
				identifiers++
			}
			if r.ID != nil && strings.TrimSpace(*r.ID) != "" {
				identifiers++
			}
			if r.InvID != nil {
				identifiers++
			}
			if identifiers == 0 {
				return errors.New("at least one identifier is required: EncodedId, Id, or InvId")
			}
			return nil
		})),
	)
	return firstValidationError(err, "MerchantLogin", "EncodedID")
}

func (r GetInvoiceInformationListRequest) validate() error {
	err := validation.ValidateStruct(&r,
		validation.Field(&r.MerchantLogin, requiredTrimmedStringRule("merchant login is required")),
		validation.Field(&r.CurrentPage, validation.By(func(interface{}) error { return validateCurrentPage(r.CurrentPage) })),
		validation.Field(&r.PageSize, validation.Required, validation.Min(1)),
		validation.Field(&r.InvoiceStatuses, validation.Required, validation.Each(validation.In(InvoiceStatusPaid, InvoiceStatusExpired, InvoiceStatusNotPaid))),
		validation.Field(&r.InvoiceTypes, validation.Each(validation.In(InvoiceTypeOneTime, InvoiceTypeReusable).Error("must be InvoiceTypeOneTime or InvoiceTypeReusable"))),
		validation.Field(
			&r.DateFrom,
			validation.By(func(interface{}) error { return validateDateRangeRequired(r.DateFrom, r.DateTo) }),
		),
		validation.Field(
			&r.DateTo,
			validation.By(func(interface{}) error { return validateDateRangeOrder(r.DateFrom, r.DateTo) }),
		),
		validation.Field(&r.SumFrom, validation.By(func(interface{}) error { return validateSumFrom(r.SumFrom) })),
		validation.Field(&r.SumTo, validation.By(func(interface{}) error { return validateSumTo(r.SumFrom, r.SumTo) })),
	)
	return firstValidationError(err,
		"MerchantLogin",
		"CurrentPage",
		"PageSize",
		"InvoiceStatuses",
		"InvoiceTypes",
		"DateFrom",
		"DateTo",
		"SumFrom",
		"SumTo",
	)
}

func validateCurrentPage(currentPage int) error {
	if currentPage < 1 {
		return errors.New("current page must be >= 1")
	}
	return nil
}

func validateDateRangeRequired(dateFrom, dateTo *time.Time) error {
	if dateFrom == nil || dateTo == nil {
		return errors.New("date range is required: DateFrom and DateTo")
	}
	return nil
}

func validateDateRangeOrder(dateFrom, dateTo *time.Time) error {
	if !internalvalidation.IsTimeBefore(dateFrom, dateTo) {
		return errors.New("date range is invalid: DateTo cannot be before DateFrom")
	}
	return nil
}

func validateSumFrom(sumFrom *float64) error {
	if sumFrom != nil && *sumFrom < 0 {
		return errors.New("sum from cannot be negative")
	}
	return nil
}

func validateSumTo(sumFrom, sumTo *float64) error {
	if sumTo != nil && *sumTo < 0 {
		return errors.New("sum to cannot be negative")
	}
	if !isValidSumTo(sumFrom, sumTo) {
		return errors.New("sum range is invalid: SumTo cannot be less than SumFrom")
	}
	return nil
}

func isValidSumTo(sumFrom, sumTo *float64) bool {
	return sumFrom != nil && sumTo != nil && *sumTo > *sumFrom
}
