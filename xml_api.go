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
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type XMLResult struct {
	Code        int    `xml:"Code"`
	Description string `xml:"Description"`
}

// XMLAPIError represents business-level XML API failure (`Result.Code != 0`).
type XMLAPIError struct {
	Code        int
	Description string
}

func (e *XMLAPIError) Error() string {
	return fmt.Sprintf("robokassa xml api error: code=%d description=%q", e.Code, e.Description)
}

type Currency struct {
	Label    string `xml:"Label,attr"`
	Alias    string `xml:"Alias,attr"`
	Name     string `xml:"Name,attr"`
	MinValue string `xml:"MinValue,attr"`
	MaxValue string `xml:"MaxValue,attr"`
}

type CurrencyGroup struct {
	Code        string     `xml:"Code,attr"`
	Description string     `xml:"Description,attr"`
	Items       []Currency `xml:"Items>Currency"`
}

type GetCurrenciesResponse struct {
	XMLName xml.Name        `xml:"CurrenciesList"`
	Result  XMLResult       `xml:"Result"`
	Groups  []CurrencyGroup `xml:"Groups>Group"`
}

type OpStateMethod struct {
	Code string `xml:"Code"`
}

type OpStateState struct {
	Code        int    `xml:"Code"`
	RequestDate string `xml:"RequestDate"`
	StateDate   string `xml:"StateDate"`
}

type OpStateInfo struct {
	IncCurrLabel  string        `xml:"IncCurrLabel"`
	IncSum        string        `xml:"IncSum"`
	IncAccount    string        `xml:"IncAccount"`
	PaymentMethod OpStateMethod `xml:"PaymentMethod"`
	OutCurrLabel  string        `xml:"OutCurrLabel"`
	OutSum        string        `xml:"OutSum"`
	OpKey         string        `xml:"OpKey"`
	BankCardRRN   string        `xml:"BankCardRRN"`
}

type OpStateUserField struct {
	Name  string `xml:"Name"`
	Value string `xml:"Value"`
}

type OpStateExtResponse struct {
	XMLName    xml.Name           `xml:"OperationStateResponse"`
	Result     XMLResult          `xml:"Result"`
	State      OpStateState       `xml:"State"`
	Info       OpStateInfo        `xml:"Info"`
	UserFields []OpStateUserField `xml:"UserField>Field"`
}

// GetCurrencies requests available payment methods/currency groups from XML API.
func (c *Client) GetCurrencies(ctx context.Context, language *Culture) (*GetCurrenciesResponse, error) {
	lang := "ru"
	if language != nil && strings.TrimSpace(language.String()) != "" {
		lang = strings.TrimSpace(language.String())
	}
	if err := validateStringIn(lang, "language must be ru or en", "ru", "en"); err != nil {
		return nil, err
	}

	params := make(url.Values)
	params.Set("MerchantLogin", c.merchantLogin)
	params.Set("Language", lang)

	body, err := c.doXMLRequest(ctx, "GetCurrencies", params)
	if err != nil {
		return nil, err
	}

	var response GetCurrenciesResponse
	if err := xml.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("unmarshal GetCurrencies response: %w", err)
	}
	if response.Result.Code != 0 {
		return nil, &XMLAPIError{Code: response.Result.Code, Description: response.Result.Description}
	}
	return &response, nil
}

// OpStateExt requests operation status/details by invoice id from XML API.
//
// password #2 must be configured via [WithPassword2].
func (c *Client) OpStateExt(ctx context.Context, invoiceID int64) (*OpStateExtResponse, error) {
	if err := validatePositiveInt64(invoiceID, "invoice id must be greater than zero"); err != nil {
		return nil, err
	}
	if err := validateRequiredTrimmed(c.password2, "password2 is required for OpStateExt"); err != nil {
		return nil, err
	}

	signatureSource := c.merchantLogin + ":" + strconv.FormatInt(invoiceID, 10) + ":" + c.password2
	signature, err := c.hashHex(signatureSource)
	if err != nil {
		return nil, err
	}

	params := make(url.Values)
	params.Set("MerchantLogin", c.merchantLogin)
	params.Set("InvoiceID", strconv.FormatInt(invoiceID, 10))
	params.Set("Signature", signature)

	body, err := c.doXMLRequest(ctx, "OpStateExt", params)
	if err != nil {
		return nil, err
	}

	var response OpStateExtResponse
	if err := xml.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("unmarshal OpStateExt response: %w", err)
	}
	if response.Result.Code != 0 {
		return nil, &XMLAPIError{Code: response.Result.Code, Description: response.Result.Description}
	}
	return &response, nil
}

func (c *Client) doXMLRequest(ctx context.Context, method string, params url.Values) ([]byte, error) {
	endpoint := c.xmlBaseURL + "/" + method
	if encoded := params.Encode(); encoded != "" {
		endpoint += "?" + encoded
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("create xml request: %w", err)
	}
	req.Header.Set("Accept", "application/xml, text/xml")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send xml request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read xml response body: %w", err)
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, &APIError{StatusCode: resp.StatusCode, Body: strings.TrimSpace(string(body))}
	}
	return body, nil
}
