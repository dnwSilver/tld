package app

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/charmbracelet/x/term"
)

func readPassphrase(in *os.File, out io.Writer) (string, error) {
	fd := in.Fd()
	if !term.IsTerminal(fd) {
		return "", errors.New("database passphrase requires an interactive terminal")
	}

	if _, err := fmt.Fprint(out, "Database passphrase: "); err != nil {
		return "", fmt.Errorf("write passphrase prompt: %w", err)
	}

	passphrase, err := term.ReadPassword(fd)
	if _, printErr := fmt.Fprintln(out); printErr != nil && err == nil {
		err = printErr
	}
	if err != nil {
		return "", fmt.Errorf("read database passphrase: %w", err)
	}

	return string(passphrase), nil
}
