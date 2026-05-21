package app

import (
	"context"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dnwSilver/tld/internal/storage"
)

func Run() error {
	ctx := context.Background()

	passphrase, err := readPassphrase(os.Stdin, os.Stderr)
	if err != nil {
		return err
	}

	store, err := storage.Open(ctx, "", passphrase)
	if err != nil {
		return fmt.Errorf("open storage: %w", err)
	}
	defer func() {
		_ = store.Close()
	}()

	_, err = tea.NewProgram(newModel(store), tea.WithAltScreen()).Run()
	return err
}
