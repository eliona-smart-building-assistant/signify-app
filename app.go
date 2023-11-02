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

package main

import (
	"context"
	"fmt"
	"github.com/eliona-smart-building-assistant/go-eliona/frontend"
	"github.com/volatiletech/null/v8"
	"net/http"
	"signify/apiserver"
	"signify/apiservices"
	"signify/appdb"
	"signify/conf"
	"signify/eliona"
	"signify/signify"
	"sync"
	"time"

	"github.com/eliona-smart-building-assistant/go-utils/common"
	utilshttp "github.com/eliona-smart-building-assistant/go-utils/http"
	"github.com/eliona-smart-building-assistant/go-utils/log"
)

func collectAssets() {
	configs, err := conf.GetConfigs(context.Background())
	if err != nil {
		log.Fatal("conf", "Couldn't read configs from DB: %v", err)
		return
	}
	if len(configs) == 0 {
		return
	}

	for _, config := range configs {

		// Skip config if disabled and set inactive
		if !conf.IsConfigEnabled(config) {
			if conf.IsConfigActive(config) {
				_, err := conf.SetConfigActiveState(context.Background(), config, false)
				if err != nil {
					log.Fatal("conf", "Couldn't set config active state to DB: %v", err)
					return
				}
			}
			continue
		}

		// Signals that this config is active
		if !conf.IsConfigActive(config) {
			_, err := conf.SetConfigActiveState(context.Background(), config, true)
			if err != nil {
				log.Fatal("conf", "Couldn't set config active state to DB: %v", err)
				return
			}
			log.Debug("conf", "Collecting initialized with Configuration %d:\n"+
				"Enable: %t\n"+
				"Refresh Interval: %d\n"+
				"Request Timeout: %d\n"+
				"Active: %t\n"+
				"Project IDs: %v\n",
				*config.Id,
				*config.Enable,
				config.RefreshInterval,
				*config.RequestTimeout,
				*config.Active,
				*config.ProjectIDs)
		}

		common.RunOnceWithParam(func(config apiserver.Configuration) {
			log.Info("main", "Start collecting for configuration id %d", *config.Id)

			spaces, err := collectObjects(config)
			if err != nil {
				log.Error("collect", "Error collect spaces: %v", err)
				return
			}

			anyCreated, err := createAssets(config, spaces)
			if err != nil {
				log.Error("send", "Error sending assets: %v", err)
				return
			}
			log.Info("main", "Finished collecting for configuration id %d successfully", *config.Id)

			if anyCreated {
				log.Info("main", "Cancel existing subscriptions and resubscribe all")
				subscribeData()
			}

			time.Sleep(time.Second * time.Duration(config.RefreshInterval))
		}, config, *config.Id)
	}

}

// createAssets creates the complete asset tree, if the asset doesn't already exist
func createAssets(config apiserver.Configuration, spaces []signify.Object) (bool, error) {

	if config.ProjectIDs == nil || len(*config.ProjectIDs) == 0 {
		log.Info("eliona", "No project id defined in configuration %d. No data is send to Eliona.", config.Id)
		return false, nil
	}

	var anyCreated = false
	for _, projectId := range *config.ProjectIDs {

		rootAssetId, created, err := createAsset(config, projectId, eliona.RootAssetType, nil, nil, eliona.RootAssetType, conf.RootAssetKind, "Signify")
		if err != nil {
			return anyCreated, fmt.Errorf("create root asset first time: %w", err)
		}
		anyCreated = anyCreated || created

		for _, site := range spaces {

			siteAssetId, created, err := createAsset(config, projectId, site.Uuid, nil, &rootAssetId, eliona.GroupAssetType, conf.SiteAssetKind, site.Name)
			if err != nil {
				return anyCreated, fmt.Errorf("create site asset first time: %w", err)
			}
			anyCreated = anyCreated || created

			for _, building := range site.Children {

				buildingAssetId, created, err := createAsset(config, projectId, building.Uuid, common.Ptr(site.Uuid), &siteAssetId, eliona.GroupAssetType, conf.BuildingAssetKind, building.Name)
				if err != nil {
					return anyCreated, fmt.Errorf("create building asset first time: %w", err)
				}
				anyCreated = anyCreated || created

				for _, storey := range building.Children {

					storeyAssetId, created, err := createAsset(config, projectId, storey.Uuid, common.Ptr(building.Uuid), &buildingAssetId, eliona.GroupAssetType, conf.StoreyAssetKind, storey.Name)
					if err != nil {
						return anyCreated, fmt.Errorf("create storey asset first time: %w", err)
					}
					anyCreated = anyCreated || created

					for _, space := range storey.Children {

						if space.SpaceType == signify.OccupancySpaceType {
							_, created, err := createAsset(config, projectId, space.Uuid, common.Ptr(storey.Uuid), &storeyAssetId, eliona.OccupancyAssetType, conf.SpaceAssetKind, space.Name)
							if err != nil {
								return anyCreated, fmt.Errorf("create space asset first time: %w", err)
							}
							anyCreated = anyCreated || created
						}
						if space.SpaceType == signify.PeopleCountSpaceType {
							_, created, err := createAsset(config, projectId, space.Uuid, common.Ptr(storey.Uuid), &storeyAssetId, eliona.PeopleCountAssetType, conf.SpaceAssetKind, space.Name)
							if err != nil {
								return anyCreated, fmt.Errorf("create space asset first time: %w", err)
							}
							anyCreated = anyCreated || created
						}
						if space.SpaceType == signify.TemperatureSpaceType {
							_, created, err := createAsset(config, projectId, space.Uuid, common.Ptr(storey.Uuid), &storeyAssetId, eliona.TemperatureAssetType, conf.SpaceAssetKind, space.Name)
							if err != nil {
								return anyCreated, fmt.Errorf("create space asset first time: %w", err)
							}
							anyCreated = anyCreated || created
						}
						if space.SpaceType == signify.HumiditySpaceType {
							_, created, err := createAsset(config, projectId, space.Uuid, common.Ptr(storey.Uuid), &storeyAssetId, eliona.HumidityAssetType, conf.SpaceAssetKind, space.Name)
							if err != nil {
								return anyCreated, fmt.Errorf("create space asset first time: %w", err)
							}
							anyCreated = anyCreated || created
						}

					}
				}
			}
		}
	}

	return anyCreated, nil
}

func collectObjects(config apiserver.Configuration) ([]signify.Object, error) {

	// Sites
	sites, err := signify.GetSites(config)
	if err != nil {
		return nil, err
	}
	for siteIdx, site := range sites {
		log.Debug("collect", "Site: %s", site.Name)

		// Buildings
		buildings, err := signify.GetBuildings(config, site)
		if err != nil {
			return nil, err
		}
		sites[siteIdx].Children = buildings
		for buildingIdx, building := range buildings {
			log.Debug("collect", "Building: %s", building.Name)

			// Storeys
			storeys, err := signify.GetStoreys(config, building)
			if err != nil {
				return nil, err
			}
			sites[siteIdx].Children[buildingIdx].Children = storeys
			for storeyIdx, storey := range storeys {
				log.Debug("collect", "Storey: %s", storey.Name)

				// Spaces
				spaces, err := signify.GetSensorSpaces(config, storey)
				if err != nil {
					return nil, err
				}
				sites[siteIdx].Children[buildingIdx].Children[storeyIdx].Children = spaces
				for _, space := range spaces {
					log.Debug("collect", "Space: %s", space.Name)
				}
			}
		}
	}

	return sites, nil
}

// createAsset creates an asset if not exists. otherwise the current asset id is returned.
func createAsset(config apiserver.Configuration, projectId string, identifier string, parentIdentifier *string, parentId *int32, assetType string, kind conf.AssetKind, name string) (int32, bool, error) {
	uniqueIdentifier := assetType + "_" + identifier
	ctx := context.Background()

	// check if asset already exists in app
	assetId, err := conf.GetAssetIdWithGAI(ctx, config, projectId, uniqueIdentifier)
	if err != nil {
		return 0, false, fmt.Errorf("get asset id for %s in app: %w", uniqueIdentifier, err)
	}

	// if not, create asset in Eliona also
	if assetId == nil {

		log.Debug("assets", "No asset id found for %s", uniqueIdentifier)
		assetId, err = eliona.UpsertAsset(projectId, uniqueIdentifier, parentId, assetType, name)
		if err != nil || assetId == nil {
			return 0, false, fmt.Errorf("upserting root asset %s in Eliona: %w", uniqueIdentifier, err)
		}

		err = conf.InsertAsset(ctx, config, projectId, identifier, parentIdentifier, uniqueIdentifier, kind, *assetId)
		if err != nil {
			return 0, false, fmt.Errorf("insert asset %s in app: %w", uniqueIdentifier, err)
		}
		log.Debug("assets", "Asset created for %s with id %d", uniqueIdentifier, *assetId)

		return *assetId, true, nil
	} else {
		log.Debug("assets", "Asset already created for %s with id %d", uniqueIdentifier, *assetId)
		return *assetId, false, nil
	}
}

var subscriptionsMutex sync.Mutex

// subscribeData subscribes for new data
func subscribeData() {

	configs, err := conf.GetConfigs(context.Background())
	if err != nil {
		log.Fatal("conf", "Couldn't read configs from DB: %v", err)
		return
	}
	if len(configs) == 0 {
		return
	}

	// wait until a previous starting of socket is finished
	// otherwise not all previous connections can be closed
	subscriptionsMutex.Lock()
	defer subscriptionsMutex.Unlock()

	signify.CloseExistingSubscriptions()

	// create all connection for all configurations
	for _, config := range configs {

		if !conf.IsConfigEnabled(config) {
			continue
		}

		log.Info("main", "Start subscribing new data for configuration id %d", *config.Id)

		buildings, err := conf.GetAssets(context.Background(),
			appdb.AssetWhere.ConfigurationID.EQ(null.Int64FromPtr(config.Id).Int64),
			appdb.AssetWhere.Kind.EQ(string(conf.BuildingAssetKind)),
		)
		if err != nil {
			log.Fatal("listening", "Error collect buildings: %v", err)
			return
		}

		for _, subscriptionType := range []signify.SubscriptionType{signify.OccupancySubscriptionType, signify.HumiditySubscriptionType, signify.TemperatureSubscriptionType, signify.PeopleCountSubscriptionType} {
			for _, building := range buildings {
				url, err := signify.GetSubscriptionUrl(config, building.UUID, subscriptionType)
				if err != nil {
					log.Error("listening", "Error getting websocket URL: %v", err)
					continue
				}
				signify.Subscribe(*url, func(message signify.Message) {
					upsertData(message, config)
				})
			}
		}

		log.Info("main", "Finished subscribing new data for configuration id %d successfully", *config.Id)
	}
}

// upsertData upsert data
func upsertData(message signify.Message, config apiserver.Configuration) {
	spaces, err := conf.GetAssets(context.Background(),
		appdb.AssetWhere.ConfigurationID.EQ(null.Int64FromPtr(config.Id).Int64),
		appdb.AssetWhere.UUID.EQ(message.SpaceId),
	)
	if err != nil {
		log.Fatal("data", "Error getting assets with UUID %s: %v", message.SpaceId, err)
	}
	for _, space := range spaces {
		err := eliona.UpsertData(space.AssetID.Int32, message)
		if err != nil {
			log.Error("data", "Error upsert data %v: %v", message, err)
		}
	}
}

// listenApi starts the API server and listen for requests
func listenApi() {
	err := http.ListenAndServe(":"+common.Getenv("API_SERVER_PORT", "3000"),
		frontend.NewEnvironmentHandler(
			utilshttp.NewCORSEnabledHandler(
				apiserver.NewRouter(
					apiserver.NewConfigurationAPIController(apiservices.NewConfigurationApiService()),
					apiserver.NewVersionAPIController(apiservices.NewVersionApiService()),
					apiserver.NewCustomizationAPIController(apiservices.NewCustomizationApiService()),
				),
			),
		),
	)
	log.Fatal("main", "API server: %v", err)
}
