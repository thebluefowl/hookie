package proxyutils

import (
	"bytes"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
)

type TargetRequest struct {
	Request *http.Request
}

func NewTargetRequest(in *http.Request, target *url.URL) (*TargetRequest, error) {
	ctx := in.Context()
	out := in.Clone(ctx)
	out.Host = in.Host

	setupURL(out, target)
	setupHeaders(out, in)

	err := copyBody(in, out)
	if err != nil {
		return nil, err
	}

	out.Close = false

	return &TargetRequest{
		Request: out,
	}, nil
}

type SerializableRequest struct {
	Headers map[string][]string
	Body    []byte
	Method  string
	URL     string
	Host    string
}

func (tr *TargetRequest) MarshalJSON() ([]byte, error) {
	payload := &SerializableRequest{
		Headers: make(map[string][]string),
		Body:    make([]byte, 0),
		Host:    tr.Request.Host,
		Method:  tr.Request.Method,
		URL:     tr.Request.URL.String(),
	}
	for k, v := range tr.Request.Header {
		payload.Headers[k] = v
	}
	defer tr.Request.Body.Close()
	buf, err := io.ReadAll(tr.Request.Body)
	if err != nil {
		return nil, err
	}
	payload.Body = buf
	return json.Marshal(payload)
}

func (tr *TargetRequest) UnmarshalJSON(data []byte) error {
	payload := &SerializableRequest{
		Headers: make(map[string][]string),
		Body:    make([]byte, 0),
	}
	if err := json.Unmarshal(data, payload); err != nil {
		return err
	}
	tr.Request = &http.Request{
		Method: payload.Method,
		URL:    &url.URL{},
		Header: make(http.Header),
	}
	for k, v := range payload.Headers {
		tr.Request.Header[k] = v
	}
	tr.Request.URL, _ = url.Parse(payload.URL)
	tr.Request.Body = io.NopCloser(bytes.NewBuffer(payload.Body))
	tr.Request.Host = payload.Host
	return nil
}

type TargetResponse struct {
	Response *http.Response
}

func NewTargetResponse(res *http.Response) *TargetResponse {
	sanitizeHeaders(res.Header)
	return &TargetResponse{
		Response: res,
	}
}

func setupURL(out *http.Request, target *url.URL) {
	targetQuery := target.RawQuery
	out.URL.Scheme = target.Scheme
	out.URL.Host = target.Host
	out.URL.Path, out.URL.RawPath = joinURLPath(target, out.URL)

	if targetQuery != "" && out.URL.RawQuery != "" {
		out.URL.RawQuery = targetQuery + "&" + out.URL.RawQuery
	} else {
		out.URL.RawQuery = targetQuery + out.URL.RawQuery
	}
}

func setupHeaders(out, in *http.Request) {
	sanitizeHeaders(out.Header)

	clientIP, _, err := net.SplitHostPort(in.RemoteAddr)
	if err == nil {
		prior := out.Header["X-Forwarded-For"]
		if len(prior) > 0 {
			clientIP = strings.Join(prior, ", ") + ", " + clientIP
		}
		out.Header.Set("X-Forwarded-For", clientIP)
	}
	out.Header.Set("X-Forwarded-Host", in.Host)
	out.Header.Set("X-Forwarded-Proto", determineProtocol(in))

	if _, ok := out.Header["User-Agent"]; !ok {
		out.Header.Set("User-Agent", "")
	}
}

func sanitizeHeaders(h http.Header) {
	headersToDelete := []string{"Forwarded", "X-Forwarded-For", "X-Forwarded-Host", "X-Forwarded-Proto"}
	for _, v := range headersToDelete {
		h.Del(v)
	}
}

func determineProtocol(r *http.Request) string {
	if r.TLS == nil {
		return "http"
	}
	return "https"
}

func copyBody(in, out *http.Request) error {
	buf, err := io.ReadAll(in.Body)
	if err != nil {
		return err
	}

	if err := in.Body.Close(); err != nil {
		return err
	}

	in.Body = io.NopCloser(bytes.NewBuffer(buf))
	out.Body = io.NopCloser(bytes.NewBuffer(buf))

	return nil
}

func joinURLPath(a, b *url.URL) (path, rawpath string) {
	if a.RawPath == "" && b.RawPath == "" {
		return singleJoiningSlash(a.Path, b.Path), ""
	}
	apath := a.EscapedPath()
	bpath := b.EscapedPath()
	return adjustPaths(a.Path, b.Path, apath, bpath)
}

func adjustPaths(aPath, bPath, aEscaped, bEscaped string) (path, rawpath string) {
	aslash := strings.HasSuffix(aEscaped, "/")
	bslash := strings.HasPrefix(bEscaped, "/")

	switch {
	case aslash && bslash:
		return aPath + bPath[1:], aEscaped + bEscaped[1:]
	case !aslash && !bslash:
		return aPath + "/" + bPath, aEscaped + "/" + bEscaped
	}
	return aPath + bPath, aEscaped + bEscaped
}

func singleJoiningSlash(a, b string) string {
	aslash := strings.HasSuffix(a, "/")
	bslash := strings.HasPrefix(b, "/")
	switch {
	case aslash && bslash:
		return a + b[1:]
	case !aslash && !bslash:
		return a + "/" + b
	}
	return a + b
}

func CopyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}
