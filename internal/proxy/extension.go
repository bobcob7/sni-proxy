package proxy

import (
	"encoding/binary"
)

type extensions []*extension

type extension struct {
	Type    uint16
	Length  uint16
	Content []byte // Length is $Length
}

func parseExtensions(buffer []byte) extensions {
	out := make([]*extension, 0)
	for i := 0; i < len(buffer); {
		newExt := extension{}
		// Parse type
		newExt.Type = binary.BigEndian.Uint16(buffer[i : i+2])
		i += 2
		// Parse type
		newExt.Length = binary.BigEndian.Uint16(buffer[i : i+2])
		i += 2
		// Parse content
		newExt.Content = buffer[i : i+int(newExt.Length)]
		i += int(newExt.Length)
		out = append(out, &newExt)
	}
	return out
}

// getServerName finds the serverName extension and return the hostname value from there.
func (e extensions) getServerName() string {
	var serverNameExtension *extension
	for i, ext := range e {
		if ext.Type == 0 {
			// Is ServerName extension
			serverNameExtension = e[i]
		}
	}
	if serverNameExtension == nil {
		return ""
	}
	buffer := serverNameExtension.Content
	// Parse entries in extension
	listLen := binary.BigEndian.Uint16(buffer[0:2])
	buffer = buffer[2:]
	for i := uint16(0); i < listLen; i++ {
		entryType := uint8(buffer[i])
		i += 1
		entryLen := binary.BigEndian.Uint16(buffer[i : i+2])
		i += 2
		if entryType == TypeServerNameHostNameExt {
			// Exit with value if entry is the hostname
			return string(buffer[i : i+entryLen])
		}
		i += entryLen
	}
	return ""
}

const TypeServerNameHostNameExt = 0
