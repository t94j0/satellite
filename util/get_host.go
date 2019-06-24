package util

import (
	"net"
	"strings"

	"github.com/t94j0/ja3-server/net/http"
)

func GetHost(req *http.Request) net.IP {
	inp := req.RemoteAddr
	split := strings.Split(inp, ":")
	filter := split[:len(split)-1]
	full := strings.Join(filter, ":")
	trimmedr := strings.TrimRight(full, "]")
	trimmed := strings.TrimLeft(trimmedr, "[")
	return net.ParseIP(trimmed)
}
