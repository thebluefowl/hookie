package listen

import (
	"net"
	"net/http"
	"net/textproto"
	"net/url"
	"strings"

	"golang.org/x/net/http/httpguts"
)

// Hop-by-hop headers. These are removed when sent to the backend.
// As of RFC 7230, hop-by-hop headers are required to appear in the
// Connection header field. These are the headers defined by the
// obsoleted RFC 2616 (section 13.5.1) and are used for backward
// compatibility.
var hopHeaders = []string{
	"Connection",
	"Proxy-Connection", // non-standard but still sent by libcurl and rejected by e.g. google
	"Keep-Alive",
	"Proxy-Authenticate",
	"Proxy-Authorization",
	"Te",      // canonicalized version of "TE"
	"Trailer", // not Trailers per URL above; https://www.rfc-editor.org/errata_search.php?eid=4522
	"Transfer-Encoding",
	"Upgrade",
}

type Server struct {
	Transport http.RoundTripper
}

func (s *Server) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if s.Transport == nil {
		s.Transport = http.DefaultTransport
	}

	out := outRequest(req)

	if out.Body != nil {
		defer out.Body.Close()
	}

	// we should not handle upgrade requests as this server is only meant to
	// handle webhook requests
	if isUpgrade(out.Header) {
		rw.WriteHeader(http.StatusNotImplemented)
		return
	}
}

func outRequest(in *http.Request) *http.Request {
	ctx := in.Context()

	out := in.Clone(ctx)
	if in.ContentLength == 0 {
		out.Body = nil
	}

	out.Close = false

	removeHopByHopHeaders(out.Header)

	cleanXForwarded(out)
	setXForwarded(in, out)

	out.URL.RawQuery = cleanQueryParams(out.URL.RawQuery)

	if _, ok := out.Header["User-Agent"]; !ok {
		out.Header.Set("User-Agent", "")
	}

	return out
}

func isUpgrade(h http.Header) bool {
	return httpguts.HeaderValuesContainsToken(h["Connection"], "Upgrade")
}

// removeHopByHopHeaders removes hop-by-hop headers.
func removeHopByHopHeaders(h http.Header) {
	// RFC 7230, section 6.1: Remove headers listed in the "Connection" header.
	for _, f := range h["Connection"] {
		for _, sf := range strings.Split(f, ",") {
			if sf = textproto.TrimString(sf); sf != "" {
				h.Del(sf)
			}
		}
	}
	// RFC 2616, section 13.5.1: Remove a set of known hop-by-hop headers.
	// This behavior is superseded by the RFC 7230 Connection header, but
	// preserve it for backwards compatibility.
	for _, f := range hopHeaders {
		h.Del(f)
	}
}

// cleanXForwarded removes X-Forwarded-* headers before new values are set
// using setXForwarded.
func cleanXForwarded(req *http.Request) {
	req.Header.Del("Forwarded")
	req.Header.Del("X-Forwarded-For")
	req.Header.Del("X-Forwarded-Host")
	req.Header.Del("X-Forwarded-Proto")
}

// setXForwarded sets X-Forwarded-* headers based on the incoming request.
func setXForwarded(in *http.Request, out *http.Request) {
	clientIP, _, err := net.SplitHostPort(in.RemoteAddr)
	if err == nil {
		prior := out.Header["X-Forwarded-For"]
		if len(prior) > 0 {
			clientIP = strings.Join(prior, ", ") + ", " + clientIP
		}
		out.Header.Set("X-Forwarded-For", clientIP)
	} else {
		out.Header.Del("X-Forwarded-For")
	}
	out.Header.Set("X-Forwarded-Host", in.Host)
	if in.TLS == nil {
		out.Header.Set("X-Forwarded-Proto", "http")
	} else {
		out.Header.Set("X-Forwarded-Proto", "https")
	}
}

// cleanQueryParams sanitizes and reencodes the query parameters.  This handles
// some edge cases where the query parameters are not properly encoded.
func cleanQueryParams(s string) string {
	reencode := func(s string) string {
		v, _ := url.ParseQuery(s)
		return v.Encode()
	}
	for i := 0; i < len(s); {
		switch s[i] {
		case ';':
			return reencode(s)
		case '%':
			if i+2 >= len(s) || !ishex(s[i+1]) || !ishex(s[i+2]) {
				return reencode(s)
			}
			i += 3
		default:
			i++
		}
	}
	return s
}

// ishex reports whether c is a valid hexadecimal digit.
func ishex(c byte) bool {
	switch {
	case '0' <= c && c <= '9':
		return true
	case 'a' <= c && c <= 'f':
		return true
	case 'A' <= c && c <= 'F':
		return true
	}
	return false
}
