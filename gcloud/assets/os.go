package assets

import (
	"fmt"
	"strings"
)

type OsChoice int

const (
	UnspecifiedOs OsChoice = iota
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

// TODO: verify this mapping, or find a better way to locate SKUs for
// operating system licenses. Can simplify based on which os choices
// are free.
func (o OsChoice) ResourceGroup() string {
	return []string{
		"Unspecified",
		"CentOS",
		"CoreOSStable", // is this right?
		"Debian",
		"Debian",       // you can choose Debian as the version with this OS.
		"FedoraCoreOS", // There are vaiants of this called Testing, Stable, Next
		// There's also entries for sql server in the Google and
		// Microsoft groups.
		"SQLServer2016Standard", // varies by vcpu count
		"RHEL7",                 // There are variants of this depending on vcpu count
		// For RHEL 7 with SAP, use ResourceGroup Google and look for
		// "RHEL 7 for SAP". Need to distinguish SAP HANA from SAP Applications.
		"Google",
		// For SUSE Enterprise without SAP, use ResourceGroup Google.
		"Google",
		// This is actually mostly the SUSE group
		"SLES12ForSAP", // there is no cpu cost for these?
		"Ubuntu1604",   // both UbuntuDev and Core have an extra CPU SKU.
		// There are variants of this but luckily they cost the same ...
		"WindowsServer2012",
		"Unspecified",
	}[o]
}

func OsChoiceByName(name string) OsChoice {
	lcname := strings.ToLower(name)
	for os := OsChoice(1); os < MaxOs; os++ {
		if lcname == strings.ToLower(os.String()) {
			return os
		}
	}
	return UnspecifiedOs
}

func IsLinux(os OsChoice) bool {
	return os != WindowsServer
}

func IsWindows(os OsChoice) bool {
	return os == WindowsServer
}

// Experimental. If specOs describes actualOs, return nil.
// Example: if specOs is "Linux", actualOs should be a linux
// variant (can be free or enterprise).
func DoesOsMatch(specOs string, actualOs string) error {
	actual := OsChoiceByName(actualOs)
	if actual == UnspecifiedOs {
		return nil // Not sure if this should match or be an error
	}
	spec := strings.ToLower(specOs)
	if spec == strings.ToLower(actual.String()) {
		return nil
	}
	if spec == "linux" {
		if IsLinux(actual) {
			return nil
		}
		return fmt.Errorf("specified os %s does not match actual %s",
			specOs, actualOs)
	}
	if spec == "windows" {
		if IsWindows(actual) {
			return nil
		}
		return fmt.Errorf("specified os %s does not match actual %s",
			specOs, actualOs)
	}
	return fmt.Errorf("unrecognised os spec: %s", specOs)
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
