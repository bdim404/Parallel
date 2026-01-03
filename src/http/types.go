package http

type HTTPRequest struct {
	Raw     []byte
	Method  string
	URI     string
	Version string
	Headers map[string][]string
	Body    []byte
}

type HTTPResponse struct {
	Raw        []byte
	StatusCode int
	Status     string
	Version    string
	Headers    map[string][]string
	Body       []byte
}

func (r *HTTPRequest) ShouldClose() bool {
	if r.Version == "HTTP/1.0" {
		return !hasKeepAlive(r.Headers)
	}
	return hasClose(r.Headers)
}

func (r *HTTPResponse) ShouldClose() bool {
	if r.Version == "HTTP/1.0" {
		return !hasKeepAlive(r.Headers)
	}
	return hasClose(r.Headers)
}

func hasKeepAlive(headers map[string][]string) bool {
	for _, v := range headers["Connection"] {
		if v == "keep-alive" {
			return true
		}
	}
	return false
}

func hasClose(headers map[string][]string) bool {
	for _, v := range headers["Connection"] {
		if v == "close" {
			return true
		}
	}
	return false
}
