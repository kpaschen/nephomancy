// Obtain information about the resource assets in use by a given
// project.
// This uses the asset api to get the assets being used; the information
// returned is more or less the same as what you get from the compute api,
// but you get all the used assets in one call rather than having to go
// over all the possible asset types.

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

	ret := make([]SmallAsset, 0)
	nextPageToken := ""
	for {
		resp, err := client.Assets.List(project).ContentType("RESOURCE").PageSize(pageSize).PageToken(nextPageToken).Do()
		if err != nil {
			return nil, err
		}
		rt := make([]SmallAsset, len(resp.Assets))
		for idx, a := range resp.Assets {
			by, berr := a.Resource.MarshalJSON()
			if berr != nil {
				return nil, berr
			}
			rt[idx] = SmallAsset{
				Name: a.Name,
				AssetType: a.AssetType,
				ResourceAsJson: string(by),
			}
		}
		ret = append(ret, rt...)
		if resp.NextPageToken == "" {
			break
		}
		nextPageToken = resp.NextPageToken
	}
	return ret, nil
}
