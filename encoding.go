package gomsg

import (
	"encoding/binary"
	"unicode/utf16"

	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/encoding/korean"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/encoding/traditionalchinese"
	"golang.org/x/text/encoding/unicode"
)

// decodeUTF16LE decodes a UTF-16LE byte slice to a Go string.
func decodeUTF16LE(b []byte) string {
	if len(b) < 2 {
		return ""
	}
	u16s := make([]uint16, len(b)/2)
	for i := range u16s {
		u16s[i] = binary.LittleEndian.Uint16(b[i*2 : i*2+2])
	}
	return string(utf16.Decode(u16s))
}

// decodeCodepage decodes bytes using the specified Windows codepage.
// If the codepage is unknown, returns the raw bytes as a string.
func decodeCodepage(b []byte, codepage uint32) (string, error) {
	enc := codepageEncoding(codepage)
	if enc == nil {
		return string(b), nil
	}
	decoded, err := enc.NewDecoder().Bytes(b)
	if err != nil {
		return string(b), err
	}
	return string(decoded), nil
}

// codepageEncoding maps a Windows codepage number to a Go text encoding.
func codepageEncoding(cp uint32) encoding.Encoding {
	switch cp {
	case 0, 1252, 28591:
		return charmap.Windows1252
	case 1250:
		return charmap.Windows1250
	case 1251:
		return charmap.Windows1251
	case 1253:
		return charmap.Windows1253
	case 1254:
		return charmap.Windows1254
	case 1255:
		return charmap.Windows1255
	case 1256:
		return charmap.Windows1256
	case 1257:
		return charmap.Windows1257
	case 1258:
		return charmap.Windows1258
	case 874:
		return charmap.Windows874
	case 28592:
		return charmap.ISO8859_2
	case 28593:
		return charmap.ISO8859_3
	case 28594:
		return charmap.ISO8859_4
	case 28595:
		return charmap.ISO8859_5
	case 28596:
		return charmap.ISO8859_6
	case 28597:
		return charmap.ISO8859_7
	case 28598:
		return charmap.ISO8859_8
	case 28599:
		return charmap.ISO8859_9
	case 28603:
		return charmap.ISO8859_13
	case 28605:
		return charmap.ISO8859_15
	case 20866:
		return charmap.KOI8R
	case 21866:
		return charmap.KOI8U
	case 65001:
		return unicode.UTF8
	case 65000:
		return unicode.UTF8 // UTF-7 not supported, fallback to UTF-8
	case 932:
		return japanese.ShiftJIS
	case 50220, 50221, 50222:
		return japanese.ISO2022JP
	case 51932:
		return japanese.EUCJP
	case 949:
		return korean.EUCKR
	case 936:
		return simplifiedchinese.GBK
	case 54936:
		return simplifiedchinese.GB18030
	case 52936:
		return simplifiedchinese.HZGB2312
	case 950:
		return traditionalchinese.Big5
	default:
		return nil
	}
}
