// Copyright (c) BlockVision, Inc. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package types

import "fmt"

type ErrorCode string

const (
	ErrorCodeInvalidConfig ErrorCode = "invalid_config"
	ErrorCodeNotFound      ErrorCode = "not_found"
	ErrorCodeTransport     ErrorCode = "transport_error"
)

// SDKError is the public typed error returned by the v2 client.
// Use errors.Is for broad category checks and errors.As for details.
type SDKError struct {
	Code      ErrorCode `json:"code"`
	Message   string    `json:"message"`
	Transport string    `json:"transport,omitempty"`
	Method    string    `json:"method,omitempty"`
	Operation string    `json:"operation,omitempty"`
	Cause     error     `json:"-"`
}

func (e *SDKError) Error() string {
	if e == nil {
		return "<nil>"
	}
	msg := e.Message
	if msg == "" {
		msg = string(e.Code)
	}
	prefix := ""
	if e.Transport != "" {
		prefix = e.Transport
	}
	if e.Method != "" {
		if prefix != "" {
			prefix += " "
		}
		prefix += e.Method
	}
	if e.Operation != "" {
		if prefix != "" {
			prefix += " "
		}
		prefix += e.Operation
	}
	switch {
	case prefix != "" && e.Cause != nil:
		return fmt.Sprintf("%s: %s: %v", prefix, msg, e.Cause)
	case prefix != "":
		return fmt.Sprintf("%s: %s", prefix, msg)
	case e.Cause != nil:
		return fmt.Sprintf("%s: %v", msg, e.Cause)
	default:
		return msg
	}
}

func (e *SDKError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Cause
}

func (e *SDKError) Is(target error) bool {
	t, ok := target.(*SDKError)
	if !ok {
		return false
	}
	return e.Code == t.Code
}

func (e *SDKError) withContext(transport, method, operation string, cause error) *SDKError {
	if e == nil {
		return nil
	}
	cp := *e
	if transport != "" {
		cp.Transport = transport
	}
	if method != "" {
		cp.Method = method
	}
	if operation != "" {
		cp.Operation = operation
	}
	cp.Cause = cause
	return &cp
}

var (
	ErrBackendRequired = &SDKError{
		Code:    ErrorCodeInvalidConfig,
		Message: "either GrpcClient or HttpConn is required",
	}
	ErrGrpcClientRequired = &SDKError{
		Code:    ErrorCodeInvalidConfig,
		Message: "gRPC client is required",
	}
	ErrHttpConnRequired = &SDKError{
		Code:    ErrorCodeInvalidConfig,
		Message: "HTTP connection is required",
	}
	ErrTransactionNotFound = &SDKError{
		Code:    ErrorCodeNotFound,
		Message: "transaction not found",
		Method:  "GetTransaction",
	}
	ErrSystemStateNotFound = &SDKError{
		Code:    ErrorCodeNotFound,
		Message: "system state not found",
		Method:  "GetCurrentSystemState",
	}
	ErrChainIdentifierNotFound = &SDKError{
		Code:    ErrorCodeNotFound,
		Message: "chain identifier not found",
		Method:  "GetChainIdentifier",
	}
)

func WrapKnownError(base *SDKError, transport, method, operation string, cause error) error {
	return base.withContext(transport, method, operation, cause)
}

func NewTransportError(transport, method, operation string, cause error) error {
	return (&SDKError{
		Code:      ErrorCodeTransport,
		Message:   "transport error",
		Transport: transport,
		Method:    method,
		Operation: operation,
		Cause:     cause,
	})
}
