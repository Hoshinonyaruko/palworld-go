package rcon

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

// Packet sizes definitions.
const (
	PacketPaddingSize int32 = 2 // Size of Packet's padding.
	PacketHeaderSize  int32 = 8 // Size of Packet's header.

	MinPacketSize = PacketPaddingSize + PacketHeaderSize
	MaxPacketSize = 4096 + MinPacketSize
)

// Packet is a rcon packet. Both requests and responses are sent as
// TCP packets. Their payload follows the following basic structure.
type Packet struct {
	// The packet size field is a 32-bit little endian integer, representing
	// the length of the request in bytes. Note that the packet size field
	// itself is not included when determining the size of the packet,
	// so the value of this field is always 4 less than the packet's actual
	// length. The minimum possible value for packet size is 10.
	// The maximum possible value of packet size is 4096.
	// If the response is too large to fit into one packet, it will be split
	// and sent as multiple packets.
	Size int32

	// The packet id field is a 32-bit little endian integer chosen by the
	// client for each request. It may be set to any positive integer.
	// When the RemoteServer responds to the request, the response packet
	// will have the same packet id as the original request (unless it is
	// a failed SERVERDATA_AUTH_RESPONSE packet).
	// It need not be unique, but if a unique packet id is assigned,
	// it can be used to match incoming responses to their corresponding requests.
	ID int32

	// The packet type field is a 32-bit little endian integer, which indicates
	// the purpose of the packet. Its value will always be either 0, 2, or 3,
	// depending on which of the following request/response types the packet
	// represents:
	// SERVERDATA_AUTH = 3,
	// SERVERDATA_AUTH_RESPONSE = 2,
	// SERVERDATA_EXECCOMMAND = 2,
	// SERVERDATA_RESPONSE_VALUE = 0.
	Type int32

	// The packet body field is a null-terminated string encoded in ASCII
	// (i.e. ASCIIZ). Depending on the packet type, it may contain either the
	// RCON MockPassword for the RemoteServer, the command to be executed,
	// or the RemoteServer's response to a request.
	body []byte
}

// NewPacket creates and initializes a new Packet using packetType,
// packetID and body as its initial contents. NewPacket is intended to
// calculate packet size from body length and 10 bytes for rcon headers
// and termination strings.
func NewPacket(packetType int32, packetID int32, body string) *Packet {
	size := len([]byte(body)) + int(PacketHeaderSize+PacketPaddingSize)

	return &Packet{
		Size: int32(size),
		Type: packetType,
		ID:   packetID,
		body: []byte(body),
	}
}

// Body returns packet bytes body as a string.
func (packet *Packet) Body() string {
	return string(packet.body)
}

// WriteTo implements io.WriterTo for write a packet to w.
func (packet *Packet) WriteTo(w io.Writer) (int64, error) {
	buffer := bytes.NewBuffer(make([]byte, 0, packet.Size+4))

	_ = binary.Write(buffer, binary.LittleEndian, packet.Size)
	_ = binary.Write(buffer, binary.LittleEndian, packet.ID)
	_ = binary.Write(buffer, binary.LittleEndian, packet.Type)

	// Write command body, null terminated ASCII string and an empty ASCIIZ string.
	buffer.Write(append(packet.body, 0x00, 0x00))

	return buffer.WriteTo(w)
}

// ReadFrom implements io.ReaderFrom for read a packet from r.
func (packet *Packet) ReadFrom(r io.Reader) (int64, error) {
	var n int64

	if err := binary.Read(r, binary.LittleEndian, &packet.Size); err != nil {
		return n, fmt.Errorf("rcon: read packet size: %w", err)
	}

	n += 4

	if packet.Size < MinPacketSize {
		return n, ErrResponseTooSmall
	}

	if err := binary.Read(r, binary.LittleEndian, &packet.ID); err != nil {
		return n, fmt.Errorf("rcon: read packet id: %w", err)
	}

	n += 4

	if err := binary.Read(r, binary.LittleEndian, &packet.Type); err != nil {
		return n, fmt.Errorf("rcon: read packet type: %w", err)
	}

	n += 4

	// String can actually include null characters which is the case in
	// response to a SERVERDATA_RESPONSE_VALUE packet.
	packet.body = make([]byte, packet.Size-PacketHeaderSize)

	var i int32
	for i < packet.Size-PacketHeaderSize {
		var m int
		var err error

		if m, err = r.Read(packet.body[i:]); err != nil {
			return n + int64(m) + int64(i), fmt.Errorf("rcon: %w", err)
		}

		i += int32(m)
	}

	n += int64(i)

	// Remove null terminated strings from response body.
	if !bytes.Equal(packet.body[len(packet.body)-int(PacketPaddingSize):], []byte{0x00, 0x00}) {
		return n, ErrInvalidPacketPadding
	}

	packet.body = packet.body[0 : len(packet.body)-int(PacketPaddingSize)]

	return n, nil
}
