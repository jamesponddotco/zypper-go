// Package zypper provides a client for openSUSE's package manager.
package zypper

import (
	"encoding/xml"
	"errors"
	"fmt"
	"os/exec"

	"git.sr.ht/~jamesponddotco/xstd-go/xerrors"
)

const (
	// ErrZypperNotFound is returned when zypper cannot be found in PATH.
	ErrZypperNotFound xerrors.Error = "zypper binary not found in PATH"

	// ErrZypperSubCommand is a generic error returned when zypper fails to run
	// a sub-command.
	ErrZypperSubCommand xerrors.Error = "failed to run zypper sub-command"

	// ErrZypperSearch is a generic error returned when zypper fails to run the
	// search sub-command.
	ErrZypperSearch xerrors.Error = "failed to search for package"

	// ErrZypperInstall is a generic error returned when zypper fails to run the
	// install sub-command.
	ErrZypperInstall xerrors.Error = "failed to run zypper install"

	// ErrNoMatchingItem is returned when no package is found matching the given
	// name.
	ErrNoMatchingItem xerrors.Error = "no matching item found"

	// ErrEmptyName is returned when an empty name is given as input.
	ErrEmptyName xerrors.Error = "name cannot be empty"

	// ErrXMLUnmarshal is returned when an XML unmarshal error occurs.
	ErrXMLUnmarshal xerrors.Error = "failed to unmarshal XML"

	// ErrRootPrivileges is returned when root privileges are required to run a
	// sub-command but none are available.
	ErrRootPrivileges xerrors.Error = "root privileges required to run command"

	// ErrInstallFailed is returned when zypper fails to install a package.
	ErrInstallFailed xerrors.Error = "failed to install package"
)

// Common sub-commands used by the package manager.
const (
	CommandSearch  string = "search"
	CommandInstall string = "install"
)

const _zypper string = "zypper"

// Do runs zypper sub-commands with the given arguments and returns its output.
func Do(subCommand string, args ...string) ([]byte, error) {
	_, err := exec.LookPath(_zypper)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrZypperNotFound, err)
	}

	cmd := exec.Command(
		_zypper,
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

// Search searches for a package in all existing repositories.
func Search(name string) ([]Package, error) {
	if name == "" {
		return nil, ErrEmptyName
	}

	output, err := Do(CommandSearch, "--details", name)
	if err != nil {
		var exitErr *exec.ExitError

		if errors.As(err, &exitErr) {
			switch exitErr.ExitCode() {
			case 104:
				return nil, fmt.Errorf("%w: %s", ErrNoMatchingItem, name)
			default:
				return nil, fmt.Errorf("%w %s: %w", ErrZypperSearch, name, err)
			}
		}

		return nil, fmt.Errorf("%w %s: %w", ErrZypperSearch, name, err)
	}

	type SolvablePackage struct {
		RepositoryAttr string `xml:"repository,attr"`
		Package
	}

	type SolvableList struct {
		Packages []SolvablePackage `xml:"solvable"`
	}

	type SearchResult struct {
		XMLName      xml.Name     `xml:"search-result"`
		SolvableList SolvableList `xml:"solvable-list"`
	}

	type xmlStream struct {
		XMLName      xml.Name     `xml:"stream"`
		SearchResult SearchResult `xml:"search-result"`
	}

	var stream xmlStream
	if err := xml.Unmarshal(output, &stream); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrXMLUnmarshal, err)
	}

	packages := make([]Package, 0, len(stream.SearchResult.SolvableList.Packages))

	for i := range stream.SearchResult.SolvableList.Packages {
		pkg := &stream.SearchResult.SolvableList.Packages[i].Package
		pkg.Repository = Repository{
			Name: stream.SearchResult.SolvableList.Packages[i].RepositoryAttr,
		}

		packages = append(packages, *pkg)
	}

	return packages, nil
}

// Install installs a package from the repositories into the system.
func Install(name string, args ...string) error {
	if name == "" {
		return ErrEmptyName
	}

	_, err := Do(CommandInstall, append([]string{name}, args...)...)
	if err != nil {
		var exitErr *exec.ExitError

		if errors.As(err, &exitErr) {
			switch exitErr.ExitCode() {
			case 5:
				return fmt.Errorf("%w: install", ErrRootPrivileges)
			case 104:
				return fmt.Errorf("%w: %s", ErrNoMatchingItem, name)
			default:
				return fmt.Errorf("%w %s: %w", ErrInstallFailed, name, err)
			}
		}

		return fmt.Errorf("%w: %w", ErrInstallFailed, err)
	}

	return nil
}
