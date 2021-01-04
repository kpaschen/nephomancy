package assets

import (
	"context"

	"google.golang.org/api/cloudasset/v1p5beta1"
)

// Remember to export GOOGLE_APPLICATION_CREDENTIALS=<wherever.json>
// also make sure cloud asset api is enabled for your project
// make sure the project you pass in the parent is the one that the application credentials
// grant access to
// you need the cloud-platform auth scope or an equivalent role.
func ListAssetsForProject(project string) ([]SmallAsset, error) {
	ctx := context.Background()
	var pageSize int64 = 100
	client, err := cloudasset.NewService(ctx)
	if err != nil {
		return nil, err
	}

	ret := make([]SmallAsset, pageSize)
	for {
		nextPageToken := ""
		// TODO: maybe set asset types here, but the new api does not support that.
		resp, err := client.Assets.List(project).ContentType("RESOURCE").PageSize(pageSize).PageToken(nextPageToken).Do()
		if err != nil {
			return nil, err
		}
		// TODO: maybe use etags and ifNotModified here to only get diffs
		for _, a := range resp.Assets {
			by, berr := a.Resource.MarshalJSON()
			if berr != nil {
				return nil, berr
			}
			ret = append(ret, SmallAsset{
				Name: a.Name,
				AssetType: a.AssetType,
				ResourceAsJson: string(by),
			})
		}
		if resp.NextPageToken == "" {
			break
		}
		nextPageToken = resp.NextPageToken
	}
	return ret, nil
}
