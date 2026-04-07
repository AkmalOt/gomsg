package gomsg

import (
	"encoding/binary"
	"fmt"
	"strings"
	"time"
)

// storageContext determines the header size of a __properties_version1.0 stream.
type storageContext int

const (
	contextMessage  storageContext = iota // 32-byte header (root message)
	contextEmbed                          // 24-byte header (embedded message)
	contextSubItem                        // 8-byte header (attachment/recipient)
)

// headerSize returns the number of header bytes before property entries.
func (c storageContext) headerSize() int {
	switch c {
	case contextMessage:
		return 32
	case contextEmbed:
		return 24
	case contextSubItem:
		return 8
	default:
		return 8
	}
}

// Property represents a single MAPI property with its parsed value.
type Property struct {
	ID    PropertyID
	Type  PropertyType
	Flags uint32
	Value interface{}
}

// PropertyStore holds all parsed MAPI properties from a properties stream.
type PropertyStore struct {
	props map[PropertyID]*Property

	// Header fields from root/embedded message context.
	NextRecipientID  uint32
	NextAttachmentID uint32
	RecipientCount   uint32
	AttachmentCount  uint32
}

// newPropertyStore creates an empty property store.
func newPropertyStore() *PropertyStore {
	return &PropertyStore{
		props: make(map[PropertyID]*Property),
	}
}

// Get returns a property by its MAPI ID, or nil if not found.
func (ps *PropertyStore) Get(id PropertyID) *Property {
	if ps == nil {
		return nil
	}
	return ps.props[id]
}

// GetString returns a string property value, or empty string if missing.
func (ps *PropertyStore) GetString(id PropertyID) string {
	p := ps.Get(id)
	if p == nil {
		return ""
	}
	switch v := p.Value.(type) {
	case string:
		return v
	case []byte:
		return string(v)
	default:
		return ""
	}
}

// GetTime returns a time.Time property value.
func (ps *PropertyStore) GetTime(id PropertyID) (time.Time, bool) {
	p := ps.Get(id)
	if p == nil {
		return time.Time{}, false
	}
	if t, ok := p.Value.(time.Time); ok {
		return t, true
	}
	return time.Time{}, false
}

// GetBytes returns a binary property value.
func (ps *PropertyStore) GetBytes(id PropertyID) []byte {
	p := ps.Get(id)
	if p == nil {
		return nil
	}
	if b, ok := p.Value.([]byte); ok {
		return b
	}
	return nil
}

// GetInt32 returns an int32 property value.
func (ps *PropertyStore) GetInt32(id PropertyID) (int32, bool) {
	p := ps.Get(id)
	if p == nil {
		return 0, false
	}
	switch v := p.Value.(type) {
	case int32:
		return v, true
	case int16:
		return int32(v), true
	default:
		return 0, false
	}
}

// GetBool returns a boolean property value.
func (ps *PropertyStore) GetBool(id PropertyID) (bool, bool) {
	p := ps.Get(id)
	if p == nil {
		return false, false
	}
	if b, ok := p.Value.(bool); ok {
		return b, true
	}
	return false, false
}

// All returns all property IDs present in the store.
func (ps *PropertyStore) All() []PropertyID {
	if ps == nil {
		return nil
	}
	ids := make([]PropertyID, 0, len(ps.props))
	for id := range ps.props {
		ids = append(ids, id)
	}
	return ids
}

// parsePropertyStore parses the __properties_version1.0 stream and resolves
// variable-length property values from sibling __substg1.0_* streams.
//
// streamLookup retrieves a sibling stream's raw data by its name within
// the same storage (e.g., "__substg1.0_0037001F" for the Unicode subject).
func parsePropertyStore(
	data []byte,
	ctx storageContext,
	streamLookup func(name string) ([]byte, bool),
	codepage uint32,
) (*PropertyStore, error) {
	ps := newPropertyStore()
	hdrSize := ctx.headerSize()

	if len(data) < hdrSize {
		return ps, fmt.Errorf("%w: stream too short for header", ErrNoProperties)
	}

	// Parse header fields for message contexts.
	if ctx == contextMessage || ctx == contextEmbed {
		offset := 8 // skip 8 reserved bytes
		if len(data) >= offset+16 {
			ps.NextRecipientID = binary.LittleEndian.Uint32(data[offset:])
			ps.NextAttachmentID = binary.LittleEndian.Uint32(data[offset+4:])
			ps.RecipientCount = binary.LittleEndian.Uint32(data[offset+8:])
			ps.AttachmentCount = binary.LittleEndian.Uint32(data[offset+12:])
		}
	}

	// Parse 16-byte property entries after the header.
	entries := data[hdrSize:]
	entryCount := len(entries) / 16

	for i := 0; i < entryCount; i++ {
		entry := entries[i*16 : (i+1)*16]

		propType := PropertyType(binary.LittleEndian.Uint16(entry[0:2]))
		propID := PropertyID(binary.LittleEndian.Uint16(entry[2:4]))
		flags := binary.LittleEndian.Uint32(entry[4:8])
		valueBytes := entry[8:16]

		var val interface{}
		var err error

		if isFixedLength(propType) {
			val, err = convertValue(propType, valueBytes, codepage)
			if err != nil {
				continue // skip malformed properties
			}
		} else {
			// Variable-length: look up the __substg1.0_XXXXYYYY stream.
			streamName := fmt.Sprintf("__substg1.0_%04X%04X",
				uint16(propID), uint16(propType))
			streamName = strings.ToUpper(streamName)

			streamData, ok := streamLookup(streamName)
			if !ok {
				continue // stream not found, skip
			}

			val, err = convertValue(propType, streamData, codepage)
			if err != nil {
				continue
			}
		}

		ps.props[propID] = &Property{
			ID:    propID,
			Type:  propType,
			Flags: flags,
			Value: val,
		}
	}

	return ps, nil
}
