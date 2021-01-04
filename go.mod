module nephomancy

go 1.15

require (
	github.com/mitchellh/cli v1.1.2
	nephomancy/gcloud/assets v0.0.0-00010101000000-000000000000
	nephomancy/gcloud/cache v0.0.0-00010101000000-000000000000
	nephomancy/gcloud/command v0.0.0-00010101000000-000000000000
)

replace nephomancy/gcloud/assets => ./gcloud/assets

replace nephomancy/gcloud/cache => ./gcloud/cache
replace nephomancy/gcloud/pricing => ./gcloud/pricing

replace nephomancy/gcloud/command => ./gcloud/command
