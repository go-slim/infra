package jwt

import (
	"strings"

	"go-slim.dev/slim"
)

type Finder func(c slim.Context) string

func DefaultFinder(c slim.Context) string {
	if c.IsWebSocket() {
		return FromWebSocket(c)
	}
	if s := FromQuery(c); s != "" {
		return s
	}
	if s := FromHeader(c); s != "" {
		return s
	}
	return FromCookie(c)
}

func FromCookie(c slim.Context, keys ...string) string {
	for _, key := range keys {
		cookie, err := c.Cookie(key)
		if err == nil {
			return cookie.Value
		}
	}
	cookie, err := c.Cookie("jwt")
	if err != nil {
		return ""
	}
	return cookie.Value
}

func FromHeader(c slim.Context) string {
	bearer := c.Header("Authorization")
	if len(bearer) > 7 && strings.ToUpper(bearer[0:6]) == "BEARER" {
		return bearer[7:]
	}
	return ""
}

func FromQuery(c slim.Context, keys ...string) string {
	for _, key := range keys {
		s := c.QueryParam(key)
		if s != "" {
			return s
		}
	}
	return c.QueryParam("jwt")
}

func FromWebSocket(c slim.Context) string {
	// parses the subprotocols requested by the client in the
	// Sec-Websocket-Protocol header.
	h := strings.TrimSpace(c.Header("Sec-Websocket-Protocol"))
	if h == "" {
		return ""
	}

	protocols := strings.Split(h, ",")
	for i := range protocols {
		protocols[i] = strings.TrimSpace(protocols[i])
	}

	// 查找子协议
	for _, sub := range protocols {
		if len(sub) > 7 && strings.ToUpper(sub[0:6]) == "BEARER" {
			return sub[7:]
		}
	}

	return ""
}
