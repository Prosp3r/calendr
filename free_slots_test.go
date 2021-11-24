package main

import (
	"fmt"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	rc := m.Run()

	if rc == 0 && testing.CoverMode() != "" {
		c := testing.Coverage()
		if c < 0.5 {
			fmt.Println("Testing passed but coverage failed at", c)
			rc = -1
		}
	}
	os.Exit(rc)
}

func TestFreeSlots(t *testing.T) {

	// expectedSlots :=
}

func TestConvertTime(t *testing.T) {

}
