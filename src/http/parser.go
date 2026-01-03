package http

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"
)

func ParseRequest(reader *bufio.Reader) (*HTTPRequest, error) {
	var raw bytes.Buffer
	teeReader := io.TeeReader(reader, &raw)
	bufferedTee := bufio.NewReader(teeReader)

	line, err := bufferedTee.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("read request line: %w", err)
	}

	parts := strings.SplitN(strings.TrimSpace(line), " ", 3)
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid request line: %s", line)
	}

	req := &HTTPRequest{
		Method:  parts[0],
		URI:     parts[1],
		Version: parts[2],
		Headers: make(map[string][]string),
	}

	headers, err := readHeaders(bufferedTee)
	if err != nil {
		return nil, fmt.Errorf("read headers: %w", err)
	}
	req.Headers = headers

	body, err := readBody(bufferedTee, headers, req.Method)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}
	req.Body = body

	req.Raw = raw.Bytes()
	return req, nil
}

func ParseResponse(reader *bufio.Reader) (*HTTPResponse, error) {
	var raw bytes.Buffer
	teeReader := io.TeeReader(reader, &raw)
	bufferedTee := bufio.NewReader(teeReader)

	line, err := bufferedTee.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("read status line: %w", err)
	}

	parts := strings.SplitN(strings.TrimSpace(line), " ", 3)
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid status line: %s", line)
	}

	statusCode, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, fmt.Errorf("parse status code: %w", err)
	}

	status := parts[1]
	if len(parts) == 3 {
		status = parts[1] + " " + parts[2]
	}

	resp := &HTTPResponse{
		Version:    parts[0],
		StatusCode: statusCode,
		Status:     status,
		Headers:    make(map[string][]string),
	}

	headers, err := readHeaders(bufferedTee)
	if err != nil {
		return nil, fmt.Errorf("read headers: %w", err)
	}
	resp.Headers = headers

	if shouldHaveBody(statusCode, "") {
		body, err := readBody(bufferedTee, headers, "")
		if err != nil {
			return nil, fmt.Errorf("read body: %w", err)
		}
		resp.Body = body
	}

	resp.Raw = raw.Bytes()
	return resp, nil
}

func readHeaders(reader *bufio.Reader) (map[string][]string, error) {
	headers := make(map[string][]string)

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return nil, err
		}

		line = strings.TrimRight(line, "\r\n")

		if line == "" {
			break
		}

		colonIndex := strings.Index(line, ":")
		if colonIndex == -1 {
			continue
		}

		key := strings.TrimSpace(line[:colonIndex])
		value := strings.TrimSpace(line[colonIndex+1:])

		headers[key] = append(headers[key], value)
	}

	return headers, nil
}

func readBody(reader *bufio.Reader, headers map[string][]string, method string) ([]byte, error) {
	if !shouldHaveBody(0, method) && method != "" {
		return nil, nil
	}

	if isChunked(headers) {
		return readChunkedBody(reader)
	}

	contentLength := getContentLength(headers)
	if contentLength > 0 {
		body := make([]byte, contentLength)
		_, err := io.ReadFull(reader, body)
		if err != nil {
			return nil, err
		}
		return body, nil
	}

	return nil, nil
}

func isChunked(headers map[string][]string) bool {
	for _, v := range headers["Transfer-Encoding"] {
		if strings.Contains(strings.ToLower(v), "chunked") {
			return true
		}
	}
	return false
}

func getContentLength(headers map[string][]string) int64 {
	for _, v := range headers["Content-Length"] {
		length, err := strconv.ParseInt(v, 10, 64)
		if err == nil {
			return length
		}
	}
	return 0
}

func readChunkedBody(reader *bufio.Reader) ([]byte, error) {
	var body bytes.Buffer

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return nil, err
		}

		line = strings.TrimSpace(line)
		chunkSize, err := strconv.ParseInt(line, 16, 64)
		if err != nil {
			return nil, fmt.Errorf("parse chunk size: %w", err)
		}

		if chunkSize == 0 {
			reader.ReadString('\n')
			break
		}

		chunk := make([]byte, chunkSize)
		_, err = io.ReadFull(reader, chunk)
		if err != nil {
			return nil, err
		}

		body.Write(chunk)

		reader.ReadString('\n')
	}

	return body.Bytes(), nil
}

func shouldHaveBody(statusCode int, method string) bool {
	if method == "HEAD" || method == "GET" || method == "DELETE" {
		return false
	}

	if statusCode >= 100 && statusCode < 200 {
		return false
	}
	if statusCode == 204 || statusCode == 304 {
		return false
	}

	return true
}
