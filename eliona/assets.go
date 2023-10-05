//  This file is part of the eliona project.
//  Copyright © 2022 LEICOM iTEC AG. All Rights Reserved.
//  ______ _ _
// |  ____| (_)
// | |__  | |_  ___  _ __   __ _
// |  __| | | |/ _ \| '_ \ / _` |
// | |____| | | (_) | | | | (_| |
// |______|_|_|\___/|_| |_|\__,_|
//
//  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING
//  BUT NOT LIMITED  TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
//  NON INFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM,
//  DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
//  OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package eliona

import (
	"fmt"
	api "github.com/eliona-smart-building-assistant/go-eliona-api-client/v2"
	"github.com/eliona-smart-building-assistant/go-eliona/asset"
	"github.com/eliona-smart-building-assistant/go-eliona/utils"
	"github.com/eliona-smart-building-assistant/go-utils/common"
	"signify/apiserver"
)

const (
	HumidityAssetType    = "signify_humidity_space"
	OccupancyAssetType   = "signify_occupancy_space"
	TemperatureAssetType = "signify_temperature_space"
	GroupAssetType       = "signify_group"
	RootAssetType        = "signify_root"
)

type Asset interface {
	AssetType() string
	Id() string
}

func UpsertAsset(projectId string, uniqueIdentifier string, parentId *int32, assetType string, name string) (*int32, error) {
	assetId, err := asset.UpsertAsset(api.Asset{
		ProjectId:               projectId,
		GlobalAssetIdentifier:   uniqueIdentifier,
		Name:                    *api.NewNullableString(common.Ptr(name)),
		AssetType:               assetType,
		Description:             *api.NewNullableString(common.Ptr(fmt.Sprintf("%s (%v)", name, uniqueIdentifier))),
		ParentLocationalAssetId: *api.NewNullableInt32(parentId),
		DeviceIds: []string{
			uniqueIdentifier,
		},
	})
	if err != nil {
		return nil, err
	}
	return assetId, nil
}

func AdheresToFilter(input interface{}, filter [][]apiserver.FilterRule) (bool, error) {
	f := apiFilterToCommonFilter(filter)
	fp, err := utils.StructToMap(input)
	if err != nil {
		return false, fmt.Errorf("converting strict to map: %v", err)
	}
	adheres, err := common.Filter(f, fp)
	if err != nil {
		return false, err
	}
	return adheres, nil
}

func apiFilterToCommonFilter(input [][]apiserver.FilterRule) [][]common.FilterRule {
	result := make([][]common.FilterRule, len(input))
	for i := 0; i < len(input); i++ {
		result[i] = make([]common.FilterRule, len(input[i]))
		for j := 0; j < len(input[i]); j++ {
			result[i][j] = common.FilterRule{
				Parameter: input[i][j].Parameter,
				Regex:     input[i][j].Regex,
			}
		}
	}
	return result
}
