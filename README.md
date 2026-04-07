# gomsg

Go library for parsing Microsoft Outlook `.msg` files. Extracts subject, sender, recipients, body (text/HTML/RTF), attachments, dates, and other MAPI properties.

`.msg` files use the OLE2 (CFB) binary format. The library parses the container and reads MAPI property streams inside it.

[Документация на русском](README.ru.md)

## Installation

```bash
go get github.com/AkmalOt/gomsg
```

## Usage

### As a library

```go
package main

import (
    "fmt"
    "log"
    "os"

    "github.com/AkmalOt/gomsg"
)

func main() {
    msg, err := gomsg.Open("email.msg")
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println("Subject:", msg.Subject)
    fmt.Println("From:", msg.SenderName, "<"+msg.SenderSMTP+">")
    fmt.Println("To:", msg.DisplayTo)
    fmt.Println("Date:", msg.Date)
    fmt.Println("Body:", msg.Body)

    // Extract attachments
    for _, a := range msg.Attachments {
        fmt.Printf("Attachment: %s (%d bytes)\n", a.DisplayName(), a.Size)
        os.WriteFile(a.DisplayName(), a.Data(), 0644)
    }
}
```

### CLI tool

The library comes with a `msgdump` utility:

```bash
go install github.com/AkmalOt/gomsg/cmd/msgdump@latest
```

```bash
# Print email summary
msgdump email.msg

# JSON output
msgdump -json email.msg

# Extract attachments to a directory
msgdump -extract ./attachments email.msg

# Print message body
msgdump -body email.msg

# Print transport headers
msgdump -headers email.msg
```

Example output:

```
Subject:      Test message
From:         John <john@example.com>
To:           jane@example.com
Date:         2024-01-15 12:30:00 UTC
Class:        IPM.Note
Importance:   Normal

Recipients (1):
  [To] Jane <jane@example.com>

Attachments (1):
  document.pdf (application/pdf, 45230 bytes)
```

## Available fields

| Field | Property |
|-------|----------|
| Subject | `Message.Subject` |
| Body (plain text) | `Message.Body` |
| Body (HTML) | `Message.BodyHTML` |
| Sender name | `Message.SenderName` |
| Sender email | `Message.SenderSMTP` |
| To | `Message.DisplayTo` |
| CC | `Message.DisplayCC` |
| BCC | `Message.DisplayBCC` |
| Recipient list | `Message.Recipients` |
| Attachments | `Message.Attachments` |
| Date sent | `Message.Date` |
| Delivery time | `Message.DeliveryTime` |
| Message class | `Message.MessageClass` |
| Importance | `Message.Importance` |
| Message-ID | `Message.MessageID` |
| Transport headers | `Message.Headers` |

Any MAPI property can also be accessed directly via `Message.Properties`.

## Attachments

The library extracts file attachments and supports embedded `.msg` files (email inside email):

```go
for _, a := range msg.Attachments {
    if a.IsEmbeddedMessage() {
        inner := a.EmbeddedMessage()
        fmt.Println("Embedded message:", inner.Subject)
    } else {
        os.WriteFile(a.DisplayName(), a.Data(), 0644)
    }
}
```

## Encoding support

Handles UTF-16LE (Unicode) strings and various Windows codepages: 1250-1258, KOI8-R, KOI8-U, ISO-8859, Shift-JIS, EUC-KR, GBK, Big5, and others. The encoding is detected automatically from message properties.

## Dependencies

- [richardlehane/mscfb](https://github.com/richardlehane/mscfb) — CFB/OLE2 format parser
- [golang.org/x/text](https://pkg.go.dev/golang.org/x/text) — encoding support

## License

MIT
