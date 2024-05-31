package zypper

import (
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"os/exec"

	"git.sr.ht/~jamesponddotco/xstd-go/xerrors"
)

const (
	// ErrEmptyPackageName is returned when an empty package name is given as
	// input to one of the package methods.
	ErrEmptyPackageName xerrors.Error = "package name cannot be empty"

	// ErrZypperSearch is a generic error returned when zypper fails to run the
	// search sub-command.
	ErrZypperSearch xerrors.Error = "failed to search for package"

	// ErrNoMatchingItem is returned when no package is found matching the given
	// name.
	ErrNoMatchingItem xerrors.Error = "no matching item found"

	// ErrXMLUnmarshal is returned when an XML unmarshal error occurs.
	ErrXMLUnmarshal xerrors.Error = "failed to unmarshal XML"

	// ErrRootPrivileges is returned when root privileges are required to run a
	// sub-command but none are available.
	ErrRootPrivileges xerrors.Error = "root privileges required to run command"

	// ErrInstallFailed is returned when zypper fails to install a package.
	ErrInstallFailed xerrors.Error = "failed to install package"
)

// Package types as defined by zypper.
const (
	PackageTypePackage = "package"
	PackageTypePatch   = "patch"
	PackageTypePattern = "pattern"
	PackageTypeProduct = "product"
)

// Package status as defined by zypper.
const (
	PackageStatusInstalled    = "installed"
	PackageStatusNotInstalled = "not-installed"
)

// Common sub-commands used by the Package category of the package manager.
const (
	_commandSearch  string = "search"
	_commandInstall string = "install"
)

// PackageService handles all operations related to openSUSE packages.
type PackageService service

// Package represents an openSUSE package in the repository.
type Package struct {
	Name       string     `xml:"name,attr"`
	Status     string     `xml:"status,attr"`
	Kind       string     `xml:"kind,attr"`
	Version    string     `xml:"edition,attr"`
	Arch       string     `xml:"arch,attr"`
	Repository Repository `xml:"repository,attr"`
}

// Search searches for a package in all existing repositories.
func (s *PackageService) Search(ctx context.Context, name string) ([]Package, error) {
	if name == "" {
		return nil, ErrEmptyPackageName
	}

	args := []string{
		"--details",
		name,
	}

	output, err := s.client.Do(ctx, _commandSearch, args...)
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

// Install installs a package from the repositories enabled in the system.
func (s *PackageService) Install(ctx context.Context, name string, args ...string) error {
	if name == "" {
		return ErrEmptyPackageName
	}

	args = append(args, name)

	_, err := s.client.Do(ctx, _commandInstall, args...)
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
