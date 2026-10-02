package projectsync

import (
	"errors"
	"testing"
)

func TestScannerOutputEnforcesLimit(t *testing.T) {
	buffer := scannerOutput{max: 4}
	if n, err := buffer.Write([]byte("abcd")); n != 4 || err != nil {
		t.Fatalf("first write = %d, %v", n, err)
	}
	if n, err := buffer.Write([]byte("e")); n != 0 || !errors.Is(err, errScannerOutputLimit) {
		t.Fatalf("overflow = %d, %v", n, err)
	}
	if got := string(buffer.Bytes()); got != "abcd" {
		t.Fatalf("buffer = %q", got)
	}
}
