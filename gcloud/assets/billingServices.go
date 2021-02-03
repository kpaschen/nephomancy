package assets

import (
	"context"
	"fmt"
	"strings"
	"time"

	// this uses the old-style go apis.
	// The new one is documented here
	// https://pkg.go.dev/cloud.google.com/go@v0.73.0/billing/apiv1
	"google.golang.org/api/cloudbilling/v1"
)

type BillingService struct {
	BusinessEntityName string
	DisplayName string
	Name string
	ServiceID string
}

type skuCategory struct {
	ResourceFamily string
	ResourceGroup string
        ServiceDisplayName string
	UsageType string
}

type PricingInfo struct {
	EffectiveFrom int64
	Summary string
	CurrencyConversionRate float64
	PricingExpression string
        AggregationInfo string
}

type BillingServiceSku struct {
	Category *skuCategory
	Description string
	// TYPE_UNSPECIFIED, GLOBAL, REGIONAL or MULTI_REGIONAL
	GeoTaxonomyType string
	GeoTaxonomyRegions string
	Name string
	PricingInfo []PricingInfo
	ServiceProviderName string
	ServiceRegions []string
	SkuId string
}

func ListBillingServices() (map[string]BillingService, error) {
	ctx := context.Background()
	client, err := cloudbilling.NewService(ctx)
	if err != nil {
		return nil, err
	}
	ret := make(map[string]BillingService)
	for {
		nextPageToken := ""
		resp, err := client.Services.List().PageToken(nextPageToken).Do()
		if err != nil {
			return nil, err
		}
		if resp.ServerResponse.HTTPStatusCode != 200 {
			err := fmt.Errorf("cloudbilling.services.list returned error response: %v", resp.ServerResponse)
			return nil, err
		}
		for _, s := range resp.Services {
			ret[s.DisplayName] = BillingService{
				BusinessEntityName: s.BusinessEntityName,
				DisplayName: s.DisplayName,
				Name: s.Name,
				ServiceID: s.ServiceId,
			}
		}
		if resp.NextPageToken == "" {
			break
		}
		nextPageToken = resp.NextPageToken
	}
	return ret, nil
}

// Might want to pass a time-since here so as to get only things changed
// since last request? And then cache information somewhere. Otherwise
// there are a lot of skus to retrieve.
// Some services (e.g. compute engine) have a lot of skus, might be worth
// looking first which regions the current project is even present in and
// then skip skus that are regional and don't overlap those regions/zones.
func ListSkus(billingServiceName *string) (map[string]BillingServiceSku, error) {
	ctx := context.Background()
	client, err := cloudbilling.NewService(ctx)
	if err != nil {
		return nil, err
	}
	ret := make(map[string]BillingServiceSku)
	// This is the magic go timestamp parsing string.
	const tsLayout = "2006-01-02T15:04:05.000Z"
	nextPageToken := ""
	for {
		resp, err := client.Services.Skus.List(*billingServiceName).PageSize(100).PageToken(nextPageToken).Do()
		if err != nil {
			return nil, err
		}
		if resp.ServerResponse.HTTPStatusCode != 200 {
			err := fmt.Errorf("cloudbilling.services.skus.list returned error response: %v", resp.ServerResponse)
			return nil, err
		}
		for _, s := range resp.Skus {
			cat := skuCategory{
				ResourceFamily: s.Category.ResourceFamily,
				ResourceGroup: s.Category.ResourceGroup,
				ServiceDisplayName: s.Category.ServiceDisplayName,
				UsageType: s.Category.UsageType,
			}
			gtType := ""
			gtRegions := ""
			if s.GeoTaxonomy != nil {
				gtType = s.GeoTaxonomy.Type
				gtRegions = strings.Join(s.GeoTaxonomy.Regions, ", ")
			}
			pInfo := make([]PricingInfo, len(s.PricingInfo))
			for idx, p := range s.PricingInfo {
				pBytes, err := p.PricingExpression.MarshalJSON()
				if err != nil {
					return nil, err
				}
				aggregationInfo := ""
				if p.AggregationInfo != nil {
					aBytes, err := p.AggregationInfo.MarshalJSON()
					if err != nil {
						return nil, err
					}
					aggregationInfo = string(aBytes)
				}
				ts, err := time.Parse(tsLayout, p.EffectiveTime)
				if err != nil {
					return nil, err
				}
				timestamp := ts.Unix()
				pInfo[idx] = PricingInfo{
					EffectiveFrom: timestamp,
					Summary: p.Summary,
					PricingExpression: string(pBytes),
					AggregationInfo: aggregationInfo,
					CurrencyConversionRate: p.CurrencyConversionRate,
				}
			}
			ret[s.Name] = BillingServiceSku{
				Category: &cat,
				Description: s.Description,
				GeoTaxonomyType: gtType,
				GeoTaxonomyRegions: gtRegions,
				Name: s.Name,
				ServiceProviderName: s.ServiceProviderName,
				SkuId: s.SkuId,
				PricingInfo: pInfo,
				ServiceRegions: s.ServiceRegions,
			}
		}
		if resp.NextPageToken == "" {
			break
		}
		nextPageToken = resp.NextPageToken
	}
	return ret, nil
}
