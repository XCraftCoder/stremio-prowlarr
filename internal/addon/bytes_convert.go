package addon

import (
	"fmt"
	"strings"
)

const (
	BYTE = 1.0 << (10 * iota)
	KIBIBYTE
	MEBIBYTE
	GIBIBYTE
)

func bytesConvert(bytes uint64) string {
	unit := ""
	value := float32(bytes)

	switch {
	case bytes >= GIBIBYTE:
		unit = "GB"
		value = value / GIBIBYTE
	case bytes >= MEBIBYTE:
		unit = "MB"
		value = value / MEBIBYTE
	case bytes >= KIBIBYTE:
		unit = "KB"
		value = value / KIBIBYTE
	case bytes >= BYTE:
		unit = "B"
	case bytes == 0:
		return "?"
	}

	stringValue := strings.TrimSuffix(
		fmt.Sprintf("%.2f", value), ".00",
	)

	return fmt.Sprintf("%s %s", stringValue, unit)
}
