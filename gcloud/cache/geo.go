package cache

import (
	"nephomancy/common/geo"
)

func RegionsByContinent(continent geo.Continent) []string {
	switch continent {
	case geo.UnknownC:
		return []string{}
	case geo.Africa:
		return []string{}
	case geo.Asia:
		return []string{
			"asia-east1", "asia-east2", "asia-northeast1",
			"asia-northeast2", "asia-northeast3",
			"asia-south1", "asia-southeast1", "asia-southeast2",
		}
	case geo.Australia:
		return []string{"australia-southeast1"}
	case geo.Europe:
		return []string{
			"europe-north1", "europe-west1", "europe-west2",
			"europe-west3", "europe-west4", "europe-west6",
		}
	case geo.LatinAmerica:
		return []string{"southamerica-east1"}
	case geo.NorthAmerica:
		return []string{
			"northamerica-northeast1", "southamerica-east1",
			"us-central1", "us-east1", "us-east4",
			"us-west1", "us-west2", "us-west3", "us-west4",
		}
	default:
		return []string{}
	}
}

func RegionsByCountry(cc string) []string {
	switch cc {
	case "TW":
		return []string{"asia-east1"}
	case "HK":
		return []string{"asia-east2"}
	case "JP":
		return []string{"asia-northeast1", "asia-northeast2"}
	case "KR":
		return []string{"asia-northeast3"}
	case "IN":
		return []string{"asia-south1"}
	case "SG":
		return []string{"asia-southeast1"}
	case "ID":
		return []string{"asia-southeast2"}
	case "AU":
		return []string{"australia-southeast1"}
	case "FI":
		return []string{"europe-north1"}
	case "BE":
		return []string{"europe-west1"}
	case "UK":
		return []string{"europe-west2"}
	case "DE":
		return []string{"europe-west3"}
	case "NL":
		return []string{"europe-west4"}
	case "CH":
		return []string{"europe-west6"}
	case "CA":
		return []string{"northamerica-northeast1"}
	case "BR":
		return []string{"southamerica-east1"}
	case "US":
		return []string{
			"us-central1", "us-east1", "us-east4",
			"us-west1", "us-west2", "us-west3", "us-west4",
		}
	default:
		return []string{}
	}
}

// Takes a gcloud region name as input and returns an ISO
// 2-letter country code.
func RegionCountry(region string) (country string) {
	switch region {
	case "asia-east1":
		return "TW"
	case "asia-east2":
		return "HK"
	case "asia-northeast1":
		return "JP"
	case "asia-northeast2":
		return "JP"
	case "asia-northeast3":
		return "KR"
	case "asia-south1":
		return "IN"
	case "asia-southeast1":
		return "SG"
	case "asia-southeast2":
		return "ID"
	case "australia-southeast1":
		return "AU"
	case "europe-north1":
		return "FI"
	case "europe-west1":
		return "BE"
	case "europe-west2":
		return "UK"
	case "europe-west3":
		return "DE"
	case "europe-west4":
		return "NL"
	case "europe-west6":
		return "CH"
	case "northamerica-northeast1":
		return "CA"
	case "southamerica-east1":
		return "BR"
	case "us-central1":
		return "US" // Iowa
	case "us-east1":
		return "US" // South Carolina
	case "us-east4":
		return "US" // Virginia
	case "us-west1":
		return "US" // Oregon
	case "us-west2":
		return "US" // Los Angeles
	case "us-west3":
		return "US" // Salt Lake City
	case "us-west4":
		return "US" // Las Vegas
	default:
		return "Unknown"
	}
}
