package http

import (
	"bufio"
	"bytes"
)

var httpMethods = [][]byte{
	[]byte("GET "),
	[]byte("POST "),
	[]byte("PUT "),
	[]byte("DELETE "),
	[]byte("HEAD "),
	[]byte("OPTIONS "),
	[]byte("PATCH "),
	[]byte("CONNECT "),
	[]byte("TRACE "),
}

func DetectProtocol(reader *bufio.Reader) (isHTTP bool, err error) {
	peek, err := reader.Peek(16)
	if err != nil {
		return false, err
	}

	for _, method := range httpMethods {
		if bytes.HasPrefix(peek, method) {
			return true, nil
		}
	}

	return false, nil
}
