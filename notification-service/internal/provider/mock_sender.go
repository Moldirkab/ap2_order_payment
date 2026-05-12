package provider

import (
	"errors"
	"math/rand"
	"time"
)

type MockSender struct{}

func NewMockSender() EmailSender {
	return &MockSender{}
}

func (m *MockSender) Send(to, subject, body string) error {
	time.Sleep(1 * time.Second)

	if rand.Intn(3) == 0 {
		return errors.New("mock email failure")
	}

	return nil
}
