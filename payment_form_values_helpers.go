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

import "net/url"

func setValueIfNotEmpty(values url.Values, key, value string) {
	if value != "" {
		values.Set(key, value)
	}
}

func setValueIfTrue(values url.Values, key, trueValue string, condition bool) {
	if condition {
		values.Set(key, trueValue)
	}
}

func addPaymentMethods(values url.Values, methods []string) {
	for _, method := range methods {
		values.Add("PaymentMethods", method)
	}
}

func setShpValues(values url.Values, shp map[string]string) {
	for _, key := range sortedKeys(shp) {
		values.Set(key, shp[key])
	}
}
