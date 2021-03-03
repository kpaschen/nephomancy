package resources

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/lightsail"
)

func ListRegions() ([]string, error) {
	sess := session.Must(session.NewSession(
		&aws.Config{
			Region: aws.String("us-east-1"),
		}))
	svc := ec2.New(sess)
	// This actually just gives you ids and endpoints.
	regions, err := svc.DescribeRegions(nil)
	if err != nil {
		return nil, err
	}
	fmt.Println("regions: %+v\n", regions)
	// Lightsail has pieces of the display names, but not for all regions
	// (e.g. govcloud is missing).
	ls := lightsail.New(sess)
	lr, err := ls.GetRegions(&lightsail.GetRegionsInput{
		IncludeAvailabilityZones: aws.Bool(true),
	})
	if err != nil {
		return nil, err
	}
	fmt.Println("lightsail regions: %+v\n", lr)
	// Or one can clone the aws go sdk and browse around ...
	resolvers := endpoints.DefaultPartitions()
	for _, partition := range resolvers {
		// The region descriptions here are pretty close to the location
		// values in the pricing data, but not the same. E.g. the pricing
		// data uses "EU", but the sdk uses "Europe".
		fmt.Printf("resolver %s: %+v\n", partition.ID(), partition.Regions())
	}

	// Alternatively, could use the systems manager:
	// aws ssm get-parameter --name /aws/service/global-infrastructure/regions/us-east-1/longName --query "Parameter.Value" --output text

	return nil, nil
}
