package main

import (
	"net"
)

// ClientID is a client identification struct
type ClientID struct {
	paths map[string][]string
}

// NewClientID creates a new ClientID object
func NewClientID() *ClientID {
	return &ClientID{
		paths: make(map[string][]string),
	}
}

// Hit notifies ClientID that an IP hit a target
func (c *ClientID) Hit(ip net.IP, path string) {
	ipstr := ip.String()
	c.paths[ipstr] = append(c.paths[ipstr], path)
}

// Match asks ClientID if the target IP has succeeded in hitting the prereqs
func (c *ClientID) Match(ip net.IP, list []string) bool {
	ipstr := ip.String()
	matchList, ok := c.paths[ipstr]
	if !ok {
		return false
	}
	if len(list) != len(matchList) {
		return false
	}

	for i := range list {
		if list[i] != matchList[i] {
			return false
		}
	}

	return true
}
