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
	"strings"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

func validateRequiredTrimmed(value, message string) error {
	return validation.Validate(strings.TrimSpace(value), validation.Required.Error(message))
}

func validatePositiveInt64(value int64, message string) error {
	return validation.Validate(value, validation.Min(int64(1)).Error(message))
}

func validateStringIn(value, message string, allowed ...string) error {
	options := make([]any, 0, len(allowed))
	for _, option := range allowed {
		options = append(options, option)
	}

	return validation.Validate(value, validation.In(options...).Error(message))
}
