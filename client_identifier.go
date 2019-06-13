package main

import (
	"net"
)

type ClientID struct {
	ids map[string][]uint
}

func NewClientID() *ClientID {
	return &ClientID{
		ids: make(map[string][]uint),
	}
}

func (c *ClientID) Hit(ip net.IP, id uint) {
	ipstr := ip.String()
	c.ids[ipstr] = append(c.ids[ipstr], id)
}

func (c *ClientID) Match(ip net.IP, list []uint) bool {
	ipstr := ip.String()
	matchList, ok := c.ids[ipstr]
	if !ok {
		return false
	}
	if len(list) != len(matchList) {
		return false
	}

	for i, _ := range list {
		if list[i] != matchList[i] {
			return false
		}
	}

	return true
}
