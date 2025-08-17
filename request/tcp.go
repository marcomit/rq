// Copyright (c) 2025 Marco Menegazzi
// Licensed under the BSD 3-Clause License.
// See the LICENSE file in the project root for full license information.
package request

import (
	"fmt"
	"net"
	"strings"
)

var EMPTY_TCP_MESSAGE = fmt.Errorf("The request should contain at least one line (the connection url)")
var SOCKET_CONNECTION_REFUSED = fmt.Errorf("Connection refused")

func executeTCPRequest(content string) error {
	lines := strings.Split(content, "\n")

	if len(lines) == 0 {
		return EMPTY_TCP_MESSAGE
	}

	conn, err := net.Dial("tcp", lines[0])

	if err != nil {
		return SOCKET_CONNECTION_REFUSED
	}
	defer conn.Close()

	return nil
}
