package ec2

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/go-test/deep"
	"nephomancy/aws/resources"
	"testing"
)

func TestMakeDto(t *testing.T) {
	var coreCount int64 = 1
	validCores := make([]*int64, 1)
	validCores[0] = &coreCount
	typeInfo := ec2.InstanceTypeInfo{
		InstanceType: aws.String("testType"),
		MemoryInfo: &ec2.MemoryInfo{
			SizeInMiB: aws.Int64(1000),
		},
		NetworkInfo: &ec2.NetworkInfo{
			NetworkPerformance: aws.String("4x 100 Gigabit"),
		},
		VCpuInfo: &ec2.VCpuInfo{
			DefaultVCpus: aws.Int64(1),
			ValidCores: validCores,
		},
		InstanceStorageSupported: aws.Bool(false),
		SupportedUsageClasses: []*string{ aws.String("on-demand"), },
	}
	it, err := MakeDto(typeInfo)
	if err != nil {
		t.Errorf("make dto failed: %v", err)
	}
	wantedDto := resources.InstanceType{
		Name: "testType",
		MemoryMiB: uint64(1000),
		NetworkPerformanceGbit: uint32(100),
		DefaultCpuCount: 1,
		ValidCores: []uint32{ uint32(1) },
		SupportedUsageClasses: []string{ "on-demand", },
	}
	diff := deep.Equal(wantedDto, *it)
	if diff != nil {
		t.Errorf("expected dto %+v but got %+v\ndiff: %+v",
		wantedDto, *it, diff)
	}
}
