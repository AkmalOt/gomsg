package gomsg

import (
	"encoding/binary"
	"testing"
)

func TestParsePropertyStoreEmpty(t *testing.T) {
	// 8-byte header for sub-item context, no entries.
	data := make([]byte, 8)
	ps, err := parsePropertyStore(data, contextSubItem, noStreams, 1252)
	if err != nil {
		t.Fatal(err)
	}
	if len(ps.All()) != 0 {
		t.Fatalf("expected 0 properties, got %d", len(ps.All()))
	}
}

func TestParsePropertyStoreFixedInt32(t *testing.T) {
	// 8-byte header + one 16-byte entry.
	data := make([]byte, 24)

	// Entry: type=0x0003 (Int32), id=0x0017 (Importance), flags=0, value=2 (High)
	binary.LittleEndian.PutUint16(data[8:10], uint16(TypeInt32))
	binary.LittleEndian.PutUint16(data[10:12], uint16(PidTagImportance))
	binary.LittleEndian.PutUint32(data[12:16], 0) // flags
	binary.LittleEndian.PutUint32(data[16:20], 2) // value = ImportanceHigh

	ps, err := parsePropertyStore(data, contextSubItem, noStreams, 1252)
	if err != nil {
		t.Fatal(err)
	}

	v, ok := ps.GetInt32(PidTagImportance)
	if !ok {
		t.Fatal("PidTagImportance not found")
	}
	if v != 2 {
		t.Fatalf("expected 2, got %d", v)
	}
}

func TestParsePropertyStoreFixedBool(t *testing.T) {
	data := make([]byte, 24)

	binary.LittleEndian.PutUint16(data[8:10], uint16(TypeBoolean))
	binary.LittleEndian.PutUint16(data[10:12], uint16(PidTagHasAttachments))
	binary.LittleEndian.PutUint32(data[16:20], 1)

	ps, err := parsePropertyStore(data, contextSubItem, noStreams, 1252)
	if err != nil {
		t.Fatal(err)
	}

	v, ok := ps.GetBool(PidTagHasAttachments)
	if !ok {
		t.Fatal("PidTagHasAttachments not found")
	}
	if !v {
		t.Fatal("expected true")
	}
}

func TestParsePropertyStoreVariableString(t *testing.T) {
	// Header (8) + one entry (16) = 24 bytes.
	data := make([]byte, 24)

	// Unicode string property for Subject.
	binary.LittleEndian.PutUint16(data[8:10], uint16(TypeUnicode))
	binary.LittleEndian.PutUint16(data[10:12], uint16(PidTagSubject))
	binary.LittleEndian.PutUint32(data[12:16], 0) // flags
	// For variable-length, bytes 16-19 = size (not used by our parser, it reads from stream).
	binary.LittleEndian.PutUint32(data[16:20], 10)

	// The stream name should be __substg1.0_0037001F (subject, unicode).
	subjectUTF16 := []byte{'T', 0, 'e', 0, 's', 0, 't', 0}

	lookup := func(name string) ([]byte, bool) {
		if name == "__SUBSTG1.0_0037001F" {
			return subjectUTF16, true
		}
		return nil, false
	}

	ps, err := parsePropertyStore(data, contextSubItem, lookup, 1252)
	if err != nil {
		t.Fatal(err)
	}

	s := ps.GetString(PidTagSubject)
	if s != "Test" {
		t.Fatalf("expected 'Test', got %q", s)
	}
}

func TestParsePropertyStoreMessageHeader(t *testing.T) {
	// 32-byte header for root message context + one entry.
	data := make([]byte, 48)

	// Header: 8 reserved + nextRecipientID + nextAttachmentID + recipientCount + attachmentCount + 8 reserved
	binary.LittleEndian.PutUint32(data[8:12], 3)  // nextRecipientID
	binary.LittleEndian.PutUint32(data[12:16], 1)  // nextAttachmentID
	binary.LittleEndian.PutUint32(data[16:20], 3)  // recipientCount
	binary.LittleEndian.PutUint32(data[20:24], 1)  // attachmentCount

	// Entry at offset 32: Int32 importance.
	binary.LittleEndian.PutUint16(data[32:34], uint16(TypeInt32))
	binary.LittleEndian.PutUint16(data[34:36], uint16(PidTagImportance))
	binary.LittleEndian.PutUint32(data[40:44], 1) // Normal

	ps, err := parsePropertyStore(data, contextMessage, noStreams, 1252)
	if err != nil {
		t.Fatal(err)
	}

	if ps.RecipientCount != 3 {
		t.Fatalf("expected 3 recipients, got %d", ps.RecipientCount)
	}
	if ps.AttachmentCount != 1 {
		t.Fatalf("expected 1 attachment, got %d", ps.AttachmentCount)
	}

	v, ok := ps.GetInt32(PidTagImportance)
	if !ok || v != 1 {
		t.Fatalf("expected importance 1, got %d (ok=%v)", v, ok)
	}
}

func noStreams(name string) ([]byte, bool) {
	return nil, false
}
