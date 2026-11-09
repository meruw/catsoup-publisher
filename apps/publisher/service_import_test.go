package main

import (
	"errors"
	"testing"
)

func TestImportDialogCancellationAndFailure(t *testing.T) {
	s := &PublisherService{chooseDelivery: func() (string, error) { return "", nil }}
	r, err := s.ImportDelivery()
	if r != nil || err != nil {
		t.Fatal(r, err)
	}
	s.chooseDelivery = func() (string, error) { return "", errors.New("dialog unavailable") }
	if _, err = s.ImportDelivery(); err == nil {
		t.Fatal("dialog error lost")
	}
	s.importing.Lock()
	defer s.importing.Unlock()
	if _, err = s.ImportDelivery(); err == nil {
		t.Fatal("concurrent import accepted")
	}
}
