package provider

import (
	"errors"
	"math/rand"
	"time"
)

type MockSender struct{}

func init() {
	rand.Seed(time.Now().UnixNano())
}

func NewMockSender() EmailSender {
	return &MockSender{}
}

func (m *MockSender) Send(to, subject, body string) error {
	time.Sleep(1 * time.Second)

	if to == "fail@example.com" {
		return errors.New("forced failure for testing retries")
	}

	if rand.Intn(3) == 0 {
		return errors.New("mock random failure")
	}

	return nil
}
