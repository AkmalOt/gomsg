package gomsg

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/richardlehane/mscfb"
)

// Importance represents the message priority level.
type Importance int32

const (
	ImportanceLow    Importance = 0
	ImportanceNormal Importance = 1
	ImportanceHigh   Importance = 2
)

// Message represents a parsed Outlook .msg file.
type Message struct {
	Subject     string
	Body        string
	BodyHTML    []byte
	BodyRTF    []byte

	SenderName  string
	SenderEmail string
	SenderSMTP  string
	SenderType  string // address type: "SMTP", "EX", etc.

	DisplayTo  string
	DisplayCC  string
	DisplayBCC string

	Recipients  []Recipient
	Attachments []Attachment

	Date         time.Time
	DeliveryTime time.Time
	MessageClass string
	Importance   Importance
	MessageID    string
	InReplyTo    string
	Headers      string

	ConversationTopic string

	Properties *PropertyStore
}

// Open parses an MSG file from the given path.
func Open(path string) (*Message, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return nil, err
	}

	return OpenReader(f, info.Size())
}

// Decode parses an MSG file from a reader by reading all data into memory.
func Decode(r io.Reader) (*Message, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	ra := bytes.NewReader(data)
	return OpenReader(ra, int64(len(data)))
}

// OpenReader parses an MSG file from an io.ReaderAt.
func OpenReader(r io.ReaderAt, size int64) (*Message, error) {
	return parseMsg(r, size, contextMessage)
}

// storageNode represents a collected CFB storage with its child streams.
type storageNode struct {
	path    string
	streams map[string][]byte // stream name -> data
	children map[string]*storageNode
}

// regex patterns for identifying sub-storages.
var (
	attachPattern = regexp.MustCompile(`(?i)^__attach_version1\.0_#([0-9A-Fa-f]{8})$`)
	recipPattern  = regexp.MustCompile(`(?i)^__recip_version1\.0_#([0-9A-Fa-f]{8})$`)
)

const propertiesStream = "__properties_version1.0"

// parseMsg reads a CFB container and assembles a Message.
func parseMsg(r io.ReaderAt, size int64, ctx storageContext) (*Message, error) {
	sr := io.NewSectionReader(r, 0, size)
	doc, err := mscfb.New(sr)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidCFB, err)
	}

	// Collect all streams organized by storage path.
	rootStreams := make(map[string][]byte)
	attachStorages := make(map[string]map[string][]byte)  // attach index -> streams
	recipStorages := make(map[string]map[string][]byte)    // recip index -> streams
	embeddedStorages := make(map[string]map[string][]byte) // attach + embedded sub-streams

	for entry, err := doc.Next(); err == nil; entry, err = doc.Next() {
		name := entry.Name
		path := entry.Path

		if entry.FileInfo().IsDir() {
			continue
		}

		data, readErr := io.ReadAll(entry)
		if readErr != nil {
			continue
		}

		pathLen := len(path)

		if pathLen == 0 {
			// Root-level stream.
			rootStreams[name] = data
		} else if pathLen == 1 {
			parentName := path[0]

			if attachPattern.MatchString(parentName) {
				if attachStorages[parentName] == nil {
					attachStorages[parentName] = make(map[string][]byte)
				}
				attachStorages[parentName][name] = data
			} else if recipPattern.MatchString(parentName) {
				if recipStorages[parentName] == nil {
					recipStorages[parentName] = make(map[string][]byte)
				}
				recipStorages[parentName][name] = data
			} else {
				// Other root-level sub-storage streams (e.g., __nameid_version1.0).
				key := parentName + "/" + name
				rootStreams[key] = data
			}
		} else if pathLen >= 2 {
			// Deeper nesting — could be embedded MSG within an attachment.
			parentAttach := path[0]
			if attachPattern.MatchString(parentAttach) {
				// Build a key that preserves the sub-path.
				subPath := strings.Join(path[1:], "/") + "/" + name
				key := parentAttach
				if embeddedStorages[key] == nil {
					embeddedStorages[key] = make(map[string][]byte)
				}
				embeddedStorages[key][subPath] = data
			}
		}
	}

	// Parse root properties.
	propsData, ok := rootStreams[propertiesStream]
	if !ok {
		return nil, ErrNoProperties
	}

	rootLookup := func(name string) ([]byte, bool) {
		d, ok := rootStreams[strings.ToUpper(name)]
		if ok {
			return d, true
		}
		// Try case-insensitive match.
		for k, v := range rootStreams {
			if strings.EqualFold(k, name) {
				return v, true
			}
		}
		return nil, false
	}

	rootProps, err := parsePropertyStore(propsData, ctx, rootLookup, 0)
	if err != nil {
		return nil, err
	}

	// Detect codepage.
	codepage := detectCodepage(rootProps)

	// Re-parse with correct codepage if not default.
	if codepage != 0 && codepage != 1252 {
		rootProps, _ = parsePropertyStore(propsData, ctx, rootLookup, codepage)
	}

	// Parse recipients.
	recipients := parseRecipients(recipStorages, codepage)

	// Parse attachments.
	attachments := parseAttachments(attachStorages, embeddedStorages, codepage)

	return assembleMessage(rootProps, recipients, attachments), nil
}

// detectCodepage extracts the codepage from message properties.
func detectCodepage(ps *PropertyStore) uint32 {
	if cp, ok := ps.GetInt32(PidTagInternetCodepage); ok && cp > 0 {
		return uint32(cp)
	}
	if cp, ok := ps.GetInt32(PidTagMessageCodepage); ok && cp > 0 {
		return uint32(cp)
	}
	return 1252 // Windows-1252 default
}

// parseRecipients processes all __recip_version1.0_#* sub-storages.
func parseRecipients(storages map[string]map[string][]byte, codepage uint32) []Recipient {
	// Sort keys for deterministic order.
	keys := make([]string, 0, len(storages))
	for k := range storages {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var recipients []Recipient
	for _, key := range keys {
		streams := storages[key]
		propsData, ok := findStream(streams, propertiesStream)
		if !ok {
			continue
		}

		lookup := makeLookup(streams)
		ps, err := parsePropertyStore(propsData, contextSubItem, lookup, codepage)
		if err != nil {
			continue
		}

		recipients = append(recipients, parseRecipient(ps))
	}
	return recipients
}

// parseAttachments processes all __attach_version1.0_#* sub-storages.
func parseAttachments(
	storages map[string]map[string][]byte,
	embeddedStorages map[string]map[string][]byte,
	codepage uint32,
) []Attachment {
	keys := make([]string, 0, len(storages))
	for k := range storages {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var attachments []Attachment
	for _, key := range keys {
		streams := storages[key]
		propsData, ok := findStream(streams, propertiesStream)
		if !ok {
			continue
		}

		lookup := makeLookup(streams)
		ps, err := parsePropertyStore(propsData, contextSubItem, lookup, codepage)
		if err != nil {
			continue
		}

		var attachData []byte
		var embeddedMsg *Message

		method, _ := ps.GetInt32(PidTagAttachMethod)

		switch AttachMethod(method) {
		case AttachEmbeddedMessage:
			// Try to parse embedded MSG from deeper sub-streams.
			if embStreams, ok := embeddedStorages[key]; ok && len(embStreams) > 0 {
				embeddedMsg = parseEmbeddedMsg(embStreams, codepage)
			}
		default:
			// Binary attachment data.
			attachData = ps.GetBytes(PidTagAttachDataBinary)
			if attachData == nil {
				// Try finding the stream directly.
				attachData, _ = findStream(streams, "__substg1.0_37010102")
			}
		}

		attachments = append(attachments, parseAttachment(ps, attachData, embeddedMsg))
	}
	return attachments
}

// parseEmbeddedMsg attempts to parse an embedded MSG from sub-streams
// that were nested within an attachment storage.
func parseEmbeddedMsg(streams map[string][]byte, codepage uint32) *Message {
	// The embedded MSG streams have paths relative to the __substg1.0_3701000D storage.
	// We need to find the properties stream and substg streams within it.

	// Flatten: find streams that belong to the embedded message root.
	rootStreams := make(map[string][]byte)
	attachStorages := make(map[string]map[string][]byte)
	recipStorages := make(map[string]map[string][]byte)
	embeddedSub := make(map[string]map[string][]byte)

	embedPrefix := "__substg1.0_3701000D/"

	for path, data := range streams {
		// Remove the embed prefix if present.
		rel := path
		if strings.HasPrefix(strings.ToUpper(path), strings.ToUpper(embedPrefix)) {
			rel = path[len(embedPrefix):]
		}

		// Check if this is inside a sub-storage of the embedded msg.
		parts := strings.SplitN(rel, "/", 2)
		if len(parts) == 1 {
			// Direct child of embedded root.
			rootStreams[parts[0]] = data
		} else {
			parentName := parts[0]
			childName := parts[1]

			if attachPattern.MatchString(parentName) {
				if strings.Contains(childName, "/") {
					if embeddedSub[parentName] == nil {
						embeddedSub[parentName] = make(map[string][]byte)
					}
					embeddedSub[parentName][childName] = data
				} else {
					if attachStorages[parentName] == nil {
						attachStorages[parentName] = make(map[string][]byte)
					}
					attachStorages[parentName][childName] = data
				}
			} else if recipPattern.MatchString(parentName) {
				if recipStorages[parentName] == nil {
					recipStorages[parentName] = make(map[string][]byte)
				}
				recipStorages[parentName][childName] = data
			} else {
				key := parentName + "/" + childName
				rootStreams[key] = data
			}
		}
	}

	propsData, ok := findStream(rootStreams, propertiesStream)
	if !ok {
		return nil
	}

	lookup := makeLookup(rootStreams)
	ps, err := parsePropertyStore(propsData, contextEmbed, lookup, codepage)
	if err != nil {
		return nil
	}

	cp := detectCodepage(ps)
	if cp != codepage && cp != 1252 {
		ps, _ = parsePropertyStore(propsData, contextEmbed, lookup, cp)
	}

	recipients := parseRecipients(recipStorages, cp)
	attachments := parseAttachments(attachStorages, embeddedSub, cp)

	return assembleMessage(ps, recipients, attachments)
}

// assembleMessage builds a Message from parsed properties, recipients, and attachments.
func assembleMessage(ps *PropertyStore, recipients []Recipient, attachments []Attachment) *Message {
	msg := &Message{
		Subject:           ps.GetString(PidTagSubject),
		Body:              ps.GetString(PidTagBody),
		BodyHTML:          ps.GetBytes(PidTagBodyHTML),
		BodyRTF:           ps.GetBytes(PidTagRTFCompressed),
		SenderName:        ps.GetString(PidTagSenderName),
		SenderEmail:       ps.GetString(PidTagSenderEmailAddress),
		SenderSMTP:        ps.GetString(PidTagSenderSMTPAddress),
		SenderType:        ps.GetString(PidTagSenderAddrType),
		DisplayTo:         ps.GetString(PidTagDisplayTo),
		DisplayCC:         ps.GetString(PidTagDisplayCC),
		DisplayBCC:        ps.GetString(PidTagDisplayBCC),
		MessageClass:      ps.GetString(PidTagMessageClass),
		MessageID:         ps.GetString(PidTagInternetMessageID),
		InReplyTo:         ps.GetString(PidTagInReplyToID),
		Headers:           ps.GetString(PidTagTransportMessageHeaders),
		ConversationTopic: ps.GetString(PidTagConversationTopic),
		Recipients:        recipients,
		Attachments:       attachments,
		Properties:        ps,
	}

	if t, ok := ps.GetTime(PidTagClientSubmitTime); ok {
		msg.Date = t
	}
	if t, ok := ps.GetTime(PidTagMessageDeliveryTime); ok {
		msg.DeliveryTime = t
	}

	if imp, ok := ps.GetInt32(PidTagImportance); ok {
		msg.Importance = Importance(imp)
	} else {
		msg.Importance = ImportanceNormal
	}

	// If SMTP address is empty, try sender email if it looks like SMTP.
	if msg.SenderSMTP == "" && msg.SenderEmail != "" {
		if strings.Contains(msg.SenderEmail, "@") {
			msg.SenderSMTP = msg.SenderEmail
		}
	}

	return msg
}

// findStream looks up a stream by name (case-insensitive).
func findStream(streams map[string][]byte, name string) ([]byte, bool) {
	if d, ok := streams[name]; ok {
		return d, true
	}
	for k, v := range streams {
		if strings.EqualFold(k, name) {
			return v, true
		}
	}
	return nil, false
}

// makeLookup creates a case-insensitive stream lookup function.
func makeLookup(streams map[string][]byte) func(string) ([]byte, bool) {
	return func(name string) ([]byte, bool) {
		return findStream(streams, name)
	}
}
