package main

import "testing"

func TestHumanBytes(t *testing.T) {
	tests := []struct {
		in   int64
		want string
	}{
		{0, "0 B"},
		{1023, "1023 B"},
		{1024, "1.00 KiB"},
		{1024 * 1024, "1.00 MiB"},
		{1149239296, "1.07 GiB"},
	}
	for _, tt := range tests {
		if got := humanBytes(tt.in); got != tt.want {
			t.Fatalf("humanBytes(%d) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestGroupUint(t *testing.T) {
	if got, want := groupUint(197934), "197,934"; got != want {
		t.Fatalf("groupUint = %q, want %q", got, want)
	}
}

func TestRedactPathStable(t *testing.T) {
	a := redactPath(`C:\Users\Max\OneDrive\Foo.txt`)
	b := redactPath(`c:\users\max\onedrive\foo.txt`)
	if a != b {
		t.Fatalf("redaction should be case-insensitive: %q != %q", a, b)
	}
	if a == `C:\Users\Max\OneDrive\Foo.txt` {
		t.Fatal("redaction returned original path")
	}
}
