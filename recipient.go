package gomsg

// RecipientType indicates whether a recipient is To, CC, or BCC.
type RecipientType int32

const (
	RecipientOriginator RecipientType = 0
	RecipientTo         RecipientType = 1
	RecipientCC         RecipientType = 2
	RecipientBCC        RecipientType = 3
)

// String returns a human-readable label for the recipient type.
func (rt RecipientType) String() string {
	switch rt {
	case RecipientTo:
		return "To"
	case RecipientCC:
		return "CC"
	case RecipientBCC:
		return "BCC"
	default:
		return "Unknown"
	}
}

// Recipient represents an email recipient.
type Recipient struct {
	DisplayName  string
	EmailAddress string
	SMTPAddress  string
	Type         RecipientType
	Properties   *PropertyStore
}

// parseRecipient builds a Recipient from a PropertyStore parsed from
// a __recip_version1.0_#XXXXXXXX sub-storage.
func parseRecipient(ps *PropertyStore) Recipient {
	r := Recipient{
		DisplayName:  ps.GetString(PidTagDisplayName),
		EmailAddress: ps.GetString(PidTagEmailAddress),
		SMTPAddress:  ps.GetString(PidTagSMTPAddress),
		Properties:   ps,
	}

	if rt, ok := ps.GetInt32(PidTagRecipientType); ok {
		r.Type = RecipientType(rt)
	}

	// Prefer SMTP address over Exchange-style address.
	if r.SMTPAddress == "" && r.EmailAddress != "" {
		r.SMTPAddress = r.EmailAddress
	}

	return r
}
