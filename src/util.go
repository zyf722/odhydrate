package main

import (
	"crypto/sha256"
	"fmt"
	"strconv"
	"strings"
)

func humanBytes(n int64) string {
	if n < 0 {
		return "-" + humanBytes(-n)
	}
	const unit = 1024
	if n < unit {
		return fmt.Sprintf("%d B", n)
	}
	div, exp := int64(unit), 0
	for q := n / unit; q >= unit && exp < 5; q /= unit {
		div *= unit
		exp++
	}
	units := []string{"KiB", "MiB", "GiB", "TiB", "PiB", "EiB"}
	return fmt.Sprintf("%.2f %s", float64(n)/float64(div), units[exp])
}

func groupUint(n uint64) string {
	s := strconv.FormatUint(n, 10)
	if len(s) <= 3 {
		return s
	}
	first := len(s) % 3
	if first == 0 {
		first = 3
	}
	var b strings.Builder
	b.Grow(len(s) + len(s)/3)
	b.WriteString(s[:first])
	for i := first; i < len(s); i += 3 {
		b.WriteByte(',')
		b.WriteString(s[i : i+3])
	}
	return b.String()
}

func redactPath(path string) string {
	sum := sha256.Sum256([]byte(strings.ToLower(path)))
	return fmt.Sprintf("<path:%x>", sum[:6])
}

func displayPath(path string, redact bool) string {
	if redact {
		return redactPath(path)
	}
	return path
}

func maxSeconds(v float64) float64 {
	if v <= 0 {
		return 1e-9
	}
	return v
}
