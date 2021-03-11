package cache

import (
	"database/sql"
	"fmt"
	"github.com/golang/protobuf/ptypes"
	_ "github.com/mattn/go-sqlite3"
	"google.golang.org/protobuf/types/known/anypb"
	"log"
	common "nephomancy/common/resources"
	"nephomancy/dcs/resources"
	"strings"
)

// Returns nil if spec location is compatible with Switzerland,
// an error otherwise.
func checkLocation(spec common.Location) error {
	if spec.GlobalRegion != "" && spec.GlobalRegion != "EMEA" {
		return fmt.Errorf("spec global region %s does not allow a provider in EMEA",
			spec.GlobalRegion)
	}
	if spec.Continent != "" && spec.Continent != "Europe" {
		return fmt.Errorf("spec continent %s does not allow a provider in Switzerland",
			spec.Continent)
	}
	if spec.CountryCode != "" && spec.CountryCode != "CH" {
		return fmt.Errorf("spec country code is %s but %s is only available in Switzerland",
			spec.CountryCode, resources.DcsProvider)
	}
	return nil
}

// Based on the vmset.Template.Os setting, choose an OS available on DCS.
func chooseOs(templateOs string) string {
	if strings.ToLower(templateOs) == "windows" {
		return "Windows"
	} else {
		return "Red Hat" // Is this a good default? Windows costs less.
	}
}

func isVmConsistent(dcsVm resources.DcsVM, template common.Instance) error {
	if template.Os == "" {
		return nil
	}
	if dcsVm.OsChoice != chooseOs(template.Os) {
		return fmt.Errorf("template os %s does not match provider settings %s\n",
			template.Os, dcsVm.OsChoice)
	}
	return nil
}

func FillInProviderDetails(db *sql.DB, p *common.Project) error {
	if p.ProviderDetails == nil {
		p.ProviderDetails = make(map[string](*anypb.Any))
	}
	sla := "Basic"
	if p.ProviderDetails[resources.DcsProvider] == nil {
		details, _ := ptypes.MarshalAny(&resources.DcsProject{
			Sla: sla,
		})
		p.ProviderDetails[resources.DcsProvider] = details
	} else {
		var dcsProject resources.DcsProject
		err := ptypes.UnmarshalAny(p.ProviderDetails[resources.DcsProvider],
			&dcsProject)
		if err != nil {
			return err
		}
		sla = dcsProject.Sla
	}
	for _, vmset := range p.InstanceSets {
		if vmset.Template.Location == nil {
			return fmt.Errorf("missing vmset location information")
		}
		if err := checkLocation(*vmset.Template.Location); err != nil {
			return err
		}
		if vmset.Template.Type == nil {
			return fmt.Errorf("missing vmset type information")
		}
		if vmset.Template.ProviderDetails == nil {
			vmset.Template.ProviderDetails = make(map[string](*anypb.Any))
		}
		// Special case: already have provider details. Consistency check.
		if vmset.Template.ProviderDetails[resources.DcsProvider] != nil {
			var dcsvm resources.DcsVM
			err := ptypes.UnmarshalAny(vmset.Template.ProviderDetails[resources.DcsProvider], &dcsvm)
			if err != nil {
				return err
			}
			if err = isVmConsistent(dcsvm, *vmset.Template); err != nil {
				return err
			}
			log.Printf("Instance Set %s already has details for provider %s, leaving them as they are.\n", vmset.Name, resources.DcsProvider)
		} else { // Normal case: no provider details yet.
			os := chooseOs(vmset.Template.Os)
			details, _ := ptypes.MarshalAny(&resources.DcsVM{
				OsChoice: os,
			})
			vmset.Template.ProviderDetails[resources.DcsProvider] = details
		}
	}
	for _, dset := range p.DiskSets {
		if dset.Template.Location == nil {
			return fmt.Errorf("missing disk set location information")
		}
		if err := checkLocation(*dset.Template.Location); err != nil {
			return err
		}
		if dset.Template.Type == nil {
			return fmt.Errorf("missing disk set type information")
		}
		if dset.Template.ProviderDetails == nil {
			dset.Template.ProviderDetails = make(map[string](*anypb.Any))
		}
		if dset.Template.ProviderDetails[resources.DcsProvider] != nil {
			var dcsdisk resources.DcsDisk
			err := ptypes.UnmarshalAny(dset.Template.ProviderDetails[resources.DcsProvider], &dcsdisk)
			if err != nil {
				return err
			}
			log.Printf("Disk Set %s already has details for provider %s, leaving them as they are.\n", dset.Name, resources.DcsProvider)
		} else {
			tech := dset.Template.Type.DiskTech
			speed := "Fast"
			if tech == "SSD" {
				speed = "UltraFast"
			}
			details, _ := ptypes.MarshalAny(&resources.DcsDisk{
				DiskType:   speed,
				WithBackup: false,
			})
			dset.Template.ProviderDetails[resources.DcsProvider] = details
		}
	}
	for _, nw := range p.Networks {
		for _, gw := range nw.Gateways {
			if gw.ProviderDetails == nil {
				gw.ProviderDetails = make(map[string](*anypb.Any))
			}
			if gw.ProviderDetails[resources.DcsProvider] != nil {
				var dcsgw resources.DcsGateway
				err := ptypes.UnmarshalAny(gw.ProviderDetails[resources.DcsProvider], &dcsgw)
				if err != nil {
					return err
				}
				log.Printf("Gateway already has details for provider %s, leaving it as it is.\n", resources.DcsProvider)
			} else {
				details, _ := ptypes.MarshalAny(&resources.DcsGateway{
					Type: "Eco",
				})
				gw.ProviderDetails[resources.DcsProvider] = details
			}
		}
	}
	return nil
}
