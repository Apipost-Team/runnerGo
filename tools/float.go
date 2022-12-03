package tools

import (
	"fmt"
	"strconv"
)

func Decimal2(value float64) float64 {
	value, _ = strconv.ParseFloat(fmt.Sprintf("%.3f", value), 64)
	return value
}
