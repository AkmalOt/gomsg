package gomsg

// AttachMethod specifies how attachment data is stored.
type AttachMethod int32

const (
	AttachByValue         AttachMethod = 1 // Binary data in PidTagAttachDataBinary
	AttachByReference     AttachMethod = 2 // Path reference
	AttachByRefOnly       AttachMethod = 4 // Path reference only
	AttachEmbeddedMessage AttachMethod = 5 // Embedded MSG in sub-storage
	AttachOLE             AttachMethod = 6 // OLE object
)

// Attachment represents a file or embedded message attached to the email.
type Attachment struct {
	FileName    string
	LongName    string
	Extension   string
	MIMEType    string
	ContentID   string
	Size        int64
	Method      AttachMethod
	data        []byte
	embeddedMsg *Message
	Properties  *PropertyStore
}

// Data returns the attachment's binary content.
func (a *Attachment) Data() []byte {
	return a.data
}

// IsEmbeddedMessage returns true if this attachment is an embedded .msg file.
func (a *Attachment) IsEmbeddedMessage() bool {
	return a.Method == AttachEmbeddedMessage && a.embeddedMsg != nil
}

// EmbeddedMessage returns the parsed embedded Message, or nil if not embedded.
func (a *Attachment) EmbeddedMessage() *Message {
	return a.embeddedMsg
}

// DisplayName returns the best available filename for the attachment.
func (a *Attachment) DisplayName() string {
	if a.LongName != "" {
		return a.LongName
	}
	if a.FileName != "" {
		return a.FileName
	}
	return "untitled"
}

// parseAttachment builds an Attachment from a PropertyStore.
// embeddedMsg is non-nil when the attachment is an embedded MSG (method 5)
// and was recursively parsed.
func parseAttachment(ps *PropertyStore, data []byte, embeddedMsg *Message) Attachment {
	a := Attachment{
		FileName:   ps.GetString(PidTagAttachFilename),
		LongName:   ps.GetString(PidTagAttachLongFilename),
		Extension:  ps.GetString(PidTagAttachExtension),
		MIMEType:   ps.GetString(PidTagAttachMIMETag),
		ContentID:  ps.GetString(PidTagAttachContentID),
		data:       data,
		Properties: ps,
	}

	if m, ok := ps.GetInt32(PidTagAttachMethod); ok {
		a.Method = AttachMethod(m)
	}

	if a.data != nil {
		a.Size = int64(len(a.data))
	}

	if embeddedMsg != nil {
		a.Method = AttachEmbeddedMessage
		a.embeddedMsg = embeddedMsg
	}

	return a
}
