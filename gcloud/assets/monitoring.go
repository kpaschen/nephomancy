package assets

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/monitoring/apiv3"
	"google.golang.org/protobuf/types/known/durationpb"
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
	filter := fmt.Sprintf(`metric.type="%s" AND resource.type="%s"`,
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
			Seconds: now - 86400 * 7,
		},
	}
	filter := fmt.Sprintf(`metric.type="%s" AND resource.type="%s"`,
		metricType, resourceType)
	groupBy := "metric.labels.instance_name"
	_ = groupBy
	req := &monitoringpb.ListTimeSeriesRequest{
		Name:     project,
		Filter:   filter,
		Interval: interval,
		View:     monitoringpb.ListTimeSeriesRequest_FULL,
		Aggregation: &monitoringpb.Aggregation{
			CrossSeriesReducer: monitoringpb.Aggregation_REDUCE_SUM,
			PerSeriesAligner:   monitoringpb.Aggregation_ALIGN_MEAN,
			AlignmentPeriod:    &durationpb.Duration{Seconds: 600},
			GroupByFields:      []string{groupBy},
		},
	}
	fmt.Printf("sending ts request: %+v\n", req)
	it := client.ListTimeSeries(ctx, req)
	for {
		resp, err := it.Next()
		if err == iterator.Done {
			fmt.Printf("done going over time series\n")
			break
		}
		if err != nil {
			fmt.Printf("error: %v\n", err)
			return err
		}
		fmt.Printf("ts: %+v\n", resp)
	}
	return nil
}

func ListMetrics(project string) error {
	err := getMetricDescriptors(project,
		// "networking.googleapis.com/vm_flow/egress_bytes_count", "gce_instance")
		"compute.googleapis.com/instance/uptime", "gce_instance")
	if err != nil {
		return err
	}
	err = getTimeSeries(project,
		"compute.googleapis.com/instance/uptime_total", "gce_instance")
	return nil
}
