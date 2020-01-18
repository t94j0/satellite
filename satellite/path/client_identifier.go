package path

import (
	"net"
)

// ClientID is client identification
type ClientID struct {
	list map[string][]string
}

// NewClientID creates a new ClientID object
func NewClientID() *ClientID {
	return &ClientID{
		list: make(map[string][]string),
	}
}

// Hit notifies ClientID that an IP hit a target
func (c *ClientID) Hit(ip net.IP, path string) {
	ipstr := ip.String()
	c.list[ipstr] = append(c.list[ipstr], path)
}

// Match asks ClientID if the target IP has succeeded in hitting the prereqs
func (c *ClientID) Match(ip net.IP, targetList []string) bool {
	if len(targetList) == 0 {
		return true
	}

	ipstr := ip.String()
	list, ok := c.list[ipstr]
	if !ok {
		return false
	}

	if len(targetList) > len(list) {
		return false
	}

	lastSubset := list[len(list)-len(targetList) : len(list)]

	for i := range lastSubset {
		if lastSubset[i] != targetList[i] {
			return false
		}
	}

	return true
}
