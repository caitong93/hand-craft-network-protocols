package main

import (
	"encoding/binary"
	"io"
)

var (
	// MagicCookie contains first 4 octets of Options part.
	MagicCookie = []byte{0x63, 0x82, 0x53, 0x63}
)

// DhcpOptionFunc add an option to the end of DHCP message.
// DHCP Options: https://tools.ietf.org/html/rfc1533
type DhcpOptionFunc func(w io.Writer)

const (
	MessageType_DHCPDISCOVER uint8 = iota + 1
	MessageType_DHCPOFFER
	MessageType_DHCPREQUEST
	MessageType_DHCPDECLINE
	MessageType_DHCPACK
	MessageType_DHCPRELEASE
)

func OptionMessageType(typ uint8) DhcpOptionFunc {
	return func(w io.Writer) {
		binary.Write(w, binary.BigEndian, []byte{53, 1, typ})
	}
}
