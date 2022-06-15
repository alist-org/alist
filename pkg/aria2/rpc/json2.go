package rpc

// based on "github.com/gorilla/rpc/v2/json2"

// Copyright 2009 The Go Authors. All rights reserved.
// Copyright 2012 The Gorilla Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
)

// ----------------------------------------------------------------------------
// Request and Response
// ----------------------------------------------------------------------------

// clientRequest represents a JSON-RPC request sent by a client.
type clientRequest struct {
	// JSON-RPC protocol.
	Version string `json:"jsonrpc"`

	// A String containing the name of the method to be invoked.
	Method string `json:"method"`

	// Object to pass as request parameter to the method.
	Params interface{} `json:"params"`

	// The request id. This can be of any type. It is used to match the
	// response with the request that it is replying to.
	Id uint64 `json:"id"`
}

// clientResponse represents a JSON-RPC response returned to a client.
type clientResponse struct {
	Version string           `json:"jsonrpc"`
	Result  *json.RawMessage `json:"result"`
	Error   *json.RawMessage `json:"error"`
	Id      *uint64          `json:"id"`
}

// EncodeClientRequest encodes parameters for a JSON-RPC client request.
func EncodeClientRequest(method string, args interface{}) (*bytes.Buffer, error) {
	var buf bytes.Buffer
	c := &clientRequest{
		Version: "2.0",
		Method:  method,
		Params:  args,
		Id:      reqid(),
	}
	if err := json.NewEncoder(&buf).Encode(c); err != nil {
		return nil, err
	}
	return &buf, nil
}

func (c clientResponse) decode(reply interface{}) error {
	if c.Error != nil {
		jsonErr := &Error{}
		if err := json.Unmarshal(*c.Error, jsonErr); err != nil {
			return &Error{
				Code:    E_SERVER,
				Message: string(*c.Error),
			}
		}
		return jsonErr
	}

	if c.Result == nil {
		return ErrNullResult
	}

	return json.Unmarshal(*c.Result, reply)
}

// DecodeClientResponse decodes the response body of a client request into
// the interface reply.
func DecodeClientResponse(r io.Reader, reply interface{}) error {
	var c clientResponse
	if err := json.NewDecoder(r).Decode(&c); err != nil {
		return err
	}
	return c.decode(reply)
}

type ErrorCode int

const (
	E_PARSE       ErrorCode = -32700
	E_INVALID_REQ ErrorCode = -32600
	E_NO_METHOD   ErrorCode = -32601
	E_BAD_PARAMS  ErrorCode = -32602
	E_INTERNAL    ErrorCode = -32603
	E_SERVER      ErrorCode = -32000
)

var ErrNullResult = errors.New("result is null")

type Error struct {
	// A Number that indicates the error type that occurred.
	Code ErrorCode `json:"code"` /* required */

	// A String providing a short description of the error.
	// The message SHOULD be limited to a concise single sentence.
	Message string `json:"message"` /* required */

	// A Primitive or Structured value that contains additional information about the error.
	Data interface{} `json:"data"` /* optional */
}

func (e *Error) Error() string {
	return e.Message
}
