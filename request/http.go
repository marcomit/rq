// Copyright (c) 2025 Marco Menegazzi
// Licensed under the BSD 3-Clause License.
// See the LICENSE file in the project root for full license information.
package request

import (
	"net/http"
)

type HttpRequestMethod string

const (
	HttpGet    HttpRequestMethod = "GET"
	HttpPost   HttpRequestMethod = "POST"
	HttpPut    HttpRequestMethod = "PUT"
	HttpPatch  HttpRequestMethod = "PATCH"
	HttpDelete HttpRequestMethod = "DELETE"
)

type HttpRequest struct {
	Method  HttpRequestMethod
	Url     string
	Headers map[string]string
	Body    string
}

func (request HttpRequest) Request() {
	_, err := http.Get(request.Url)
	if err != nil {

	}
}
