package rcon

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"testing"
)

func TestNewPacket(t *testing.T) {
	body := []byte("testdata")
	packet := NewPacket(SERVERDATA_RESPONSE_VALUE, 42, string(body))

	if packet.Body() != string(body) {
		t.Errorf("%q, want %q", packet.Body(), body)
	}

	want := int32(len([]byte(body))) + PacketHeaderSize + PacketPaddingSize
	if packet.Size != want {
		t.Errorf("got %d, want %d", packet.Size, want)
	}
}

func TestPacket_WriteTo(t *testing.T) {
	t.Run("check bytes written", func(t *testing.T) {
		body := []byte("testdata")
		packet := NewPacket(SERVERDATA_RESPONSE_VALUE, 42, string(body))

		var buffer bytes.Buffer
		n, err := packet.WriteTo(&buffer)
		if err != nil {
			t.Fatal(err)
		}

		wantN := packet.Size + 4
		if n != int64(wantN) {
			t.Errorf("got %d, want %d", n, int64(wantN))
		}
	})
}

func TestPacket_ReadFrom(t *testing.T) {
	t.Run("check read", func(t *testing.T) {
		body := []byte("testdata")
		packetWant := NewPacket(SERVERDATA_RESPONSE_VALUE, 42, string(body))

		var buffer bytes.Buffer
		nWant, err := packetWant.WriteTo(&buffer)
		if err != nil {
			t.Fatal(err)
		}

		packetGot := new(Packet)
		nGot, err := packetGot.ReadFrom(&buffer)
		if err != nil {
			t.Fatal(err)
		}

		if nGot != nWant {
			t.Fatalf("got %d, want %d", nGot, nWant)
		}

		if packetGot.Body() != packetWant.Body() {
			t.Fatalf("got %q, want %q", packetGot.body, packetWant.body)
		}
	})

	t.Run("EOF", func(t *testing.T) {
		var buffer bytes.Buffer

		packetGot := new(Packet)
		nGot, err := packetGot.ReadFrom(&buffer)
		if !errors.Is(err, io.EOF) {
			t.Fatalf("got %q, want %q", err, io.EOF)
		}

		if nGot != 0 {
			t.Fatalf("got %d, want %d", nGot, 0)
		}
	})

	t.Run("response too small", func(t *testing.T) {
		var buffer bytes.Buffer

		packetGot := new(Packet)
		binary.Write(&buffer, binary.LittleEndian, packetGot.Size)

		_, err := packetGot.ReadFrom(&buffer)
		if !errors.Is(err, ErrResponseTooSmall) {
			t.Fatalf("got %q, want %q", err, ErrResponseTooSmall)
		}
	})

	t.Run("EOF 2", func(t *testing.T) {
		var buffer bytes.Buffer
		binary.Write(&buffer, binary.LittleEndian, int32(18))

		packetGot := new(Packet)
		nGot, err := packetGot.ReadFrom(&buffer)
		if !errors.Is(err, io.EOF) {
			t.Fatalf("got %q, want %q", err, io.EOF)
		}

		if nGot != 4 {
			t.Fatalf("got %d, want %d", nGot, 4)
		}
	})

	t.Run("EOF 3", func(t *testing.T) {
		var buffer bytes.Buffer
		binary.Write(&buffer, binary.LittleEndian, int32(18))
		binary.Write(&buffer, binary.LittleEndian, int32(42))

		packetGot := new(Packet)
		nGot, err := packetGot.ReadFrom(&buffer)
		if !errors.Is(err, io.EOF) {
			t.Fatalf("got %q, want %q", err, io.EOF)
		}

		if nGot != 8 {
			t.Fatalf("got %d, want %d", nGot, 8)
		}
	})

	t.Run("padding", func(t *testing.T) {
		body := []byte("testdata")
		packetWant := NewPacket(SERVERDATA_RESPONSE_VALUE, 42, string(body))
		packetWant.Size = 10
		var buffer bytes.Buffer
		_, err := packetWant.WriteTo(&buffer)
		if err != nil {
			t.Fatal(err)
		}

		packetGot := new(Packet)
		_, err = packetGot.ReadFrom(&buffer)
		if !errors.Is(err, ErrInvalidPacketPadding) {
			t.Fatalf("got %q, want %q", err, ErrInvalidPacketPadding)
		}
	})

	t.Run("EOF 4", func(t *testing.T) {
		var buffer bytes.Buffer
		binary.Write(&buffer, binary.LittleEndian, int32(18))
		binary.Write(&buffer, binary.LittleEndian, int32(42))
		buffer.Write(append([]byte("testdata"), 0x00, 0x00))

		packetGot := new(Packet)
		nGot, err := packetGot.ReadFrom(&buffer)
		if !errors.Is(err, io.EOF) {
			t.Fatalf("got %q, want %q", err, io.EOF)
		}

		if nGot != 18 {
			t.Fatalf("got %d, want %d", nGot, 18)
		}
	})
}
