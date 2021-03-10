package assets

import (
	"strings"
)

type OsChoice int

const (
	UnspecifiedOS OsChoice = iota
	CentOs
	ContainerOptimizedOs
	Debian
	DeepLearningOnLinux
	FedoraCoreOs
	RedHatEnterpriseLinux
	RedHatEnterpriseLinuxForSAP
	SQLServerOnWindowsServer
	SUSELinuxEnterpriseServer
	SUSELinuxEnterpriseServerForSAP
	Ubuntu
	WindowsServer
	MaxOs // Sentinel value
)

func (o OsChoice) String() string {
	return []string{
		"Unspecified",
		"CentOS",
		"Container Optimized OS",
		"Debian",
		"Deep Learning on Linux",
		"Fedora Core OS",
		"Red Hat Enterprise Linux",
		"Red Hat Enterprise Linux for SAP",
		"SQL Sever On Windows Server",
		"SUSE Linux Enterprise Server",
		"SUSE Linux Enterprise Server for SAP",
		"Ubuntu",
		"Windows Server",
		"Unspecified",
	}[o]
}

// License names are usually of the form
// <base os name>-<version>
func OsFromLicenseName(lname string) string {
	parts := strings.Split(lname, "-")
	basename := strings.ToLower(parts[0])
	switch basename {
	case "ubuntu":
		return Ubuntu.String()
	case "centos":
		return CentOs.String()
	case "debian":
		return Debian.String()
	case "fedora":
		return FedoraCoreOs.String()
	case "rhel":
		return RedHatEnterpriseLinux.String()
	case "windows":
		return WindowsServer.String()
	case "sles":
		return SUSELinuxEnterpriseServer.String()
	default:
		return Debian.String()
	}
}
