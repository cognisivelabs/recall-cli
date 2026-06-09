package main

import (
	"bytes"

	"github.com/CognisiveLabs/recall-cli/internal/storage"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// mockStore implements storage.Storage for testing without a real database.
type mockStore struct {
	commands []storage.Command
	nextID   int
}

func newMockStore() *mockStore {
	return &mockStore{nextID: 1}
}

func (m *mockStore) List() ([]storage.Command, error) {
	return m.commands, nil
}

func (m *mockStore) GetByID(id int) (*storage.Command, error) {
	for i := range m.commands {
		if m.commands[i].ID == id {
			return &m.commands[i], nil
		}
	}
	return nil, nil
}

func (m *mockStore) GetByPattern(pattern string) (*storage.Command, error) {
	for i := range m.commands {
		if m.commands[i].Pattern == pattern {
			return &m.commands[i], nil
		}
	}
	return nil, nil
}

func (m *mockStore) Delete(id int) error {
	for i := range m.commands {
		if m.commands[i].ID == id {
			m.commands = append(m.commands[:i], m.commands[i+1:]...)
			return nil
		}
	}
	return nil
}

func (m *mockStore) Update(cmd storage.Command) error {
	for i := range m.commands {
		if m.commands[i].ID == cmd.ID {
			m.commands[i] = cmd
			return nil
		}
	}
	return nil
}

func (m *mockStore) Upsert(cmd storage.Command) error {
	for i := range m.commands {
		if m.commands[i].Pattern == cmd.Pattern {
			m.commands[i].Description = cmd.Description
			m.commands[i].Tags = cmd.Tags
			m.commands[i].Source = cmd.Source
			return nil
		}
	}
	cmd.ID = m.nextID
	m.nextID++
	if cmd.Source == "" {
		cmd.Source = "local"
	}
	m.commands = append(m.commands, cmd)
	return nil
}

func (m *mockStore) RecordUsage(id int) error {
	for i := range m.commands {
		if m.commands[i].ID == id {
			m.commands[i].UsageCount++
			return nil
		}
	}
	return nil
}

func (m *mockStore) Close() error { return nil }

// seed adds a command to the mock store. Returns the store for chaining.
func (m *mockStore) seed(cmd storage.Command) *mockStore {
	m.Upsert(cmd)
	return m
}

// execCmd runs a cobra command with the given args and returns captured stdout, stderr, and error.
// Resets the command's flag state after execution to prevent leaks between tests.
func execCmd(cmd *cobra.Command, args ...string) (stdout string, stderr string, err error) {
	outBuf := &bytes.Buffer{}
	errBuf := &bytes.Buffer{}
	cmd.SetOut(outBuf)
	cmd.SetErr(errBuf)
	cmd.SetArgs(args)
	err = cmd.Execute()

	// Reset flags so a reused cmd doesn't carry state from a previous call
	cmd.Flags().Visit(func(f *pflag.Flag) {
		f.Changed = false
		f.Value.Set(f.DefValue)
	})

	return outBuf.String(), errBuf.String(), err
}
