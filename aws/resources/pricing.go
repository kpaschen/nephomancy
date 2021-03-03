package resources

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/pricing"
)

func ListServices() ([]string, error) {
	// pricing endpoints only exist in us-east-1 and ap-south-1
	sess := session.Must(session.NewSession(
		&aws.Config{
			Region: aws.String("us-east-1"),
		}))
	svc := pricing.New(sess)
	services, err := svc.DescribeServices(&pricing.DescribeServicesInput{
		ServiceCode: aws.String("AmazonEC2"),
	})
	if err != nil {
		return nil, err
	}
	fmt.Println("services: %+v\n", services)
	// The products are actually the resource types.
	products, err := svc.GetProducts(&pricing.GetProductsInput{
		ServiceCode: aws.String("AmazonEC2"),
	})
	if err != nil {
		return nil, err
	}
	fmt.Println("products: %+v\n", products)

	values, err := svc.GetAttributeValues(&pricing.GetAttributeValuesInput{
		ServiceCode:   aws.String("AmazonEC2"),
		AttributeName: aws.String("location"),
	})
	if err != nil {
		return nil, err
	}
	fmt.Println("values for location: %+v\n", values)
	return nil, nil
}
