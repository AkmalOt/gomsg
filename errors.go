package gomsg

import "errors"

var (
	// ErrNotMSG indicates the file is not a valid MSG file.
	ErrNotMSG = errors.New("gomsg: not a valid MSG file")

	// ErrInvalidCFB indicates the CFB container is malformed.
	ErrInvalidCFB = errors.New("gomsg: invalid CFB container")

	// ErrNoProperties indicates the __properties_version1.0 stream is missing.
	ErrNoProperties = errors.New("gomsg: missing properties stream")

	// ErrPropertyType indicates an unexpected property type was encountered.
	ErrPropertyType = errors.New("gomsg: unexpected property type")
)
