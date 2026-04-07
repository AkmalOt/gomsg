package gomsg

// Named properties (0x8000+ range) support.
// In v1, we provide basic structure; full named property resolution
// will be added in a future version.

// NamedPropertyMapping holds the mapping from named property IDs
// to their MAPI property set GUID and name/ID.
type NamedPropertyMapping struct {
	entries []namedEntry
}

type namedEntry struct {
	PropertyID PropertyID // The mapped ID (0x8000+)
	GUID       [16]byte   // Property set GUID
	Kind       uint32     // 0 = numeric ID, 1 = string name
	NumericID  uint32     // If Kind == 0
	StringName string     // If Kind == 1
}

// Well-known property set GUIDs.
var (
	PSPublicStrings = [16]byte{
		0x29, 0x03, 0x02, 0x00, 0x00, 0x00, 0x00, 0x00,
		0xC0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46,
	}
	PSInternetHeaders = [16]byte{
		0x20, 0x38, 0x60, 0x00, 0xFA, 0xB0, 0x9C, 0x01,
		0x10, 0x49, 0x00, 0x00, 0x0B, 0x6B, 0x3A, 0x03,
	}
)

// parseNamedProperties attempts to parse the __nameid_version1.0 sub-storage.
// This is a placeholder for future implementation.
func parseNamedProperties(rootStreams map[string][]byte) *NamedPropertyMapping {
	// The named properties mapping is stored in three streams:
	// __nameid_version1.0/__substg1.0_00020102 - GUID stream
	// __nameid_version1.0/__substg1.0_00030102 - entry stream
	// __nameid_version1.0/__substg1.0_00040102 - string stream
	//
	// TODO: Full implementation in v1.1
	return nil
}
