package cache

import (
	"database/sql"
	"fmt"
	"github.com/golang/protobuf/ptypes"
	"google.golang.org/protobuf/types/known/anypb"
	"log"
	common "nephomancy/common/resources"
	"nephomancy/aws/resources"
)

func checkVmSpec(db *sql.DB, avm resources.Ec2VM, spec common.Instance) error {
	avmLocation, err := resolveLocation(avm.Region)
	if err != nil {
		return err
	}
	if l := spec.Location; l != nil {
		if err := common.CheckLocation(avmLocation, *l); err != nil {
			return err
		}
	}
	return nil
}

func resolveLocation(region string) (common.Location, error) {
	cc := CountryByRegion(region)
	if cc == "Unknown" {
		return common.Location{}, fmt.Errorf("unknown region %s", region)
	}
	return common.CountryCodeToLocation(cc)
}

func FillInProviderDetails(db *sql.DB, p *common.Project) error {
	locations := make(map[string]string)
        for _, vmset := range p.InstanceSets {
                if vmset.Template.Location == nil {
                        return fmt.Errorf("missing vmset location information")
                }
                if vmset.Template.Type == nil {
                        return fmt.Errorf("missing vmset type information")
                }
                if vmset.Template.Os == "" {
                        return fmt.Errorf("missing vmset os information")
                }
                if vmset.Template.ProviderDetails == nil {
                        vmset.Template.ProviderDetails = make(map[string](*anypb.Any))
                }
                // Special case: there already are provider details.
                // Make sure they are consistent, otherwise print a warning.
                if vmset.Template.ProviderDetails[resources.AwsProvider] != nil {
			var avm resources.Ec2VM
			err := ptypes.UnmarshalAny(vmset.Template.ProviderDetails[
			resources.AwsProvider], &avm)
			if err != nil {
				return err
			}
			if err = checkVmSpec(db, avm, *vmset.Template); err != nil {
				return err
			}
			log.Printf("Instance Set %s already has details for provider %s, leaving them a they are.\n",
			vmset.Name, resources.AwsProvider)
			locstring := common.PrintLocation(*vmset.Template.Location)
			if locations[locstring] == "" {
				locations[locstring] = avm.Region
			}
		} else {  // no provider details yet
			regions := RegionsForLocation(*vmset.Template.Location, "")
			if len(regions) == 0 {
				return fmt.Errorf("provider %s does not support regions matching location %v",
				resources.AwsProvider, vmset.Template.Location)
			}
			it, r, err := getInstanceTypeForSpec(db, *vmset.Template.Type, regions)
			if err != nil {
				return err
			}
			// vmset.Template.Os

			details, err := ptypes.MarshalAny(&resources.Ec2VM{
				InstanceType: it,
				Region: r[0],
			})
			if err != nil {
				return err
			}
			vmset.Template.ProviderDetails[resources.AwsProvider] = details
		}
	}
	for _, nw := range p.Networks {
		_ = nw
	}
	for _, dset := range p.DiskSets {
		_ = dset
	}

	return nil
}
