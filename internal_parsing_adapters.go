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
	"net/url"
	"time"

	internalparsing "github.com/mikhail5545/go-robokassa-sdk/internal/parsing"
)

func rawResponseObject(raw *RawResponse) (map[string]any, error) {
	return internalparsing.RawResponseObject(raw.Object, raw.RawJSON)
}

func nestedDataMap(object map[string]any) map[string]any {
	return internalparsing.NestedDataMap(object)
}

func invoiceRowsFrom(object map[string]any) []any {
	return internalparsing.InvoiceRowsFrom(object)
}

func findPaymentURL(obj map[string]any) string {
	return internalparsing.FindPaymentURL(obj)
}

func firstBool(object map[string]any, keys ...string) bool {
	return internalparsing.FirstBool(object, keys...)
}

func firstString(object map[string]any, keys ...string) string {
	return internalparsing.FirstString(object, keys...)
}

func firstInt(primary, fallback map[string]any, keys ...string) int {
	return internalparsing.FirstInt(primary, fallback, keys...)
}

func firstFloat(object map[string]any, keys ...string) (float64, bool) {
	return internalparsing.FirstFloat(object, keys...)
}

func firstTime(object map[string]any, keys ...string) *time.Time {
	return internalparsing.FirstTime(object, keys...)
}

func callbackValue(values url.Values, keys ...string) string {
	return internalparsing.CallbackValue(values, keys...)
}
