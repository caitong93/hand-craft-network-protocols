package main

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"log"
)

func GenRandomBytes(size int) (blk []byte, err error) {
	blk = make([]byte, size)
	_, err = rand.Read(blk)
	return
}

func MustGenRandomBytes(size int) []byte {
	ret, err := GenRandomBytes(size)
	if err != nil {
		panic(err)
	}
	return ret
}

type HardwareAddressType struct {
	Htype uint8
	Hlen  uint8
}

const (
	OpCode_BOOTREQUEST uint8 = 0x01
	OpCode_BOOTREPLY   uint8 = 0x02
)

var (
	HardwareAddressTypeEthernet = HardwareAddressType{0x01, 0x06}
)

// https://tools.ietf.org/html/rfc2131#section-4.1
type DhcpDiscoveryMessage struct {
	// Message op code / message type.
	Op uint8
	// Hardware address type.
	// https://tools.ietf.org/html/rfc1700
	// Number Hardware Type (hrd)                           References
	// ------ -----------------------------------           ----------
	// 		 1 Ethernet (10Mb)                                    [JBP]
	// 2 Experimental Ethernet (3Mb)                        [JBP]
	// 3 Amateur Radio AX.25                                [PXK]
	// 4 Proteon ProNET Token Ring                          [JBP]
	// 5 Chaos                                              [GXP]
	// 6 IEEE 802 Networks                                  [JBP]
	// 7 ARCNET                                             [JBP]
	// 8 Hyperchannel                                       [JBP]
	// 9 Lanstar                                             [TU]
	Htype uint8
	// Hardware address length.
	Hlen uint8
	// Client sets to zero.
	Hops uint8
	// Transaction ID, a random number chosen by the
	// client, used by the client and server to associate
	// messages and responses between a client and a
	// server.
	Xid uint32
	// Filled in by client, seconds elapsed since client
	// began address acquisition or renewal process.
	Secs  uint16
	Flags uint16
	// Client IP address.
	Ciaddr uint32
	// 'your' (client) IP address.
	Yiaddr uint32
	// IP address of next server to use in bootstrap;
	// returned in DHCPOFFER, DHCPACK by server.
	Siaddr uint32
	// Relay agent IP address, used in booting via a
	// relay agent.
	Giaddr uint32
	// Client hardware address.
	Chaddr [16]byte
	// Optional server host name, null terminated string.
	Sname [64]byte
	// Boot file name, null terminated string; "generic"
	// name or null in DHCPDISCOVER, fully qualified
	// directory-path name in DHCPOFFER.
	File [128]byte
}

// Bytes converts DhcpDiscoveryMessage to payload of L3.
func (m *DhcpDiscoveryMessage) Bytes(options []DhcpOptionFunc) []byte {
	var buf bytes.Buffer
	if err := binary.Write(&buf, binary.BigEndian, m); err != nil {
		panic(err)
	}

	// Write MagicCookie.
	binary.Write(&buf, binary.BigEndian, MagicCookie)

	// Write Options.
	for _, o := range options {
		o(&buf)
	}

	// Padding to word boundaries. https://tools.ietf.org/html/rfc1533#section-3.1
	nPadding := (buf.Len()+3)/4*4 - buf.Len()
	if nPadding > 0 {
		for i := 0; i < nPadding-1; i++ {
			buf.WriteByte(0)
		}
		buf.WriteByte(255)
	}

	return buf.Bytes()
}

func NewDescoveryMessage(macAddr []byte, xid uint32) *DhcpDiscoveryMessage {
	msg := &DhcpDiscoveryMessage{
		Op:    OpCode_BOOTREQUEST,
		Htype: HardwareAddressTypeEthernet.Htype,
		Hlen:  HardwareAddressTypeEthernet.Hlen,
		Xid:   xid,
	}
	copy(msg.Chaddr[0:6], macAddr)
	return msg
}

var (
	gxid uint32 = 0
)

func Discovery(conn *Conn, macAddr []byte, options ...DhcpOptionFunc) error {
	msg := NewDescoveryMessage(macAddr, gxid)
	gxid = gxid + 1
	options = append(options, OptionMessageType(MessageType_DHCPDISCOVER))
	if err := conn.Send(msg.Bytes(options)); err != nil {
		return err
	}
	log.Println("sent discovery message")
	return nil
}
