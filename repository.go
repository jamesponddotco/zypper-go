package zypper

// RepositoryService handles all operations related to openSUSE repositories.
type RepositoryService service

// Repository represents an openSUSE repository.
type Repository struct {
	Alias        string `xml:"alias,attr"`
	Name         string `xml:"name,attr"`
	Type         string `xml:"type,attr"`
	GPGKey       string `xml:"gpgkey,attr"`
	URL          string `xml:"url"`
	Priority     int    `xml:"priority,attr"`
	Enabled      bool   `xml:"enabled,attr"`
	Autorefresh  bool   `xml:"autorefresh,attr"`
	GPGCheck     bool   `xml:"gpgcheck,attr"`
	RepoGPGCheck bool   `xml:"repo_gpgcheck,attr"`
	PkgGPGCheck  bool   `xml:"pkg_gpgcheck,attr"`
	RawGPGCheck  bool   `xml:"raw_gpgcheck,attr"`
}
