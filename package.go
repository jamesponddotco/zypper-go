package zypper

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

// Package represents an openSUSE package in the repository.
type Package struct {
	Name       string     `xml:"name,attr"`
	Status     string     `xml:"status,attr"`
	Kind       string     `xml:"kind,attr"`
	Version    string     `xml:"edition,attr"`
	Arch       string     `xml:"arch,attr"`
	Repository Repository `xml:"repository,attr"`
}
