package eliona

import (
	"fmt"
	api "github.com/eliona-smart-building-assistant/go-eliona-api-client/v2"
	"github.com/eliona-smart-building-assistant/go-eliona/asset"
)

func UpsertData(assetId int32, data any) error {
	subtypes := asset.SplitBySubtype(data)
	for subtype, data := range subtypes {
		if subtype != "" {
			if err := asset.UpsertData(api.Data{
				AssetId: assetId,
				Subtype: subtype,
				Data:    data,
			}); err != nil {
				return fmt.Errorf("upserting data for subtype %s: %w", subtype, err)
			}
		}
	}
	return nil
}
