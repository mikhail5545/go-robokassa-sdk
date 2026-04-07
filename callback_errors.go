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
	"errors"
	"fmt"
	"strings"
)

var (
	// ErrInvalidCallbackSignature indicates that callback signature comparison failed.
	ErrInvalidCallbackSignature = errors.New("invalid callback signature")
	// ErrUnsupportedCallbackSignatureKind indicates unsupported callback verification mode.
	ErrUnsupportedCallbackSignatureKind = errors.New("unsupported callback signature kind")

	// ErrResultURL2InvalidToken indicates malformed JWS token structure.
	ErrResultURL2InvalidToken = errors.New("invalid ResultURL2 JWS token")
	// ErrResultURL2InvalidHeader indicates malformed JWS header JSON.
	ErrResultURL2InvalidHeader = errors.New("invalid ResultURL2 JWS header")
	// ErrResultURL2InvalidPayload indicates malformed JWS payload JSON.
	ErrResultURL2InvalidPayload = errors.New("invalid ResultURL2 JWS payload")
	// ErrResultURL2UnsupportedAlgorithm indicates unsupported JWS signature algorithm.
	ErrResultURL2UnsupportedAlgorithm = errors.New("unsupported ResultURL2 JWS algorithm")
	// ErrResultURL2InvalidCertificate indicates invalid certificate format/data.
	ErrResultURL2InvalidCertificate = errors.New("invalid ResultURL2 certificate")
	// ErrResultURL2InvalidCertificateKey indicates that certificate does not contain RSA public key.
	ErrResultURL2InvalidCertificateKey = errors.New("invalid ResultURL2 certificate public key")
	// ErrResultURL2SignatureVerification indicates failed RSA signature verification.
	ErrResultURL2SignatureVerification = errors.New("failed ResultURL2 signature verification")
)

// CallbackSignatureKind identifies callback signature flow.
type CallbackSignatureKind string

const (
	// CallbackSignatureKindResult verifies ResultURL signatures (password #2).
	CallbackSignatureKindResult CallbackSignatureKind = "result"
	// CallbackSignatureKindSuccess verifies SuccessURL signatures (password #1).
	CallbackSignatureKindSuccess CallbackSignatureKind = "success"
)

// CallbackSignatureMismatchError provides details for signature mismatch errors.
type CallbackSignatureMismatchError struct {
	Kind  CallbackSignatureKind
	InvID string
}

func (e *CallbackSignatureMismatchError) Error() string {
	kind := strings.TrimSpace(string(e.Kind))
	if kind == "" {
		kind = "callback"
	}
	invID := strings.TrimSpace(e.InvID)
	if invID == "" {
		return fmt.Sprintf("%s (%s)", ErrInvalidCallbackSignature, kind)
	}
	return fmt.Sprintf("%s (%s, invID=%s)", ErrInvalidCallbackSignature, kind, invID)
}

func (e *CallbackSignatureMismatchError) Is(target error) bool {
	return target == ErrInvalidCallbackSignature
}

// ResultURL2VerificationError classifies JWS parsing/verifying failures.
type ResultURL2VerificationError struct {
	Category error
	Err      error
}

func (e *ResultURL2VerificationError) Error() string {
	if e == nil {
		return "resulturl2 verification error"
	}
	if e.Category == nil && e.Err == nil {
		return "resulturl2 verification error"
	}
	if e.Category == nil {
		return e.Err.Error()
	}
	if e.Err == nil {
		return e.Category.Error()
	}
	return fmt.Sprintf("%s: %v", e.Category, e.Err)
}

func (e *ResultURL2VerificationError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

func (e *ResultURL2VerificationError) Is(target error) bool {
	return e != nil && e.Category != nil && target == e.Category
}

func wrapResultURL2Error(category error, err error) error {
	return &ResultURL2VerificationError{
		Category: category,
		Err:      err,
	}
}
