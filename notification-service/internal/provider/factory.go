package provider

import "os"

func NewEmailSender() EmailSender {
	mode := os.Getenv("PROVIDER_MODE")

	if mode == "REAL" {
		return NewSMTPSender()
	}

	return NewMockSender()
}
