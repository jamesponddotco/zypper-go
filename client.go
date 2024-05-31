package zypper

import (
	"context"
	"fmt"
	"log/slog"
	"os/exec"

	"git.sr.ht/~jamesponddotco/xstd-go/xerrors"
)

const (
	// ErrZypperNotFound is returned when zypper cannot be found in PATH.
	ErrZypperNotFound xerrors.Error = "zypper binary not found in PATH"

	// ErrZypperSubCommand is a generic error returned when zypper fails to run
	// a sub-command.
	ErrZypperSubCommand xerrors.Error = "failed to run zypper sub-command"
)

const _zypper string = "zypper"

type (
	// Service is a common struct that can be reused instead of allocating a new
	// one for each service on the heap.
	service struct {
		client *Client
	}

	// Client is the client for interacting with the openSUSE package manager.
	Client struct {
		// Logger is the structured logger used by the client. If not specified,
		// no logging will be done.
		Logger *slog.Logger

		// Services are the service instances that can be used to interact with
		// the zypper package manager.
		Package    *PackageService
		Repository *RepositoryService

		// common specifies a common service shared by all services.
		common service

		// Path specifies the path to the zypper binary including the binary
		// itself. If not specified, the binary will be looked up in PATH.
		Path string
	}
)

// NewClient creates a new Client instance with the given logger and path.
func NewClient(logger *slog.Logger, path string) (*Client, error) {
	var err error

	if path == "" {
		path, err = exec.LookPath(_zypper)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", ErrZypperNotFound, err)
		}
	}

	if err := exec.Command(path).Run(); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrZypperNotFound, err)
	}

	client := &Client{
		Logger: logger,
		Path:   path,
	}

	client.common.client = client
	client.Package = (*PackageService)(&client.common)
	client.Repository = (*RepositoryService)(&client.common)

	return client, nil
}

// Do runs zypper sub-commands with the given arguments and returns its output.
func (c *Client) Do(ctx context.Context, subCommand string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext( //nolint:gosec // the caller is responsible for sanitizing the input
		ctx,
		c.Path,
		"--quiet",
		"--non-interactive",
		"--non-interactive-include-reboot-patches",
		"--xmlout",
		subCommand,
	)
	cmd.Args = append(cmd.Args, args...)

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("%w: %s: %w", ErrZypperSubCommand, subCommand, err)
	}

	return output, nil
}
