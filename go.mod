module nephomancy

go 1.15

require (
	github.com/gobwas/glob v0.2.3 // indirect
	github.com/gogo/protobuf v1.3.1 // indirect
	github.com/grpc-ecosystem/grpc-gateway v1.16.0 // indirect
	github.com/iancoleman/strcase v0.1.2 // indirect
	github.com/lyft/protoc-gen-star v0.5.2 // indirect
	github.com/mitchellh/cli v1.1.2
	github.com/nbutton23/zxcvbn-go v0.0.0-20201221231540-e56b841a3c88 // indirect
	nephomancy/gcloud/assets v0.0.0-00010101000000-000000000000
	nephomancy/gcloud/cache v0.0.0-00010101000000-000000000000
	nephomancy/gcloud/command v0.0.0-00010101000000-000000000000
)

replace nephomancy/gcloud/assets => ./gcloud/assets

replace nephomancy/gcloud/cache => ./gcloud/cache

replace nephomancy/gcloud/pricing => ./gcloud/pricing

replace nephomancy/gcloud/command => ./gcloud/command
