package gomsg

// PropertyID represents a MAPI property identifier.
type PropertyID uint16

// Standard MAPI property IDs used in MSG files.
const (
	PidTagMessageClass             PropertyID = 0x001A
	PidTagSubject                  PropertyID = 0x0037
	PidTagSubjectPrefix            PropertyID = 0x003D
	PidTagClientSubmitTime         PropertyID = 0x0039
	PidTagSentRepresentingName     PropertyID = 0x0042
	PidTagSentRepresentingEmail    PropertyID = 0x0065
	PidTagImportance               PropertyID = 0x0017
	PidTagSensitivity              PropertyID = 0x0036
	PidTagTransportMessageHeaders  PropertyID = 0x007D
	PidTagDisplayTo                PropertyID = 0x0E04
	PidTagDisplayCC                PropertyID = 0x0E03
	PidTagDisplayBCC               PropertyID = 0x0E02
	PidTagMessageFlags             PropertyID = 0x0E07
	PidTagNormalizedSubject        PropertyID = 0x0E1D
	PidTagHasAttachments           PropertyID = 0x0E1B
	PidTagBody                     PropertyID = 0x1000
	PidTagBodyHTML                 PropertyID = 0x1013
	PidTagRTFCompressed            PropertyID = 0x1009
	PidTagInternetMessageID        PropertyID = 0x1035
	PidTagInReplyToID              PropertyID = 0x1042
	PidTagSenderName               PropertyID = 0x0C1A
	PidTagSenderEmailAddress       PropertyID = 0x0C1F
	PidTagSenderAddrType           PropertyID = 0x0C1E
	PidTagSenderSMTPAddress        PropertyID = 0x5D01
	PidTagRecipientType            PropertyID = 0x0C15
	PidTagDisplayName              PropertyID = 0x3001
	PidTagEmailAddress             PropertyID = 0x3003
	PidTagAddrType                 PropertyID = 0x3002
	PidTagSMTPAddress              PropertyID = 0x39FE
	PidTagAttachDataBinary         PropertyID = 0x3701
	PidTagAttachDataObject         PropertyID = 0x3701 // same ID, type 0x000D for embedded
	PidTagAttachEncoding           PropertyID = 0x3702
	PidTagAttachFilename           PropertyID = 0x3704
	PidTagAttachMethod             PropertyID = 0x3705
	PidTagAttachLongFilename       PropertyID = 0x3707
	PidTagAttachMIMETag            PropertyID = 0x370E
	PidTagAttachExtension          PropertyID = 0x3703
	PidTagAttachSize               PropertyID = 0x0E20
	PidTagAttachContentID          PropertyID = 0x3712
	PidTagInternetCodepage         PropertyID = 0x3FDE
	PidTagMessageCodepage          PropertyID = 0x3FFD
	PidTagCreationTime             PropertyID = 0x3007
	PidTagLastModificationTime     PropertyID = 0x3008
	PidTagMessageDeliveryTime      PropertyID = 0x0E06
	PidTagConversationTopic        PropertyID = 0x0070
	PidTagConversationIndex        PropertyID = 0x0071
)
