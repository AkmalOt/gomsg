package gomsg

import (
	"encoding/binary"
	"fmt"
	"math"
	"time"
)

// PropertyType represents a MAPI property type tag.
type PropertyType uint16

// Standard MAPI property types.
const (
	TypeUnspecified PropertyType = 0x0000
	TypeNull       PropertyType = 0x0001
	TypeInt16      PropertyType = 0x0002
	TypeInt32      PropertyType = 0x0003
	TypeFloat32    PropertyType = 0x0004
	TypeFloat64    PropertyType = 0x0005
	TypeCurrency   PropertyType = 0x0006
	TypeAppTime    PropertyType = 0x0007
	TypeBoolean    PropertyType = 0x000B
	TypeObject     PropertyType = 0x000D
	TypeInt64      PropertyType = 0x0014
	TypeString8    PropertyType = 0x001E
	TypeUnicode    PropertyType = 0x001F
	TypeSysTime    PropertyType = 0x0040
	TypeGUID       PropertyType = 0x0048
	TypeBinary     PropertyType = 0x0102

	// Multi-valued types.
	TypeMultiInt16   PropertyType = 0x1002
	TypeMultiInt32   PropertyType = 0x1003
	TypeMultiFloat32 PropertyType = 0x1004
	TypeMultiFloat64 PropertyType = 0x1005
	TypeMultiInt64   PropertyType = 0x1014
	TypeMultiString8 PropertyType = 0x101E
	TypeMultiUnicode PropertyType = 0x101F
	TypeMultiSysTime PropertyType = 0x1040
	TypeMultiBinary  PropertyType = 0x1102
)

// isFixedLength returns true if the property value fits within 8 bytes inline.
func isFixedLength(t PropertyType) bool {
	switch t {
	case TypeInt16, TypeInt32, TypeFloat32, TypeFloat64,
		TypeCurrency, TypeAppTime, TypeBoolean, TypeInt64, TypeSysTime:
		return true
	default:
		return false
	}
}

// fixedSize returns the byte width of a fixed-length property type.
func fixedSize(t PropertyType) int {
	switch t {
	case TypeInt16:
		return 2
	case TypeInt32, TypeFloat32, TypeBoolean:
		return 4
	case TypeFloat64, TypeCurrency, TypeAppTime, TypeInt64, TypeSysTime:
		return 8
	default:
		return 0
	}
}

// filetimeToTime converts a Windows FILETIME (100-ns intervals since 1601-01-01)
// to a Go time.Time.
func filetimeToTime(ft uint64) time.Time {
	// Number of 100-ns intervals between 1601-01-01 and 1970-01-01.
	const epochDiff = 116444736000000000
	if ft <= epochDiff {
		return time.Time{}
	}
	nsec := (int64(ft) - epochDiff) * 100
	return time.Unix(0, nsec)
}

// convertValue converts raw bytes to a Go value based on the MAPI property type.
func convertValue(t PropertyType, data []byte, codepage uint32) (interface{}, error) {
	switch t {
	case TypeInt16:
		if len(data) < 2 {
			return nil, fmt.Errorf("%w: int16 needs 2 bytes, got %d", ErrPropertyType, len(data))
		}
		return int16(binary.LittleEndian.Uint16(data[:2])), nil

	case TypeInt32:
		if len(data) < 4 {
			return nil, fmt.Errorf("%w: int32 needs 4 bytes, got %d", ErrPropertyType, len(data))
		}
		return int32(binary.LittleEndian.Uint32(data[:4])), nil

	case TypeFloat32:
		if len(data) < 4 {
			return nil, fmt.Errorf("%w: float32 needs 4 bytes, got %d", ErrPropertyType, len(data))
		}
		bits := binary.LittleEndian.Uint32(data[:4])
		return math.Float32frombits(bits), nil

	case TypeFloat64, TypeAppTime:
		if len(data) < 8 {
			return nil, fmt.Errorf("%w: float64 needs 8 bytes, got %d", ErrPropertyType, len(data))
		}
		bits := binary.LittleEndian.Uint64(data[:8])
		return math.Float64frombits(bits), nil

	case TypeCurrency, TypeInt64:
		if len(data) < 8 {
			return nil, fmt.Errorf("%w: int64 needs 8 bytes, got %d", ErrPropertyType, len(data))
		}
		return int64(binary.LittleEndian.Uint64(data[:8])), nil

	case TypeBoolean:
		if len(data) < 4 {
			return nil, fmt.Errorf("%w: boolean needs 4 bytes, got %d", ErrPropertyType, len(data))
		}
		return binary.LittleEndian.Uint32(data[:4]) != 0, nil

	case TypeSysTime:
		if len(data) < 8 {
			return nil, fmt.Errorf("%w: systime needs 8 bytes, got %d", ErrPropertyType, len(data))
		}
		ft := binary.LittleEndian.Uint64(data[:8])
		return filetimeToTime(ft), nil

	case TypeString8:
		// Trim trailing null byte if present.
		b := trimNull8(data)
		s, err := decodeCodepage(b, codepage)
		if err != nil {
			return string(b), nil // fallback to raw bytes as string
		}
		return s, nil

	case TypeUnicode:
		return decodeUTF16LE(trimNull16(data)), nil

	case TypeBinary, TypeObject:
		cp := make([]byte, len(data))
		copy(cp, data)
		return cp, nil

	case TypeGUID:
		if len(data) < 16 {
			return nil, fmt.Errorf("%w: GUID needs 16 bytes, got %d", ErrPropertyType, len(data))
		}
		cp := make([]byte, 16)
		copy(cp, data[:16])
		return cp, nil

	default:
		// Unknown type: store raw bytes.
		cp := make([]byte, len(data))
		copy(cp, data)
		return cp, nil
	}
}

// trimNull8 trims trailing zero bytes from an ANSI string.
func trimNull8(b []byte) []byte {
	for len(b) > 0 && b[len(b)-1] == 0 {
		b = b[:len(b)-1]
	}
	return b
}

// trimNull16 trims trailing UTF-16LE null characters (0x0000) from data.
func trimNull16(b []byte) []byte {
	for len(b) >= 2 && b[len(b)-2] == 0 && b[len(b)-1] == 0 {
		b = b[:len(b)-2]
	}
	return b
}
