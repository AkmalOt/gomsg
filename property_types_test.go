package gomsg

import (
	"encoding/binary"
	"math"
	"testing"
	"time"
)

func TestIsFixedLength(t *testing.T) {
	fixed := []PropertyType{
		TypeInt16, TypeInt32, TypeFloat32, TypeFloat64,
		TypeCurrency, TypeAppTime, TypeBoolean, TypeInt64, TypeSysTime,
	}
	for _, pt := range fixed {
		if !isFixedLength(pt) {
			t.Errorf("expected %04X to be fixed length", uint16(pt))
		}
	}

	variable := []PropertyType{
		TypeString8, TypeUnicode, TypeBinary, TypeObject, TypeGUID,
		TypeMultiInt32, TypeMultiUnicode, TypeMultiBinary,
	}
	for _, pt := range variable {
		if isFixedLength(pt) {
			t.Errorf("expected %04X to be variable length", uint16(pt))
		}
	}
}

func TestConvertValueInt16(t *testing.T) {
	data := make([]byte, 2)
	binary.LittleEndian.PutUint16(data, 42)
	v, err := convertValue(TypeInt16, data, 0)
	if err != nil {
		t.Fatal(err)
	}
	if v.(int16) != 42 {
		t.Fatalf("expected 42, got %v", v)
	}
}

func TestConvertValueInt32(t *testing.T) {
	data := make([]byte, 4)
	binary.LittleEndian.PutUint32(data, 123456)
	v, err := convertValue(TypeInt32, data, 0)
	if err != nil {
		t.Fatal(err)
	}
	if v.(int32) != 123456 {
		t.Fatalf("expected 123456, got %v", v)
	}
}

func TestConvertValueBoolean(t *testing.T) {
	data := make([]byte, 4)
	binary.LittleEndian.PutUint32(data, 1)
	v, err := convertValue(TypeBoolean, data, 0)
	if err != nil {
		t.Fatal(err)
	}
	if v.(bool) != true {
		t.Fatal("expected true")
	}

	binary.LittleEndian.PutUint32(data, 0)
	v, err = convertValue(TypeBoolean, data, 0)
	if err != nil {
		t.Fatal(err)
	}
	if v.(bool) != false {
		t.Fatal("expected false")
	}
}

func TestConvertValueFloat64(t *testing.T) {
	data := make([]byte, 8)
	binary.LittleEndian.PutUint64(data, math.Float64bits(3.14))
	v, err := convertValue(TypeFloat64, data, 0)
	if err != nil {
		t.Fatal(err)
	}
	if v.(float64) != 3.14 {
		t.Fatalf("expected 3.14, got %v", v)
	}
}

func TestConvertValueSysTime(t *testing.T) {
	// 2023-01-15 12:00:00 UTC
	expected := time.Date(2023, 1, 15, 12, 0, 0, 0, time.UTC)
	// Convert to FILETIME.
	const epochDiff = 116444736000000000
	ft := uint64(expected.UnixNano()/100) + epochDiff

	data := make([]byte, 8)
	binary.LittleEndian.PutUint64(data, ft)

	v, err := convertValue(TypeSysTime, data, 0)
	if err != nil {
		t.Fatal(err)
	}

	got := v.(time.Time)
	if !got.Equal(expected) {
		t.Fatalf("expected %v, got %v", expected, got)
	}
}

func TestConvertValueUnicode(t *testing.T) {
	// "Hello" in UTF-16LE
	data := []byte{'H', 0, 'e', 0, 'l', 0, 'l', 0, 'o', 0}
	v, err := convertValue(TypeUnicode, data, 0)
	if err != nil {
		t.Fatal(err)
	}
	if v.(string) != "Hello" {
		t.Fatalf("expected Hello, got %v", v)
	}
}

func TestConvertValueUnicodeWithNull(t *testing.T) {
	// "Hi" + null terminator in UTF-16LE
	data := []byte{'H', 0, 'i', 0, 0, 0}
	v, err := convertValue(TypeUnicode, data, 0)
	if err != nil {
		t.Fatal(err)
	}
	if v.(string) != "Hi" {
		t.Fatalf("expected Hi, got %q", v)
	}
}

func TestConvertValueString8(t *testing.T) {
	data := []byte("Hello World")
	v, err := convertValue(TypeString8, data, 1252)
	if err != nil {
		t.Fatal(err)
	}
	if v.(string) != "Hello World" {
		t.Fatalf("expected Hello World, got %v", v)
	}
}

func TestConvertValueBinary(t *testing.T) {
	data := []byte{0xDE, 0xAD, 0xBE, 0xEF}
	v, err := convertValue(TypeBinary, data, 0)
	if err != nil {
		t.Fatal(err)
	}
	b := v.([]byte)
	if len(b) != 4 || b[0] != 0xDE || b[3] != 0xEF {
		t.Fatalf("unexpected binary: %x", b)
	}
}

func TestTrimNull8(t *testing.T) {
	if got := trimNull8([]byte("abc\x00\x00")); string(got) != "abc" {
		t.Fatalf("expected abc, got %q", got)
	}
	if got := trimNull8([]byte("abc")); string(got) != "abc" {
		t.Fatalf("expected abc, got %q", got)
	}
}

func TestTrimNull16(t *testing.T) {
	data := []byte{'A', 0, 0, 0}
	got := trimNull16(data)
	if len(got) != 2 || got[0] != 'A' || got[1] != 0 {
		t.Fatalf("unexpected: %x", got)
	}
}

func TestFiletimeToTime(t *testing.T) {
	// Zero filetime should return zero time.
	if !filetimeToTime(0).IsZero() {
		t.Fatal("expected zero time for filetime 0")
	}

	// Known conversion: 2020-01-01 00:00:00 UTC
	expected := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	const epochDiff = 116444736000000000
	ft := uint64(expected.UnixNano()/100) + epochDiff
	got := filetimeToTime(ft)
	if !got.Equal(expected) {
		t.Fatalf("expected %v, got %v", expected, got)
	}
}
