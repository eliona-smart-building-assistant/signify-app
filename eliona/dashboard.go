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
	"context"
	api "github.com/eliona-smart-building-assistant/go-eliona-api-client/v2"
	"github.com/eliona-smart-building-assistant/go-utils/common"
	"github.com/volatiletech/null/v8"
	"signify/appdb"
	"signify/conf"
)

func SignifyPeopleCountDashboard(projectId string) (api.Dashboard, error) {
	dashboard := api.Dashboard{}
	dashboard.Name = "Signify People Count"
	dashboard.ProjectId = projectId
	dashboard.Widgets = []api.Widget{}

	storeys, err := conf.GetAssets(context.Background(),
		appdb.AssetWhere.Kind.EQ(string(conf.StoreyAssetKind)),
		appdb.AssetWhere.ProjectID.EQ(projectId),
	)
	if err != nil {
		return api.Dashboard{}, err
	}
	for _, storey := range storeys {

		apiStorey, err := getAssetById(storey.AssetID.Int32)
		if err != nil {
			return api.Dashboard{}, err
		}

		if apiStorey == nil {
			continue
		}

		widget := api.Widget{
			WidgetTypeName: "GeneralDisplay",
			AssetId:        apiStorey.Id,
			Sequence:       nullableInt32(int32(storey.ID)),
			Details: map[string]any{
				"size":     4,
				"timespan": 7,
			},
			Data: []api.WidgetData{},
		}

		spaces, err := conf.GetAssets(context.Background(),
			appdb.AssetWhere.ParentUUID.EQ(null.StringFrom(storey.UUID)),
			appdb.AssetWhere.Kind.EQ(string(conf.SpaceAssetKind)),
			appdb.AssetWhere.ProjectID.EQ(projectId),
		)
		if err != nil {
			return api.Dashboard{}, err
		}

		for _, space := range spaces {

			apiSpace, err := getAssetById(space.AssetID.Int32)
			if err != nil {
				return api.Dashboard{}, err
			}

			if apiSpace == nil || apiSpace.AssetType != PeopleCountAssetType {
				continue
			}

			widgetData := api.WidgetData{
				ElementSequence: nullableInt32(1),
				AssetId:         apiSpace.Id,
				Data: map[string]interface{}{
					"aggregatedDataField": nil,
					"aggregatedDataType":  "heap",
					"attribute":           "people_count",
					"description":         apiSpace.Name.Get(),
					"key":                 "",
					"seq":                 nullableInt32(int32(space.ID)),
					"subtype":             "input",
				},
			}

			widget.Data = append(widget.Data, widgetData)

		}

		// add station widget to dashboard
		if len(widget.Data) > 0 {
			dashboard.Widgets = append(dashboard.Widgets, widget)
		}
	}
	return dashboard, nil
}

func SignifyOccupancyDashboard(projectId string) (api.Dashboard, error) {
	dashboard := api.Dashboard{}
	dashboard.Name = "Signify Occupancy"
	dashboard.ProjectId = projectId
	dashboard.Widgets = []api.Widget{}

	storeys, err := conf.GetAssets(context.Background(),
		appdb.AssetWhere.Kind.EQ(string(conf.StoreyAssetKind)),
		appdb.AssetWhere.ProjectID.EQ(projectId),
	)
	if err != nil {
		return api.Dashboard{}, err
	}
	for _, storey := range storeys {

		apiStorey, err := getAssetById(storey.AssetID.Int32)
		if err != nil {
			return api.Dashboard{}, err
		}

		if apiStorey == nil {
			continue
		}

		widget := api.Widget{
			WidgetTypeName: "GeneralDisplay",
			AssetId:        apiStorey.Id,
			Sequence:       nullableInt32(int32(storey.ID)),
			Details: map[string]any{
				"size":     4,
				"timespan": 7,
			},
			Data: []api.WidgetData{},
		}

		spaces, err := conf.GetAssets(context.Background(),
			appdb.AssetWhere.ParentUUID.EQ(null.StringFrom(storey.UUID)),
			appdb.AssetWhere.Kind.EQ(string(conf.SpaceAssetKind)),
			appdb.AssetWhere.ProjectID.EQ(projectId),
		)
		if err != nil {
			return api.Dashboard{}, err
		}

		for _, space := range spaces {

			apiSpace, err := getAssetById(space.AssetID.Int32)
			if err != nil {
				return api.Dashboard{}, err
			}

			if apiSpace == nil || apiSpace.AssetType != OccupancyAssetType {
				continue
			}

			widgetData := api.WidgetData{
				ElementSequence: nullableInt32(1),
				AssetId:         apiSpace.Id,
				Data: map[string]interface{}{
					"aggregatedDataField": nil,
					"aggregatedDataType":  "heap",
					"attribute":           "occupancy",
					"description":         apiSpace.Name.Get(),
					"key":                 "",
					"seq":                 nullableInt32(int32(space.ID)),
					"subtype":             "input",
				},
			}

			widget.Data = append(widget.Data, widgetData)

		}

		// add station widget to dashboard
		if len(widget.Data) > 0 {
			dashboard.Widgets = append(dashboard.Widgets, widget)
		}
	}
	return dashboard, nil
}

func SignifyTemperatureDashboard(projectId string) (api.Dashboard, error) {
	dashboard := api.Dashboard{}
	dashboard.Name = "Signify Temperature"
	dashboard.ProjectId = projectId
	dashboard.Widgets = []api.Widget{}

	storeys, err := conf.GetAssets(context.Background(),
		appdb.AssetWhere.Kind.EQ(string(conf.StoreyAssetKind)),
		appdb.AssetWhere.ProjectID.EQ(projectId),
	)
	if err != nil {
		return api.Dashboard{}, err
	}
	for _, storey := range storeys {

		apiStorey, err := getAssetById(storey.AssetID.Int32)
		if err != nil {
			return api.Dashboard{}, err
		}

		if apiStorey == nil {
			continue
		}

		widget := api.Widget{
			WidgetTypeName: "GeneralDisplay",
			AssetId:        apiStorey.Id,
			Sequence:       nullableInt32(int32(storey.ID)),
			Details: map[string]any{
				"size":     4,
				"timespan": 7,
			},
			Data: []api.WidgetData{},
		}

		spaces, err := conf.GetAssets(context.Background(),
			appdb.AssetWhere.ParentUUID.EQ(null.StringFrom(storey.UUID)),
			appdb.AssetWhere.Kind.EQ(string(conf.SpaceAssetKind)),
			appdb.AssetWhere.ProjectID.EQ(projectId),
		)
		if err != nil {
			return api.Dashboard{}, err
		}

		for _, space := range spaces {

			apiSpace, err := getAssetById(space.AssetID.Int32)
			if err != nil {
				return api.Dashboard{}, err
			}

			if apiSpace == nil || apiSpace.AssetType != TemperatureAssetType {
				continue
			}

			widgetData := api.WidgetData{
				ElementSequence: nullableInt32(1),
				AssetId:         apiSpace.Id,
				Data: map[string]interface{}{
					"aggregatedDataField": nil,
					"aggregatedDataType":  "heap",
					"attribute":           "temperature",
					"description":         apiSpace.Name.Get(),
					"key":                 "",
					"seq":                 nullableInt32(int32(space.ID)),
					"subtype":             "input",
				},
			}

			widget.Data = append(widget.Data, widgetData)

		}

		// add station widget to dashboard
		if len(widget.Data) > 0 {
			dashboard.Widgets = append(dashboard.Widgets, widget)
		}
	}
	return dashboard, nil
}

func SignifyHumidityDashboard(projectId string) (api.Dashboard, error) {
	dashboard := api.Dashboard{}
	dashboard.Name = "Signify Humidity"
	dashboard.ProjectId = projectId
	dashboard.Widgets = []api.Widget{}

	storeys, err := conf.GetAssets(context.Background(),
		appdb.AssetWhere.Kind.EQ(string(conf.StoreyAssetKind)),
		appdb.AssetWhere.ProjectID.EQ(projectId),
	)
	if err != nil {
		return api.Dashboard{}, err
	}
	for _, storey := range storeys {

		apiStorey, err := getAssetById(storey.AssetID.Int32)
		if err != nil {
			return api.Dashboard{}, err
		}

		if apiStorey == nil {
			continue
		}

		widget := api.Widget{
			WidgetTypeName: "GeneralDisplay",
			AssetId:        apiStorey.Id,
			Sequence:       nullableInt32(int32(storey.ID)),
			Details: map[string]any{
				"size":     4,
				"timespan": 7,
			},
			Data: []api.WidgetData{},
		}

		spaces, err := conf.GetAssets(context.Background(),
			appdb.AssetWhere.ParentUUID.EQ(null.StringFrom(storey.UUID)),
			appdb.AssetWhere.Kind.EQ(string(conf.SpaceAssetKind)),
			appdb.AssetWhere.ProjectID.EQ(projectId),
		)
		if err != nil {
			return api.Dashboard{}, err
		}

		for _, space := range spaces {

			apiSpace, err := getAssetById(space.AssetID.Int32)
			if err != nil {
				return api.Dashboard{}, err
			}

			if apiSpace == nil || apiSpace.AssetType != HumidityAssetType {
				continue
			}

			widgetData := api.WidgetData{
				ElementSequence: nullableInt32(1),
				AssetId:         apiSpace.Id,
				Data: map[string]interface{}{
					"aggregatedDataField": nil,
					"aggregatedDataType":  "heap",
					"attribute":           "humidity",
					"description":         apiSpace.Name.Get(),
					"key":                 "",
					"seq":                 nullableInt32(int32(space.ID)),
					"subtype":             "input",
				},
			}

			widget.Data = append(widget.Data, widgetData)

		}

		// add station widget to dashboard
		if len(widget.Data) > 0 {
			dashboard.Widgets = append(dashboard.Widgets, widget)
		}
	}
	return dashboard, nil
}

func nullableInt32(val int32) api.NullableInt32 {
	return *api.NewNullableInt32(common.Ptr[int32](val))
}
