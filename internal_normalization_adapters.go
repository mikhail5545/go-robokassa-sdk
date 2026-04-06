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

	internalnormalization "github.com/mikhail5545/go-robokassa-sdk/internal/normalization"
)

func normalizeHTTPMethod(method *string) (string, error) {
	return internalnormalization.NormalizeHTTPMethod(method)
}

func normalizeShpParams(in map[string]string) (map[string]string, error) {
	return internalnormalization.NormalizeShpParams(in)
}

func normalizeInvoiceID(invoiceID int64) (string, error) {
	return internalnormalization.NormalizeInvoiceID(invoiceID)
}

func buildURLWithValues(rawURL string, values url.Values) (string, error) {
	return internalnormalization.BuildURLWithValues(rawURL, values)
}

func outAmountToRaw(value float64) (int64, error) {
	return internalnormalization.OutAmountToRaw(value)
}

func formatOutSum(value float64) string {
	return internalnormalization.FormatOutSum(value)
}

func trimPtr(v *string) string {
	return internalnormalization.TrimPtr(v)
}

func sortedKeys[K ~string, V any](m map[K]V) []K {
	return internalnormalization.SortedKeys(m)
}
