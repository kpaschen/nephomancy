// Obtain list of instance types.

package ec2

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"nephomancy/aws/resources"
)

func DescribeInstanceTypes(instanceTypes []string, region string) (
	[]*resources.InstanceType, error) {
	if region == "" {
		region = "us-east-1"
	}
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(region),
	}))
	svc := ec2.New(sess)
	var pageSize int64 = 10
	ret := make([]*resources.InstanceType, 0)
	request := &ec2.DescribeInstanceTypesInput{
		MaxResults: aws.Int64(pageSize),
	}
	if instanceTypes != nil {
		request.InstanceTypes = []*string{}
		for _, it := range instanceTypes {
			request.InstanceTypes = append(request.InstanceTypes, aws.String(it))
		}
	}
	for {
		typeInfoList, err := svc.DescribeInstanceTypes(request)
		if err != nil {
			return nil, fmt.Errorf("could not get instance types at location %s\n", region)
		}
		instanceTypes := make([]*resources.InstanceType,
		                      len(typeInfoList.InstanceTypes))
		for idx, typeInfo := range typeInfoList.InstanceTypes {
			instanceTypes[idx], err = MakeDto(*typeInfo)
			if err != nil {
				return nil, fmt.Errorf(
					"could not convert instance type info %v to dto: %v",
					*typeInfo, err)
			}
		}
		ret = append(ret, instanceTypes...)
		if typeInfoList.NextToken == nil {
			break
		}
		request.NextToken = typeInfoList.NextToken
	}

	return ret, nil
}

func ListInstanceTypesByLocation(region string) ([]string, error) {
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(region),
	}))
	svc := ec2.New(sess)
	var pageSize int64 = 100
	request := &ec2.DescribeInstanceTypeOfferingsInput{
		// AvailabilityZoneId (not AvailabilityZone ... they get remapped)
		// would be a more reliable choice here.
		LocationType: aws.String("region"),
		Filters: []*ec2.Filter{
			{
				Name: aws.String("location"),
				Values: aws.StringSlice([]string{ region }),
			},
		},
		MaxResults: aws.Int64(pageSize),
	}
	ret := make([]string, 0)
	for {
		offeringsObj, err := svc.DescribeInstanceTypeOfferings(request)
		if err != nil {
			return nil, fmt.Errorf("could not get offerings by location: %v", err)
		}
		ofs := make([]string, len(offeringsObj.InstanceTypeOfferings))
		for idx, offering := range offeringsObj.InstanceTypeOfferings {
			ofs[idx] = *offering.InstanceType
		}
		ret = append(ret, ofs...)
		if offeringsObj.NextToken == nil {
			break
		}
		request.NextToken = offeringsObj.NextToken
	}
	return ret, nil
}

func MakeDto(typeInfo ec2.InstanceTypeInfo) (*resources.InstanceType, error) {
        var mem uint64 = 0
        if typeInfo.MemoryInfo != nil {
                mem = uint64(*typeInfo.MemoryInfo.SizeInMiB)
        } else {
                return nil, fmt.Errorf("no memory info found in instance type %+v", typeInfo)
        }
	var np uint32 = 0
	if typeInfo.NetworkInfo != nil {
		perfString := *typeInfo.NetworkInfo.NetworkPerformance
		switch perfString {
		// These numbers are based on some stackoverflow research.
		case "Very Low": np = 1
		case "Low": np = 1
		case "Moderate": np = 5
		case "Low to Moderate": np = 1
		case "High": np = 10
		default:
			// Should maybe divide np by 5 or so if the spec says "Up to". I'm pretty
			// sure "Up to" is some kind of burst/peak performance only.
			count, err := fmt.Sscanf(perfString, "Up to %d Gigabit", &np)
			if count != 1 || err != nil {
				count, err = fmt.Sscanf(perfString, "%d Gigabit", &np)
				if count != 1 || err != nil {
					count, err = fmt.Sscanf(perfString, "%*s %d Gigabit", &np)
					if count != 2 || err != nil {
						fmt.Printf(
							"failed to parse network performance out of spec: %s\n",
							perfString)
					}
				}
			}
		}
	} else {
		return nil, fmt.Errorf("network info missing on instance type %+v", typeInfo)
	}
	var usageClasses []string
	if typeInfo.SupportedUsageClasses != nil {
		usageClasses = make([]string, len(typeInfo.SupportedUsageClasses))
		for idx, uc := range typeInfo.SupportedUsageClasses {
			usageClasses[idx] = *uc
		}
	} else {
		usageClasses = make([]string, 0)
	}
	cpuInfo := typeInfo.VCpuInfo
	var cpuCount uint32 = uint32(*cpuInfo.DefaultVCpus)
	validCores := make([]uint32, len(cpuInfo.ValidCores))
	for idx, vc := range cpuInfo.ValidCores {
		validCores[idx] = uint32(*vc)
	}
	var gpuCount uint32 = 0
	if typeInfo.GpuInfo != nil {
		gpus := typeInfo.GpuInfo.Gpus
		for _, gpu := range gpus {
			gpuCount += uint32(*gpu.Count)
		}
	}
	var instanceStorage uint64 = 0
	var instanceStorageType string
	if *typeInfo.InstanceStorageSupported && typeInfo.InstanceStorageInfo != nil {
		instanceStorage = uint64(*typeInfo.InstanceStorageInfo.TotalSizeInGB)
		instanceStorageType = *typeInfo.InstanceStorageInfo.Disks[0].Type
	}

        return &resources.InstanceType{
                Name: *typeInfo.InstanceType,
                MemoryMiB: mem,
		NetworkPerformanceGbit: np,
		SupportedUsageClasses: usageClasses,
		DefaultCpuCount: cpuCount,
		ValidCores: validCores,
		InstanceStorageSupported: *typeInfo.InstanceStorageSupported,
		InstanceStorageMaxSizeGb: instanceStorage,
		InstanceStorageType: instanceStorageType,
		GpuCount: gpuCount,
        }, nil
}
