package command

import (
	"fmt"
	"log"
	"nephomancy/aws/ec2"
	"nephomancy/common/command"
)

type ListCommand struct {
	command.Command

	region string
}

func (*ListCommand) Help() string {
	return "list some stuff"
}

func (*ListCommand) Synopsis() string {
	return "lists some stuff"
}

func (r *ListCommand) Run(args []string) int {
	fs := r.Command.DefaultFlagSet("awsRegions")
	fs.StringVar(&r.region, "region", "us-east-1", "The region to list things for.")
	fs.Parse(args)

	itypes, err := ec2.ListInstanceTypesByLocation(r.region)
	if err != nil {
		log.Fatalf("Failed to list aws instance type offerings: %v\n", err)
	}
	fmt.Printf("instance types in region %s: %+v\n", r.region, itypes)

	instanceTypes, err := ec2.DescribeInstanceTypes(nil, r.region)
	if err != nil {
		log.Fatalf("Failed to describe aws instance type offerings: %v\n", err)
	}
	for _, it := range instanceTypes {
		fmt.Printf("description: %+v\n", *it)
	}
	return 0
}
