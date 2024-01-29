package test

import (
	"testing"
)

func FirstFailTest(t *testing.T) {
	t.Fatalf(`Hello("") = %q, %v, want "", error`)
}
