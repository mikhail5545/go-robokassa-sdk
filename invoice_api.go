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
)

const (
	createInvoicePath             = "/api/CreateInvoice"
	deactivateInvoicePath         = "/api/DeactivateInvoice"
	getInvoiceInformationListPath = "/api/GetInvoiceInformationList"
)

type URLData struct {
	URL    string `json:"Url"`
	Method string `json:"Method,omitempty"`
}

type CreateInvoiceRequest struct {
	MerchantLogin    string            `json:"MerchantLogin,omitempty"`
	InvoiceType      InvoiceType       `json:"InvoiceType"`
	Culture          *Culture          `json:"Culture,omitempty"`
	InvID            *int64            `json:"InvId,omitempty"`
	OutSum           float64           `json:"OutSum"`
	Description      *string           `json:"Description,omitempty"`
	MerchantComments *string           `json:"MerchantComments,omitempty"`
	UserFields       map[string]string `json:"UserFields,omitempty"`
	InvoiceItems     []*InvoiceItem    `json:"InvoiceItems,omitempty"`
	SuccessURL2Data  *URLData          `json:"SuccessUrl2Data,omitempty"`
	FailURL2Data     *URLData          `json:"FailUrl2Data,omitempty"`
}

type DeactivateInvoiceRequest struct {
	MerchantLogin string  `json:"MerchantLogin,omitempty"`
	EncodedID     *string `json:"EncodedId,omitempty"`
	ID            *string `json:"Id,omitempty"`
	InvID         *int64  `json:"InvId,omitempty"`
}

type GetInvoiceInformationListRequest struct {
	MerchantLogin   string          `json:"MerchantLogin,omitempty"`
	CurrentPage     int             `json:"CurrentPage"`
	PageSize        int             `json:"PageSize"`
	InvoiceStatuses []InvoiceStatus `json:"InvoiceStatuses"`
	Keywords        *string         `json:"Keywords,omitempty"`
	DateFrom        *time.Time      `json:"DateFrom"`
	DateTo          *time.Time      `json:"DateTo"`
	IsAscending     *bool           `json:"IsAscending,omitempty"`
	InvoiceTypes    []InvoiceType   `json:"InvoiceTypes"`
	PaymentAliases  []string        `json:"PaymentAliases,omitempty"`
	SumFrom         *float64        `json:"SumFrom,omitempty"`
	SumTo           *float64        `json:"SumTo,omitempty"`
}

type CreateInvoiceResponse struct {
	URL string
	RawResponse
}

func (c *Client) CreateInvoice(ctx context.Context, req CreateInvoiceRequest) (*CreateInvoiceResponse, error) {
	if strings.TrimSpace(req.MerchantLogin) == "" {
		req.MerchantLogin = c.merchantLogin
	}

	if err := req.validate(); err != nil {
		return nil, err
	}

	raw, err := c.doJWTRequest(ctx, createInvoicePath, req)
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

func (c *Client) DeactivateInvoice(ctx context.Context, req DeactivateInvoiceRequest) (*RawResponse, error) {
	if strings.TrimSpace(req.MerchantLogin) == "" {
		req.MerchantLogin = c.merchantLogin
	}
	if err := req.validate(); err != nil {
		return nil, err
	}
	return c.doJWTRequest(ctx, deactivateInvoicePath, req)
}

func (c *Client) GetInvoiceInformationList(ctx context.Context, req GetInvoiceInformationListRequest) (*RawResponse, error) {
	if strings.TrimSpace(req.MerchantLogin) == "" {
		req.MerchantLogin = c.merchantLogin
	}
	if err := req.validate(); err != nil {
		return nil, err
	}
	return c.doJWTRequest(ctx, getInvoiceInformationListPath, req)
}

func (r CreateInvoiceRequest) validate() error {
	err := validation.ValidateStruct(&r,
		validation.Field(&r.MerchantLogin, requiredTrimmedStringRule("merchant login is required")),
		validation.Field(&r.InvoiceType, validation.By(func(interface{}) error {
			if r.InvoiceType != InvoiceTypeOneTime && r.InvoiceType != InvoiceTypeReusable {
				return fmt.Errorf("invoice type must be %q or %q", InvoiceTypeOneTime, InvoiceTypeReusable)
			}
			return nil
		})),
		validation.Field(&r.OutSum, validation.By(func(interface{}) error {
			if r.OutSum <= 0 {
				return errors.New("out sum must be greater than zero")
			}
			return nil
		})),
		validation.Field(&r.UserFields, validation.By(func(interface{}) error {
			for key := range r.UserFields {
				if strings.TrimSpace(key) == "" {
					return errors.New("user fields cannot contain empty keys")
				}
			}
			return nil
		})),
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
		validation.Field(&r.CurrentPage, validation.By(func(interface{}) error {
			if r.CurrentPage < 1 {
				return errors.New("current page must be >= 1")
			}
			return nil
		})),
		validation.Field(&r.PageSize, validation.By(func(interface{}) error {
			if r.PageSize < 1 {
				return errors.New("page size must be >= 1")
			}
			return nil
		})),
		validation.Field(&r.InvoiceStatuses, validation.By(func(interface{}) error {
			if len(r.InvoiceStatuses) == 0 {
				return errors.New("invoice statuses are required")
			}
			for _, status := range r.InvoiceStatuses {
				if status != InvoiceStatusPaid && status != InvoiceStatusExpired && status != InvoiceStatusNotPaid {
					return fmt.Errorf("unsupported invoice status: %q", status)
				}
			}
			return nil
		})),
		validation.Field(&r.InvoiceTypes, validation.By(func(interface{}) error {
			if len(r.InvoiceTypes) == 0 {
				return errors.New("invoice types are required")
			}
			for _, invoiceType := range r.InvoiceTypes {
				if invoiceType != InvoiceTypeOneTime && invoiceType != InvoiceTypeReusable {
					return fmt.Errorf("unsupported invoice type: %q", invoiceType)
				}
			}
			return nil
		})),
		validation.Field(&r.DateFrom, validation.By(func(interface{}) error {
			if r.DateFrom == nil || r.DateTo == nil {
				return errors.New("date range is required: DateFrom and DateTo")
			}
			return nil
		})),
		validation.Field(&r.DateTo, validation.By(func(interface{}) error {
			if r.DateFrom != nil && r.DateTo != nil && r.DateTo.Before(*r.DateFrom) {
				return errors.New("date range is invalid: DateTo cannot be before DateFrom")
			}
			return nil
		})),
		validation.Field(&r.SumFrom, validation.By(func(interface{}) error {
			if r.SumFrom != nil && *r.SumFrom < 0 {
				return errors.New("sum from cannot be negative")
			}
			return nil
		})),
		validation.Field(&r.SumTo, validation.By(func(interface{}) error {
			if r.SumTo != nil && *r.SumTo < 0 {
				return errors.New("sum to cannot be negative")
			}
			if r.SumFrom != nil && r.SumTo != nil && *r.SumTo < *r.SumFrom {
				return errors.New("sum range is invalid: SumTo cannot be less than SumFrom")
			}
			return nil
		})),
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

func validateURLData(urlData *URLData) error {
	if urlData == nil {
		return nil
	}
	err := validation.ValidateStruct(urlData,
		validation.Field(&urlData.URL, requiredTrimmedStringRule("url is required")),
		validation.Field(&urlData.Method, validation.By(func(interface{}) error {
			if urlData.Method == "" {
				return nil
			}
			method := strings.ToUpper(urlData.Method)
			if method != "GET" && method != "POST" {
				return errors.New("method must be GET or POST")
			}
			return nil
		})),
	)
	return firstValidationError(err, "URL", "Method")
}

func findPaymentURL(obj map[string]any) string {
	keys := []string{
		"InvoiceUrl", "invoiceUrl",
		"PaymentUrl", "paymentUrl",
		"ShortUrl", "shortUrl",
		"Url", "url",
		"Link", "link",
	}
	for _, key := range keys {
		if value, ok := obj[key]; ok {
			if s, ok := value.(string); ok && strings.TrimSpace(s) != "" {
				return s
			}
		}
	}

	if data, ok := obj["Data"]; ok {
		if nested, ok := data.(map[string]any); ok {
			return findPaymentURL(nested)
		}
	}
	if data, ok := obj["data"]; ok {
		if nested, ok := data.(map[string]any); ok {
			return findPaymentURL(nested)
		}
	}

	return ""
}
