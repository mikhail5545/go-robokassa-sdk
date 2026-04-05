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
	"strings"
	"testing"
	"time"

	robokassa "github.com/mikhail5545/go-robokassa-sdk"
	"github.com/mikhail5545/go-robokassa-sdk/models/items"
	"github.com/mikhail5545/go-robokassa-sdk/types"
)

func TestPaymentRequestToInitPaymentRequest(t *testing.T) {
	invID := "42"
	email := "buyer@example.com"
	encoding := "utf-8"
	culture := types.CultureRu
	isTest := 1
	expirationDate := "2026-01-02T15:04"

	legacy := PaymentRequest{
		OutSum:         100.25,
		Description:    "Order #42",
		InvID:          &invID,
		Email:          &email,
		Culture:        &culture,
		Encoding:       &encoding,
		IsTest:         &isTest,
		ExpirationDate: &expirationDate,
		UserParameters: map[string]string{"order": "42"},
	}

	converted, err := legacy.ToInitPaymentRequest()
	if err != nil {
		t.Fatalf("convert payment request: %v", err)
	}

	if converted.OutSum != legacy.OutSum {
		t.Fatalf("unexpected OutSum: got=%.2f want=%.2f", converted.OutSum, legacy.OutSum)
	}
	if converted.InvID == nil || *converted.InvID != 42 {
		t.Fatalf("unexpected InvID: %+v", converted.InvID)
	}
	if converted.Description == nil || *converted.Description != legacy.Description {
		t.Fatalf("unexpected Description: %+v", converted.Description)
	}
	if !converted.IsTest {
		t.Fatal("expected IsTest to be true")
	}
	if converted.ExpirationDate == nil || converted.ExpirationDate.UTC().Format(legacyExpirationDateLayout) != expirationDate {
		t.Fatalf("unexpected ExpirationDate: %+v", converted.ExpirationDate)
	}
	if converted.Shp["order"] != "42" {
		t.Fatalf("unexpected Shp values: %+v", converted.Shp)
	}

	legacy.UserParameters["order"] = "99"
	if converted.Shp["order"] != "42" {
		t.Fatalf("expected converted shp map to be copied, got: %+v", converted.Shp)
	}
}

func TestPaymentRequestToInitPaymentRequest_InvalidValues(t *testing.T) {
	invalidInvID := "forty-two"
	_, err := (PaymentRequest{InvID: &invalidInvID}).ToInitPaymentRequest()
	if err == nil || !strings.Contains(err.Error(), "inv_id") {
		t.Fatalf("expected inv_id conversion error, got: %v", err)
	}

	invalidExpirationDate := "2026/01/02"
	_, err = (PaymentRequest{ExpirationDate: &invalidExpirationDate}).ToInitPaymentRequest()
	if err == nil || !strings.Contains(err.Error(), "expiration_date") {
		t.Fatalf("expected expiration_date conversion error, got: %v", err)
	}
}

func TestPaymentRequestFromInitPaymentRequest(t *testing.T) {
	invID := int64(7)
	description := "Order #7"
	email := "buyer@example.com"
	encoding := "utf-8"
	culture := types.CultureEn
	expirationDate := time.Date(2026, time.January, 2, 15, 4, 0, 0, time.UTC)

	current := robokassa.InitPaymentRequest{
		OutSumText:     "50.50",
		InvID:          &invID,
		Description:    &description,
		Email:          &email,
		Culture:        &culture,
		Encoding:       &encoding,
		IsTest:         true,
		ExpirationDate: &expirationDate,
		Shp:            map[string]string{"Shp_order": "7"},
	}

	legacy := PaymentRequestFromInitPaymentRequest(current)
	if legacy.OutSum != 50.50 {
		t.Fatalf("unexpected OutSum: got=%.2f want=50.50", legacy.OutSum)
	}
	if legacy.InvID == nil || *legacy.InvID != "7" {
		t.Fatalf("unexpected InvID: %+v", legacy.InvID)
	}
	if legacy.IsTest == nil || *legacy.IsTest != 1 {
		t.Fatalf("unexpected IsTest: %+v", legacy.IsTest)
	}
	if legacy.ExpirationDate == nil || *legacy.ExpirationDate != "2026-01-02T15:04" {
		t.Fatalf("unexpected ExpirationDate: %+v", legacy.ExpirationDate)
	}
	if legacy.UserParameters["Shp_order"] != "7" {
		t.Fatalf("unexpected user parameters: %+v", legacy.UserParameters)
	}

	current.Shp["Shp_order"] = "8"
	if legacy.UserParameters["Shp_order"] != "7" {
		t.Fatalf("expected legacy user parameters to be copied, got: %+v", legacy.UserParameters)
	}
}

func TestRefundRequestAdapters(t *testing.T) {
	refundSum := 10.0
	item := &items.InvoiceItem{Name: "Subscription"}
	legacy := RefundRequest{
		OpKey:        "op-key-1",
		RefundSum:    &refundSum,
		InvoiceItems: []*items.InvoiceItem{item},
	}

	current := legacy.ToCreateRefundRequest()
	if current.OpKey != legacy.OpKey {
		t.Fatalf("unexpected OpKey: got=%q want=%q", current.OpKey, legacy.OpKey)
	}
	if len(current.InvoiceItems) != 1 || current.InvoiceItems[0] != item {
		t.Fatalf("unexpected InvoiceItems: %+v", current.InvoiceItems)
	}

	legacy.InvoiceItems[0] = nil
	if current.InvoiceItems[0] == nil {
		t.Fatal("expected converted invoice item slice to be copied")
	}

	backToLegacy := RefundRequestFromCreateRefundRequest(current)
	if backToLegacy.OpKey != current.OpKey {
		t.Fatalf("unexpected round-trip OpKey: got=%q want=%q", backToLegacy.OpKey, current.OpKey)
	}

	current.InvoiceItems[0] = nil
	if backToLegacy.InvoiceItems[0] == nil {
		t.Fatal("expected round-trip invoice item slice to be copied")
	}
}
