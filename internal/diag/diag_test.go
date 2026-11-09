package diag

import (
	"strings"
	"testing"
)

func TestKeepsLatest(t *testing.T) {
	for i := 0; i < keep+10; i++ {
		Logf("line %d", i)
	}
	got := strings.Split(Lines(), "\n")
	if len(got) != keep || !strings.HasSuffix(got[len(got)-1], "line 409") || !strings.HasSuffix(got[0], "line 10") {
		t.Fatalf("%d lines, first %q, last %q", len(got), got[0], got[len(got)-1])
	}
}
