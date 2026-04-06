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
	"encoding/json"
	"strings"
	"testing"
)

func TestQuantity3MarshalJSON_BoundsAndFormatting(t *testing.T) {
	t.Run("formats fixed 3 decimals", func(t *testing.T) {
		b, err := json.Marshal(Quantity3(12345))
		if err != nil {
			t.Fatalf("marshal quantity: %v", err)
		}
		if string(b) != "12.345" {
			t.Fatalf("unexpected quantity json: got=%s want=12.345", string(b))
		}
	})

	t.Run("rejects values above docs limit", func(t *testing.T) {
		_, err := json.Marshal(Quantity3(maxQuantity3Raw + 1))
		if err == nil {
			t.Fatal("expected out-of-range quantity error")
		}
	})
}

func TestPrice8x2MarshalJSON_BoundsAndFormatting(t *testing.T) {
	t.Run("formats fixed 2 decimals", func(t *testing.T) {
		b, err := json.Marshal(Price8x2(12345))
		if err != nil {
			t.Fatalf("marshal price: %v", err)
		}
		if string(b) != "123.45" {
			t.Fatalf("unexpected price json: got=%s want=123.45", string(b))
		}
	})

	t.Run("rejects values above docs limit", func(t *testing.T) {
		_, err := json.Marshal(Price8x2(maxPrice8x2Raw + 1))
		if err == nil {
			t.Fatal("expected out-of-range price error")
		}
	})
}

func TestBuildPaymentFormValues_RejectsOutOfRangeReceiptValues(t *testing.T) {
	client := mustClient(t, Config{
		MerchantLogin: "merchant",
		Password1:     "password1",
	})

	_, err := client.BuildPaymentFormValues(InitPaymentRequest{
		OutSum: 10,
		Receipt: &Receipt{
			Items: []*ReceiptItem{
				{
					Name:     "Subscription",
					Quantity: Quantity3(maxQuantity3Raw + 1),
					Sum:      Price8x2(1000),
					Tax:      TaxRateVat20,
				},
			},
		},
	})
	if err == nil {
		t.Fatal("expected quantity limit validation error")
	}
	if !strings.Contains(err.Error(), "quantity") {
		t.Fatalf("expected quantity-related error, got: %v", err)
	}
}

func TestBuildSplitPaymentFormValues_ValidatesSplitSumAndAmountRange(t *testing.T) {
	client := mustClient(t, Config{
		MerchantLogin: "merchant",
		Password1:     "password1",
	})

	_, err := client.BuildSplitPaymentFormValues(SplitPaymentInvoice{
		OutAmount: 700,
		Merchant:  SplitMasterMerchant{ID: "master-shop"},
		SplitMerchants: []SplitMerchant{
			{ID: "master-shop", Amount: Amount(60000)},
			{ID: "partner-shop", Amount: Amount(20000)},
		},
	})
	if err == nil {
		t.Fatal("expected split sum mismatch validation error")
	}

	_, err = client.BuildSplitPaymentFormValues(SplitPaymentInvoice{
		OutAmount: 700,
		Merchant:  SplitMasterMerchant{ID: "master-shop"},
		SplitMerchants: []SplitMerchant{
			{ID: "master-shop", Amount: Amount(maxPrice8x2Raw + 1)},
		},
	})
	if err == nil {
		t.Fatal("expected split amount range validation error")
	}
}
