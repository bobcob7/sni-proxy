package proxy

import (
	"encoding/binary"
	"errors"
	"time"
)

const (
	typeTCPHandshake = 0x16
)

type helloClientHandshake struct {
	Type    uint8
	Length  uint32 // Actually uint24
	Version uint16
	Random  struct {
		Timestamp time.Time // uint32
		Buffer    []byte    // 28 bytes
	}
	SessionID         uint8
	CipherSuiteLength uint16
	// Skip next $CipherSuiteLength
	CompressionMethods uint8
	// Skip next $CompressionMethods
	ExtentionsLength uint16
	Extensions       extensions
}

func newHelloClientHandshake(buffer []byte) (out helloClientHandshake) {
	out.Type = uint8(buffer[0])
	length := buffer[0:4]
	// length[0] = 0
	out.Length = binary.BigEndian.Uint32(length)
	out.Version = binary.BigEndian.Uint16(buffer[5:7])
	out.Random.Timestamp = time.Unix(int64(binary.BigEndian.Uint32(buffer[7:11])), 0)
	out.Random.Buffer = buffer[11:38]
	out.SessionID = uint8(buffer[38])
	out.CipherSuiteLength = binary.BigEndian.Uint16(buffer[39:41])
	// Skip the Cipher suits
	bufferPointer := 41 + out.CipherSuiteLength
	out.CompressionMethods = uint8(buffer[bufferPointer])
	// Skip the Compression methods
	bufferPointer += uint16(out.CompressionMethods) + 1
	out.ExtentionsLength = binary.BigEndian.Uint16(buffer[bufferPointer : bufferPointer+2])
	extensionBuffer := buffer[bufferPointer+2 : bufferPointer+2+out.ExtentionsLength]
	out.Extensions = parseExtensions(extensionBuffer)
	return
}

var (
	ErrHandshakeNotFound = errors.New("handshake not found")
)

// GetSNI will return back the valid SNI from a TLS handshake.
func GetSNI(buffer []byte) (string, error) {
	// Check if first part is x16. Which indicates that it's a handshake
	// Try to parse in hello packet
	tcpType := buffer[0]
	if tcpType != typeTCPHandshake {
		return "", ErrHandshakeNotFound
	}
	handshakeLength := binary.BigEndian.Uint16(buffer[3:5])
	handshake := buffer[5 : 5+handshakeLength]
	hs := newHelloClientHandshake(handshake)
	// Look for ServerName extension
	return hs.Extensions.getServerName(), nil
}
