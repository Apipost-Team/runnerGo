package tools

import (
	"bytes"
	"fmt"
)

func FloatToPercent(num float64) string {
	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("%2.2f", num*100))
	// buf.WriteString(fmt.Sprintf("  %2.2f", num*100))
	buf.WriteString("%")
	return buf.String()
}
