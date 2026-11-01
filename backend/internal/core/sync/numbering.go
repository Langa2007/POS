package sync

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

func OfflineReceiptNumber(terminalPrefix string, sequence uint64) (string, error) {
	prefix := strings.ToUpper(strings.TrimSpace(terminalPrefix))
	if prefix == "" || len(prefix) > 12 {
		return "", errors.New("terminal prefix must contain 1 to 12 characters")
	}
	if sequence == 0 {
		return "", errors.New("sequence must be positive")
	}
	for _, char := range prefix {
		if !((char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9') || char == '-') {
			return "", errors.New("terminal prefix contains unsupported characters")
		}
	}
	return fmt.Sprintf("%s-%06s", prefix, strconv.FormatUint(sequence, 10)), nil
}
