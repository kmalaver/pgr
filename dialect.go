package pgr

import (
	"fmt"
	"strings"
	"time"
)

const (
	quote      = `"`
	timeFormat = "2006-01-02 15:04:05.000000"
)

func QuoteIdent(s string) string {
	part := strings.SplitN(s, ".", 2)
	if len(part) == 2 {
		return QuoteIdent(part[0]) + "." + QuoteIdent(part[1])
	}
	return quote + s + quote
}

func EncodeString(s string) string {
	return `'` + strings.Replace(s, `'`, `''`, -1) + `'`
}

func EncodeBool(b bool) string {
	if b {
		return "TRUE"
	}
	return "FALSE"
}

func EncodeTime(t time.Time) string {
	return `'` + t.UTC().Format(timeFormat) + `'`
}

func EncodeBytes(b []byte) string {
	return fmt.Sprintf(`E'\\x%x'`, b)
}

func Placeholder(n int) string {
	return fmt.Sprintf("$%d", n+1)
}
