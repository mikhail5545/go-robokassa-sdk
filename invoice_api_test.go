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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateInvoiceRequestValidation(t *testing.T) {
	for _, tt := range []struct {
		name         string
		buildRequest func() CreateInvoiceRequest
		wantErr      string
	}{
		{
			name: "valid request",
			buildRequest: func() CreateInvoiceRequest {
				return validCreateInvoiceRequest()
			},
		},
		{
			name: "no merchant login",
			buildRequest: func() CreateInvoiceRequest {
				valid := validCreateInvoiceRequest()
				valid.MerchantLogin = ""
				return valid
			},
			wantErr: "merchant login is required",
		},
		{
			name: "invalid invoice type",
			buildRequest: func() CreateInvoiceRequest {
				valid := validCreateInvoiceRequest()
				valid.InvoiceType = "unsupported"
				return valid
			},
			wantErr: "must be InvoiceTypeOneTime or InvoiceTypeReusable",
		},
		{
			name: "invalid culture",
			buildRequest: func() CreateInvoiceRequest {
				valid := validCreateInvoiceRequest()
				culture := Culture("unsupported")
				valid.Culture = &culture
				return valid
			},
			wantErr: "must be CultureEn or CultureRu",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			req := tt.buildRequest()
			err := req.validate()
			if tt.wantErr != "" {
				assert.Error(t, err)
				assert.ErrorContains(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func validCreateInvoiceRequest() CreateInvoiceRequest {
	return CreateInvoiceRequest{
		MerchantLogin: "merchant",
		InvoiceType:   InvoiceTypeOneTime,
		OutSum:        10.50,
	}
}
