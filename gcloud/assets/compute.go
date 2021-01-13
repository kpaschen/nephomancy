package assets

import (
	"context"
	"fmt"
	"strings"
	"google.golang.org/api/compute/v1"
)

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
					Zone: t.Name,
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
			rt[idx] = MachineType{
				CpuCount: t.GuestCpus,
				IsSharedCpu: shared,
				MemoryMb: t.MemoryMb,
				Name: t.Name,
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
				Name: t.Name,
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
				Name: t.Name,
				DefaultSizeGb: t.DefaultDiskSizeGb,
				Region: region,
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
		qu, err  := q.MarshalJSON()
		if err != nil {
			return err
		}
		fmt.Printf("quota: %+v\n", string(qu))
	}
	return nil
}
