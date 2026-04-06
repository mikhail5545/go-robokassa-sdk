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

package crypto

import (
	"bytes"
	"crypto"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"hash"
	"strings"

	"golang.org/x/crypto/ripemd160"
)

func SignerForAlgorithm(algorithm string) (func() hash.Hash, error) {
	switch strings.ToUpper(strings.TrimSpace(algorithm)) {
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

func RefundSignerForAlgorithm(algorithm string) (headerAlgorithm string, factory func() hash.Hash, err error) {
	switch strings.ToUpper(strings.TrimSpace(algorithm)) {
	case "SHA512", "HS512":
		return "HS512", sha512.New, nil
	case "SHA384", "HS384":
		return "HS384", sha512.New384, nil
	case "SHA256", "HS256", "MD5", "RIPEMD160", "SHA1", "HS1", "":
		return "HS256", sha256.New, nil
	default:
		return "", nil, fmt.Errorf("unsupported refund signature algorithm: %q", algorithm)
	}
}

func JWTRSHash(alg string) (crypto.Hash, hash.Hash, error) {
	switch strings.ToUpper(strings.TrimSpace(alg)) {
	case "RS256":
		return crypto.SHA256, sha256.New(), nil
	case "RS384":
		return crypto.SHA384, sha512.New384(), nil
	case "RS512":
		return crypto.SHA512, sha512.New(), nil
	default:
		return 0, nil, fmt.Errorf("unsupported jws algorithm: %q", alg)
	}
}

func SplitJWS(token string) (headerRaw []byte, payloadRaw []byte, signingInput string, signature []byte, err error) {
	parts := strings.Split(strings.TrimSpace(token), ".")
	if len(parts) != 3 {
		return nil, nil, "", nil, errors.New("invalid JWS: expected 3 parts")
	}

	headerRaw, err = base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, nil, "", nil, fmt.Errorf("decode jws header: %w", err)
	}
	payloadRaw, err = base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, nil, "", nil, fmt.Errorf("decode jws payload: %w", err)
	}
	signature, err = base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return nil, nil, "", nil, fmt.Errorf("decode jws signature: %w", err)
	}

	return headerRaw, payloadRaw, parts[0] + "." + parts[1], signature, nil
}

func ParseCertificate(certificateData []byte) (*x509.Certificate, error) {
	certificateData = bytes.TrimSpace(certificateData)
	if len(certificateData) == 0 {
		return nil, errors.New("certificate data is empty")
	}

	if block, _ := pem.Decode(certificateData); block != nil {
		certificateData = block.Bytes
	}

	cert, err := x509.ParseCertificate(certificateData)
	if err != nil {
		return nil, fmt.Errorf("parse certificate: %w", err)
	}
	return cert, nil
}
