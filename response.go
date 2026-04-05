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
	"bytes"
	"encoding/json"
)

type RawResponse struct {
	RawJSON json.RawMessage
	String  string
	Object  map[string]any
}

func parseRawResponse(body []byte) (*RawResponse, error) {
	raw := make([]byte, len(body))
	copy(raw, body)

	trimmed := bytes.TrimSpace(body)
	resp := &RawResponse{RawJSON: raw}

	if len(trimmed) == 0 {
		return resp, nil
	}

	var asString string
	if err := json.Unmarshal(trimmed, &asString); err == nil {
		resp.String = asString
		return resp, nil
	}

	var asObject map[string]any
	if err := json.Unmarshal(trimmed, &asObject); err == nil {
		resp.Object = asObject
		return resp, nil
	}

	resp.String = string(trimmed)
	return resp, nil
}
