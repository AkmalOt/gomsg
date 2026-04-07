// Package gomsg parses Microsoft Outlook .msg files and extracts
// email fields such as subject, sender, recipients, body, and attachments.
//
// Usage:
//
//	msg, err := gomsg.Open("email.msg")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(msg.Subject)
//	fmt.Println(msg.SenderName, msg.SenderEmail)
//	for _, a := range msg.Attachments {
//	    fmt.Println(a.FileName, a.Size)
//	}
package gomsg
