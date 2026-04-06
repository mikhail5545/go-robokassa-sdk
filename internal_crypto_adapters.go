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
	"crypto"
	"crypto/x509"
	"hash"

	internalcrypto "github.com/mikhail5545/go-robokassa-sdk/internal/crypto"
)

func signerForAlgorithm(algorithm SignatureAlgorithm) (func() hash.Hash, error) {
	return internalcrypto.SignerForAlgorithm(string(algorithm))
}

func refundSignerForAlgorithm(algorithm SignatureAlgorithm) (headerAlgorithm string, factory func() hash.Hash, err error) {
	return internalcrypto.RefundSignerForAlgorithm(string(algorithm))
}

func splitJWS(token string) (headerRaw []byte, payloadRaw []byte, signingInput string, signature []byte, err error) {
	return internalcrypto.SplitJWS(token)
}

func parseCertificate(certificateData []byte) (*x509.Certificate, error) {
	return internalcrypto.ParseCertificate(certificateData)
}

func jwtRSHash(alg string) (crypto.Hash, hash.Hash, error) {
	return internalcrypto.JWTRSHash(alg)
}
