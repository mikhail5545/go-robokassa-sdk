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

package requests

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	robokassa "github.com/mikhail5545/go-robokassa-sdk"
)

const legacyExpirationDateLayout = "2006-01-02T15:04"

var legacyExpirationDateLayouts = []string{
	legacyExpirationDateLayout,
	time.RFC3339,
	time.RFC3339Nano,
}

// ToInitPaymentRequest converts legacy PaymentRequest into modern root request type.
func (r PaymentRequest) ToInitPaymentRequest() (robokassa.InitPaymentRequest, error) {
	req := robokassa.InitPaymentRequest{
		OutSum:   r.OutSum,
		Email:    cloneStringPtr(r.Email),
		Culture:  r.Culture,
		Encoding: cloneStringPtr(r.Encoding),
		Receipt:  r.Receipt,
		Shp:      cloneStringMap(r.UserParameters),
	}

	if strings.TrimSpace(r.Description) != "" {
		description := r.Description
		req.Description = &description
	}

	if r.InvID != nil {
		rawInvID := strings.TrimSpace(*r.InvID)
		if rawInvID != "" {
			invID, err := strconv.ParseInt(rawInvID, 10, 64)
			if err != nil {
				return robokassa.InitPaymentRequest{}, fmt.Errorf("invalid inv_id %q: %w", *r.InvID, err)
			}
			req.InvID = &invID
		}
	}

	if r.IsTest != nil {
		req.IsTest = *r.IsTest != 0
	}

	if r.ExpirationDate != nil {
		rawExpirationDate := strings.TrimSpace(*r.ExpirationDate)
		if rawExpirationDate != "" {
			expirationDate, err := parseLegacyExpirationDate(rawExpirationDate)
			if err != nil {
				return robokassa.InitPaymentRequest{}, fmt.Errorf("invalid expiration_date %q: %w", *r.ExpirationDate, err)
			}
			req.ExpirationDate = &expirationDate
		}
	}

	return req, nil
}

// PaymentRequestFromInitPaymentRequest converts modern root request type into legacy shape.
func PaymentRequestFromInitPaymentRequest(req robokassa.InitPaymentRequest) PaymentRequest {
	legacy := PaymentRequest{
		OutSum:         req.OutSum,
		Email:          cloneStringPtr(req.Email),
		Culture:        req.Culture,
		Encoding:       cloneStringPtr(req.Encoding),
		UserParameters: cloneStringMap(req.Shp),
		Receipt:        req.Receipt,
	}

	if legacy.OutSum == 0 && strings.TrimSpace(req.OutSumText) != "" {
		if outSum, err := strconv.ParseFloat(strings.TrimSpace(req.OutSumText), 64); err == nil {
			legacy.OutSum = outSum
		}
	}

	if req.Description != nil {
		legacy.Description = *req.Description
	}
	if req.InvID != nil {
		invID := strconv.FormatInt(*req.InvID, 10)
		legacy.InvID = &invID
	}
	if req.IsTest {
		isTest := 1
		legacy.IsTest = &isTest
	}
	if req.ExpirationDate != nil {
		expirationDate := req.ExpirationDate.UTC().Format(legacyExpirationDateLayout)
		legacy.ExpirationDate = &expirationDate
	}

	return legacy
}

func parseLegacyExpirationDate(value string) (time.Time, error) {
	var lastErr error
	for _, layout := range legacyExpirationDateLayouts {
		var (
			parsed time.Time
			err    error
		)
		if layout == legacyExpirationDateLayout {
			parsed, err = time.ParseInLocation(layout, value, time.UTC)
		} else {
			parsed, err = time.Parse(layout, value)
		}
		if err == nil {
			return parsed.UTC(), nil
		}
		lastErr = err
	}
	if lastErr == nil {
		lastErr = fmt.Errorf("unsupported expiration date format")
	}
	return time.Time{}, lastErr
}

func cloneStringMap(in map[string]string) map[string]string {
	if len(in) == 0 {
		return nil
	}

	out := make(map[string]string, len(in))
	for key, value := range in {
		out[key] = value
	}
	return out
}

func cloneStringPtr(value *string) *string {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}
