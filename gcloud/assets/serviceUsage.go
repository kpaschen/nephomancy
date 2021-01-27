package assets

import (
	"context"
	"fmt"
	"google.golang.org/api/serviceusage/v1"
)

// code is here: https://code.googlesource.com/google-api-go-client/+/master/serviceusage/v1/serviceusage-gen.go

// This api has two services: Operations and Services
// type GoogleapiServiceusageV1ServiceConfig is a struct that has
// a *Quota field
// service config has list of metrics. quota.limits defines limits on the metrics.
// type QuotaLimit
// there are overrides for quota limits

// ListServices method. returns ListServicesResponse with
// Services []*GoogleApiServiceusageV1Service
// Get method takes a name, returns one *GoogleApiServiceusageV1Service

func ListServices(project string) error {
	ctx := context.Background()
	const pageSize int64 = 100
	client, err := serviceusage.NewService(ctx)
	if err != nil {
		return err
	}
	nextPageToken := ""
	for {
		resp, err := client.Services.List(project).PageSize(pageSize).PageToken(nextPageToken).Do()
		if err != nil {
			return err
		}
		for _, svc := range resp.Services {
			if svc.State == "ENABLED" {
				fmt.Printf("service: %s\n", svc.Name)
				config, err := svc.Config.MarshalJSON()
				if err != nil {
					return err
				}
				// the config has some quotas and the names of the metrics
				// used to enforce them.
				fmt.Printf("config: %+v\n", string(config))
			}
		}
		if resp.NextPageToken == "" {
			break
		}
		nextPageToken = resp.NextPageToken
	}
	return nil
}
