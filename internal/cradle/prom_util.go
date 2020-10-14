package cradle

import (
	"regexp"
	"strings"
)

var newLine = regexp.MustCompile(`\n`)

func promCommentOut(in string) string {
	var b strings.Builder
	lines := newLine.Split(in, -1)
	for idx, line := range lines {
		if idx == len(lines)-1 && len(line) == 0 {
			b.WriteString("\n")
			continue
		}
		b.WriteString("### ")
		b.WriteString(line)
		b.WriteString("\n")
	}
	return b.String()
}
