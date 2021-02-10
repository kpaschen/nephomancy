package assets

import (
	"context"
	"fmt"
	"google.golang.org/api/compute/v1"
	"strings"
)

func ListRegions(project string) ([]string, error) {
	ctx := context.Background()
	const pageSize int64 = 100
	client, err := compute.NewService(ctx)
	if err != nil {
		return nil, err
	}

	ret := make([]string, 0)
	resp, err := client.Regions.List(project).MaxResults(pageSize).Do()
	if err != nil {
		return nil, err
	}
	for _, t := range resp.Items {
		fmt.Printf("region: %+v\n", t)
		if t.Deprecated != nil {
			state := t.Deprecated.State
			if state == "ACTIVE" || state == "DEPRECATED" {
				fmt.Printf("region %s is going to be deprecated.\n",
					t.Name)
			} else {
				fmt.Printf("region %s has been deprecated\n",
					t.Name)
			}
			continue
		}
		ret = append(ret, t.Name)

	}
	return ret, nil
}

func ListZones(project string) ([]RegionZone, error) {
	ctx := context.Background()
	const pageSize int64 = 100
	client, err := compute.NewService(ctx)
	if err != nil {
		return nil, err
	}

	ret := make([]RegionZone, 0)
	nextPageToken := ""
	for {
		resp, err := client.Zones.List(project).MaxResults(pageSize).PageToken(nextPageToken).Do()
		if err != nil {
			return nil, err
		}
		rt := make([]RegionZone, len(resp.Items))
		for idx, t := range resp.Items {
			x := strings.Split(t.Region, "/")
			region := x[len(x)-1]
			if region != "" && t.Name != "" {
				rt[idx] = RegionZone{
					Region: region,
					Zone:   t.Name,
				}
			}
		}
		ret = append(ret, rt...)
		if resp.NextPageToken == "" {
			break
		}
		nextPageToken = resp.NextPageToken
	}
	return ret, nil
}

func ListMachineTypes(project string, zone string) ([]MachineType, error) {
	ctx := context.Background()
	const pageSize int64 = 100
	client, err := compute.NewService(ctx)
	if err != nil {
		return nil, err
	}

	ret := make([]MachineType, 0)
	nextPageToken := ""
	for {
		resp, err := client.MachineTypes.List(project, zone).MaxResults(pageSize).PageToken(nextPageToken).Do()
		if err != nil {
			return nil, err
		}
		rt := make([]MachineType, len(resp.Items))
		for idx, t := range resp.Items {
			shared := false
			if t.IsSharedCpu {
				shared = true
			}
			at := make([]AcceleratorType, len(t.Accelerators))
			for id2, a := range t.Accelerators {
				at[id2] = AcceleratorType{
					Name:  a.GuestAcceleratorType,
					Count: a.GuestAcceleratorCount,
				}
			}
			gpuCount := 0
			if strings.HasPrefix(t.Name, "a2-") {
				parts := strings.Split(t.Name, "-")
				if len(parts) == 3 {
					n, err := fmt.Scanf(parts[2], "%dg", &gpuCount)
					if err != nil || n != 1 {
						return nil, fmt.Errorf("could not get gcpu count for machine type %s", t.Name)
					}
				}
			}
			rt[idx] = MachineType{
				CpuCount:     uint32(t.GuestCpus),
				IsSharedCpu:  shared,
				MemoryMb:     uint64(t.MemoryMb),
				Name:         t.Name,
				GpuCount:     uint32(gpuCount),
				Accelerators: at,
			}
		}
		ret = append(ret, rt...)
		if resp.NextPageToken == "" {
			break
		}
		nextPageToken = resp.NextPageToken
	}
	return ret, nil
}

func ListDiskTypes(project string, zone string) ([]DiskType, error) {
	ctx := context.Background()
	const pageSize int64 = 100
	client, err := compute.NewService(ctx)
	if err != nil {
		return nil, err
	}

	ret := make([]DiskType, 0)
	nextPageToken := ""
	for {
		resp, err := client.DiskTypes.List(project, zone).MaxResults(pageSize).PageToken(nextPageToken).Do()
		if err != nil {
			return nil, err
		}
		rt := make([]DiskType, len(resp.Items))
		for idx, t := range resp.Items {
			rt[idx] = DiskType{
				Name:          t.Name,
				DefaultSizeGb: t.DefaultDiskSizeGb,
			}
		}
		ret = append(ret, rt...)
		if resp.NextPageToken == "" {
			break
		}
		nextPageToken = resp.NextPageToken
	}
	return ret, nil
}

func ListRegionDiskTypes(project string, region string) ([]DiskType, error) {
	ctx := context.Background()
	const pageSize int64 = 100
	client, err := compute.NewService(ctx)
	if err != nil {
		return nil, err
	}

	ret := make([]DiskType, 0)
	nextPageToken := ""
	for {
		resp, err := client.RegionDiskTypes.List(project, region).MaxResults(pageSize).PageToken(nextPageToken).Do()
		if err != nil {
			return nil, err
		}
		rt := make([]DiskType, len(resp.Items))
		for idx, t := range resp.Items {
			rt[idx] = DiskType{
				Name:          t.Name,
				DefaultSizeGb: t.DefaultDiskSizeGb,
				Region:        region,
			}
		}
		ret = append(ret, rt...)
		if resp.NextPageToken == "" {
			break
		}
		nextPageToken = resp.NextPageToken
	}
	return ret, nil
}

func GetProject(project string) error {
	ctx := context.Background()
	client, err := compute.NewService(ctx)
	if err != nil {
		return err
	}

	resp, err := client.Projects.Get(project).Do()
	if err != nil {
		return err
	}
	for _, q := range resp.Quotas {
		qu, err := q.MarshalJSON()
		if err != nil {
			return err
		}
		fmt.Printf("quota: %+v\n", string(qu))
	}
	return nil
}

// ListInstances; Disks; BackendServices; BackendBuckets; GlobalAddresses; Images; Licenses / LicenseCodes; Networks; NodeGroups; Subnetworks; and more
// there's commitments too

// DisksService: Get(project, zone, disk)  -- calls projects/{project}/zones/{zone}/disks/{disk}
// Disks.List takes project and zone  (these are always zonal disks, use RegionDisks for the others)

func ListDisks(project string) error {
	ctx := context.Background()
	const pageSize int64 = 100
	client, err := compute.NewService(ctx)
	if err != nil {
		return err
	}

	// ret := make([]Instance, 0)
	nextPageToken := ""
	for {
		resp, err := client.Disks.AggregatedList(project).MaxResults(pageSize).PageToken(nextPageToken).Do()
		if err != nil {
			return err
		}
		// rt := make([]Disk, len(resp.Items))
		// Items is a map with zone as key
		for _, scopedList := range resp.Items {
			for _, instance := range scopedList.Disks {
				fmt.Printf("disk: %+v\n", instance)
			}
		}
		// ret = append(ret, rt...)
		if resp.NextPageToken == "" {
			break
		}
		nextPageToken = resp.NextPageToken
	}
	return nil
}

func ListInstances(project string) error {
	ctx := context.Background()
	const pageSize int64 = 100
	client, err := compute.NewService(ctx)
	if err != nil {
		return err
	}

	// ret := make([]Instance, 0)
	nextPageToken := ""
	for {
		resp, err := client.Instances.AggregatedList(project).MaxResults(pageSize).PageToken(nextPageToken).Do()
		if err != nil {
			return err
		}
		// rt := make([]Instance, len(resp.Items))
		// Items is a map with zone as key
		for _, scopedList := range resp.Items {
			for _, instance := range scopedList.Instances {
				// Good fields: Id, Labels, MachineType, Name, Scheduling, Status
				fmt.Printf("instance: %+v\n", instance)
			}
		}
		// ret = append(ret, rt...)
		if resp.NextPageToken == "" {
			break
		}
		nextPageToken = resp.NextPageToken
	}
	return nil
}
