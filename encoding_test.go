package gomsg

import "testing"

func TestDecodeUTF16LE(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected string
	}{
		{"empty", nil, ""},
		{"single byte", []byte{0x41}, ""},
		{"hello", []byte{'H', 0, 'e', 0, 'l', 0, 'l', 0, 'o', 0}, "Hello"},
		{"cyrillic", []byte{0x1F, 0x04, 0x40, 0x04, 0x38, 0x04, 0x32, 0x04, 0x35, 0x04, 0x42, 0x04}, "Привет"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := decodeUTF16LE(tt.input)
			if got != tt.expected {
				t.Fatalf("expected %q, got %q", tt.expected, got)
			}
		})
	}
}

func TestDecodeCodepage1252(t *testing.T) {
	// Basic ASCII should pass through.
	got, err := decodeCodepage([]byte("Hello"), 1252)
	if err != nil {
		t.Fatal(err)
	}
	if got != "Hello" {
		t.Fatalf("expected Hello, got %q", got)
	}
}

func TestDecodeCodepageKOI8R(t *testing.T) {
	// KOI8-R encoded "Привет"
	koi8r := []byte{0xF0, 0xD2, 0xC9, 0xD7, 0xC5, 0xD4}
	got, err := decodeCodepage(koi8r, 20866)
	if err != nil {
		t.Fatal(err)
	}
	if got != "Привет" {
		t.Fatalf("expected Привет, got %q", got)
	}
}

func TestDecodeCodepageUnknown(t *testing.T) {
	// Unknown codepage should return raw bytes as string.
	data := []byte("raw")
	got, err := decodeCodepage(data, 99999)
	if err != nil {
		t.Fatal(err)
	}
	if got != "raw" {
		t.Fatalf("expected raw, got %q", got)
	}
}

func TestCodepageEncoding(t *testing.T) {
	known := []uint32{0, 1250, 1251, 1252, 1253, 1254, 1255, 1256, 1257, 1258,
		874, 20866, 21866, 65001, 932, 949, 936, 950}
	for _, cp := range known {
		if codepageEncoding(cp) == nil {
			t.Errorf("codepage %d should return a non-nil encoding", cp)
		}
	}

	if codepageEncoding(12345) != nil {
		t.Error("unknown codepage should return nil")
	}
}
