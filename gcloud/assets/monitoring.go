package assets

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/monitoring/apiv3"
	//	metricpb "google.golang.org/genproto/googleapis/api/metric"
	timestamppb "github.com/golang/protobuf/ptypes/timestamp"
	"google.golang.org/api/iterator"
	monitoringpb "google.golang.org/genproto/googleapis/monitoring/v3"
)

// code is here: https://godoc.org/cloud.google.com/go/monitoring/apiv3#pkg-files

// e.g. metricType=networking.googleapis.com/vm_flow/egress_bytes_count
// resource.type=gce_instance
// Probably no need for this, can just get the timeseries directly.
func getMetricDescriptors(project string, metricType string, resourceType string) error {
	ctx := context.Background()
	client, err := monitoring.NewMetricClient(ctx)
	if err != nil {
		return err
	}
	nextPageToken := ""
	// filter := fmt.Sprintf(`metric.type=starts_with("%s")`, "networking.googleapis.com")
	filter := fmt.Sprintf(`metric.type=starts_with("%s") AND resource.type=starts_with("%s")`,
		metricType, resourceType)
	for {
		req := &monitoringpb.ListMetricDescriptorsRequest{
			Name:      project,
			Filter:    filter,
			PageSize:  100,
			PageToken: nextPageToken,
		}
		it := client.ListMetricDescriptors(ctx, req)
		for {
			resp, err := it.Next()
			if err == iterator.Done {
				break
			}
			if err != nil {
				return err
			}
			fmt.Printf("resp: %+v\n", resp)
		}
		if it.PageInfo().Token == "" {
			break
		}
		nextPageToken = it.PageInfo().Token
	}
	return nil
}

// This gets the most recent value of the given time series, averaged
// over all instances.
// TODO: finish debugging this once there is monitoring data.
func getTimeSeries(project string, metricType string, resourceType string) error {
	ctx := context.Background()
	client, err := monitoring.NewMetricClient(ctx)
	if err != nil {
		return err
	}
	now := time.Now().Unix()
	interval := &monitoringpb.TimeInterval{
		EndTime: &timestamppb.Timestamp{
			Seconds: now,
		},
		StartTime: &timestamppb.Timestamp{
			Seconds: now - 60,
		},
	}
	filter := fmt.Sprintf(`metric.type=starts_with("%s") AND resource.type=starts_with("%s")`,
		metricType, resourceType)
	req := &monitoringpb.ListTimeSeriesRequest{
		Name:     project,
		Filter:   filter,
		Interval: interval,
		Aggregation: &monitoringpb.Aggregation{
			CrossSeriesReducer: monitoringpb.Aggregation_REDUCE_SUM,
		},
	}
	it := client.ListTimeSeries(ctx, req)
	for {
		resp, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return err
		}
		fmt.Printf("ts: %+v\n", resp)
	}
	return nil
}

func ListMetrics(project string) error {
	err := getMetricDescriptors(project,
		"networking.googleapis.com/vm_flow/egress_bytes_count", "gce_instance")
	if err != nil {
		return err
	}
	err = getTimeSeries(project,
		"networking.googleapis.com/vm_flow/egress_bytes_count", "gce_instance")
	return nil
}
