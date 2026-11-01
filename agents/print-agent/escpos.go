package printagent

import (
	"bytes"
	"errors"
	"strings"
)

var (
	initPrinter = []byte{0x1b, 0x40}
	alignCenter = []byte{0x1b, 0x61, 0x01}
	alignLeft   = []byte{0x1b, 0x61, 0x00}
	cutPaper    = []byte{0x1d, 0x56, 0x00}
)

type Receipt struct {
	Header string
	Lines  []string
	Total  string
}

func EncodeReceipt(receipt Receipt) ([]byte, error) {
	if strings.TrimSpace(receipt.Header) == "" || strings.TrimSpace(receipt.Total) == "" {
		return nil, errors.New("receipt header and total are required")
	}
	if len(receipt.Header)+len(receipt.Total) > 512 || len(receipt.Lines) > 100 {
		return nil, errors.New("receipt is too large")
	}
	lineBytes := 0
	for _, line := range receipt.Lines {
		lineBytes += len(line)
		if len(line) > 512 || lineBytes > 12*1024 {
			return nil, errors.New("receipt lines are too large")
		}
	}
	var output bytes.Buffer
	output.Write(initPrinter)
	output.Write(alignCenter)
	output.WriteString(printerSafeText(receipt.Header))
	output.WriteByte('\n')
	output.Write(alignLeft)
	for _, line := range receipt.Lines {
		if line == "" {
			continue
		}
		output.WriteString(printerSafeText(line))
		output.WriteByte('\n')
	}
	output.WriteString("--------------------------------\n")
	output.WriteString("TOTAL: ")
	output.WriteString(printerSafeText(receipt.Total))
	output.WriteString("\n\n")
	output.Write(cutPaper)
	return output.Bytes(), nil
}

func printerSafeText(value string) string {
	var safe strings.Builder
	for _, char := range value {
		if char == '\n' || char == '\t' || char >= 0x20 && char != 0x7f {
			safe.WriteRune(char)
		}
	}
	return safe.String()
}
