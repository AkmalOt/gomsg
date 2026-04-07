package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/AkmalOt/gomsg"
)

func main() {
	jsonOutput := flag.Bool("json", false, "output as JSON")
	showHeaders := flag.Bool("headers", false, "print transport headers")
	showBody := flag.Bool("body", false, "print message body")
	extractDir := flag.String("extract", "", "extract attachments to directory")
	showAll := flag.Bool("all", false, "show all properties (debug)")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: msgdump [options] file.msg\n\nOptions:\n")
		flag.PrintDefaults()
	}

	flag.Parse()
	if flag.NArg() < 1 {
		flag.Usage()
		os.Exit(1)
	}

	msg, err := gomsg.Open(flag.Arg(0))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if *jsonOutput {
		printJSON(msg)
		return
	}

	if *showHeaders {
		fmt.Println(msg.Headers)
		return
	}

	if *showBody {
		fmt.Println(msg.Body)
		return
	}

	if *showAll {
		printAllProperties(msg)
		return
	}

	if *extractDir != "" {
		extractAttachments(msg, *extractDir)
		return
	}

	// Default: print summary.
	printSummary(msg)
}

func printSummary(msg *gomsg.Message) {
	fmt.Printf("Subject:      %s\n", msg.Subject)
	fmt.Printf("From:         %s <%s>\n", msg.SenderName, senderAddr(msg))
	fmt.Printf("To:           %s\n", msg.DisplayTo)
	if msg.DisplayCC != "" {
		fmt.Printf("CC:           %s\n", msg.DisplayCC)
	}
	if msg.DisplayBCC != "" {
		fmt.Printf("BCC:          %s\n", msg.DisplayBCC)
	}
	fmt.Printf("Date:         %s\n", msg.Date.Format("2006-01-02 15:04:05 MST"))
	fmt.Printf("Class:        %s\n", msg.MessageClass)
	fmt.Printf("Importance:   %s\n", importanceStr(msg.Importance))

	if msg.MessageID != "" {
		fmt.Printf("Message-ID:   %s\n", msg.MessageID)
	}

	if len(msg.Recipients) > 0 {
		fmt.Printf("\nRecipients (%d):\n", len(msg.Recipients))
		for _, r := range msg.Recipients {
			addr := r.SMTPAddress
			if addr == "" {
				addr = r.EmailAddress
			}
			fmt.Printf("  [%s] %s <%s>\n", r.Type, r.DisplayName, addr)
		}
	}

	if len(msg.Attachments) > 0 {
		fmt.Printf("\nAttachments (%d):\n", len(msg.Attachments))
		for _, a := range msg.Attachments {
			if a.IsEmbeddedMessage() {
				fmt.Printf("  [embedded MSG] %s\n", a.DisplayName())
			} else {
				fmt.Printf("  %s (%s, %d bytes)\n", a.DisplayName(), a.MIMEType, a.Size)
			}
		}
	}
}

func printJSON(msg *gomsg.Message) {
	type jsonRecipient struct {
		DisplayName string `json:"display_name"`
		Email       string `json:"email"`
		Type        string `json:"type"`
	}

	type jsonAttach struct {
		FileName string `json:"filename"`
		MIMEType string `json:"mime_type"`
		Size     int64  `json:"size"`
		Embedded bool   `json:"embedded"`
	}

	type jsonMsg struct {
		Subject     string `json:"subject"`
		Body        string `json:"body"`
		SenderName  string `json:"sender_name"`
		SenderEmail string `json:"sender_email"`
		DisplayTo   string `json:"display_to"`
		DisplayCC   string `json:"display_cc,omitempty"`
		DisplayBCC  string `json:"display_bcc,omitempty"`
		Date        string `json:"date"`
		Class       string `json:"message_class"`
		Importance  string `json:"importance"`
		MessageID   string `json:"message_id,omitempty"`
		Recipients  []jsonRecipient `json:"recipients,omitempty"`
		Attachments []jsonAttach    `json:"attachments,omitempty"`
	}

	jm := jsonMsg{
		Subject:     msg.Subject,
		Body:        msg.Body,
		SenderName:  msg.SenderName,
		SenderEmail: senderAddr(msg),
		DisplayTo:   msg.DisplayTo,
		DisplayCC:   msg.DisplayCC,
		DisplayBCC:  msg.DisplayBCC,
		Date:        msg.Date.Format("2006-01-02T15:04:05Z07:00"),
		Class:       msg.MessageClass,
		Importance:  importanceStr(msg.Importance),
		MessageID:   msg.MessageID,
	}

	for _, r := range msg.Recipients {
		addr := r.SMTPAddress
		if addr == "" {
			addr = r.EmailAddress
		}
		jm.Recipients = append(jm.Recipients, jsonRecipient{
			DisplayName: r.DisplayName,
			Email:       addr,
			Type:        r.Type.String(),
		})
	}

	for _, a := range msg.Attachments {
		jm.Attachments = append(jm.Attachments, jsonAttach{
			FileName: a.DisplayName(),
			MIMEType: a.MIMEType,
			Size:     a.Size,
			Embedded: a.IsEmbeddedMessage(),
		})
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.Encode(jm)
}

func printAllProperties(msg *gomsg.Message) {
	if msg.Properties == nil {
		fmt.Println("No properties found.")
		return
	}
	for _, id := range msg.Properties.All() {
		p := msg.Properties.Get(id)
		fmt.Printf("0x%04X (type 0x%04X): %v\n", uint16(id), uint16(p.Type), p.Value)
	}
}

func extractAttachments(msg *gomsg.Message, dir string) {
	if len(msg.Attachments) == 0 {
		fmt.Println("No attachments.")
		return
	}

	if err := os.MkdirAll(dir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating directory: %v\n", err)
		os.Exit(1)
	}

	for i, a := range msg.Attachments {
		if a.IsEmbeddedMessage() {
			fmt.Printf("Skipping embedded MSG: %s\n", a.DisplayName())
			continue
		}
		data := a.Data()
		if data == nil {
			fmt.Printf("Skipping attachment with no data: %s\n", a.DisplayName())
			continue
		}

		name := a.DisplayName()
		path := filepath.Join(dir, name)

		// Avoid overwriting.
		if _, err := os.Stat(path); err == nil {
			path = filepath.Join(dir, fmt.Sprintf("%d_%s", i, name))
		}

		if err := os.WriteFile(path, data, 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing %s: %v\n", name, err)
			continue
		}
		fmt.Printf("Extracted: %s (%d bytes)\n", path, len(data))
	}
}

func senderAddr(msg *gomsg.Message) string {
	if msg.SenderSMTP != "" {
		return msg.SenderSMTP
	}
	return msg.SenderEmail
}

func importanceStr(imp gomsg.Importance) string {
	switch imp {
	case gomsg.ImportanceLow:
		return "Low"
	case gomsg.ImportanceHigh:
		return "High"
	default:
		return "Normal"
	}
}
