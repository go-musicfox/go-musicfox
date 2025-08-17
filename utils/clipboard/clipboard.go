package clipboard

import (
	"github.com/atotto/clipboard"
)

func Read() (string, error) {
	return clipboard.ReadAll()
}

func Write(text string) error {
	return clipboard.WriteAll(text)
}
