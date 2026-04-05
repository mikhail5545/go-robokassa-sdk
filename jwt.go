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
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"hash"
	"strings"

	"golang.org/x/crypto/ripemd160"
)

type jwtHeader struct {
	Typ string `json:"typ"`
	Alg string `json:"alg,omitempty"`
}

func (c *Client) createToken(payload any) (string, error) {
	header := jwtHeader{
		Typ: "JWT",
		Alg: string(c.algorithm),
	}

	headerBytes, err := json.Marshal(header)
	if err != nil {
		return "", fmt.Errorf("marshal jwt header: %w", err)
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshal jwt payload: %w", err)
	}

	headerEncoded := base64.RawURLEncoding.EncodeToString(headerBytes)
	payloadEncoded := base64.RawURLEncoding.EncodeToString(payloadBytes)
	signingInput := headerEncoded + "." + payloadEncoded

	signature, err := c.sign(signingInput)
	if err != nil {
		return "", err
	}

	return signingInput + "." + signature, nil
}

func (c *Client) sign(input string) (string, error) {
	hashFactory, err := signerForAlgorithm(c.algorithm)
	if err != nil {
		return "", err
	}

	secret := []byte(c.merchantLogin + ":" + c.password1)
	mac := hmac.New(hashFactory, secret)
	_, _ = mac.Write([]byte(input))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil)), nil
}

func signerForAlgorithm(algorithm SignatureAlgorithm) (func() hash.Hash, error) {
	switch strings.ToUpper(string(algorithm)) {
	case "MD5":
		return md5.New, nil
	case "RIPEMD160":
		return ripemd160.New, nil
	case "SHA1", "HS1":
		return sha1.New, nil
	case "SHA256", "HS256":
		return sha256.New, nil
	case "SHA384", "HS384":
		return sha512.New384, nil
	case "SHA512", "HS512":
		return sha512.New, nil
	default:
		return nil, fmt.Errorf("unsupported signature algorithm: %q", algorithm)
	}
}
