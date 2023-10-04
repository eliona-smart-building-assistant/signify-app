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
	"net/http"
	"signify/apiserver"
	"signify/apiservices"
	"signify/conf"
	"signify/eliona"
	"signify/signify"
	"time"

	"github.com/eliona-smart-building-assistant/go-utils/common"
	utilshttp "github.com/eliona-smart-building-assistant/go-utils/http"
	"github.com/eliona-smart-building-assistant/go-utils/log"
)

func collectData() {
	configs, err := conf.GetConfigs(context.Background())
	if err != nil {
		log.Fatal("conf", "couldn't read configs from DB: %v", err)
		return
	}
	if len(configs) == 0 {
		log.Info("conf", "no configs in DB")
		return
	}

	for _, config := range configs {

		// Skip config if disabled and set inactive
		if !conf.IsConfigEnabled(config) {
			if conf.IsConfigActive(config) {
				_, err := conf.SetConfigActiveState(context.Background(), config, false)
				if err != nil {
					log.Fatal("conf", "couldn't set config active state to DB: %v", err)
					return
				}
			}
			continue
		}

		// Signals that this config is active
		if !conf.IsConfigActive(config) {
			_, err := conf.SetConfigActiveState(context.Background(), config, true)
			if err != nil {
				log.Fatal("conf", "couldn't set config active state to DB: %v", err)
				return
			}
			log.Info("conf", "collecting initialized with Configuration %d:\n"+
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
			log.Info("main", "collecting %d started", *config.Id)

			spaces, err := collectSpaces(config)
			if err != nil {
				log.Error("collect", "error collect spaces: %v", err)
				return
			} else {
				err = sendSpaces(config, spaces)
				if err != nil {
					log.Error("send", "error sending assets: %v", err)
					return
				} else {
					log.Info("main", "collecting %d successful finished", *config.Id)
				}
			}

			time.Sleep(time.Second * time.Duration(config.RefreshInterval))
		}, config, *config.Id)
	}
}

func sendSpaces(config apiserver.Configuration, spaces []signify.Object) error {

	if config.ProjectIDs == nil || len(*config.ProjectIDs) == 0 {
		log.Info("eliona", "No project id defined in configuration %d. No data is send to Eliona.", config.Id)
		return nil
	}
	for _, projectId := range *config.ProjectIDs {

		rootAssetId, err := createAssetFirstTime(config, projectId, eliona.SignifyRootAssetType, nil, eliona.SignifyRootAssetType, "Signify")
		if err != nil {
			return fmt.Errorf("create root asset first time: %w", err)
		}

		for _, site := range spaces {

			siteAssetId, err := createAssetFirstTime(config, projectId, site.Uuid, &rootAssetId, eliona.SignifyGroupAssetType, site.Name)
			if err != nil {
				return fmt.Errorf("create site asset first time: %w", err)
			}

			for _, building := range site.Children {

				buildingAssetId, err := createAssetFirstTime(config, projectId, building.Uuid, &siteAssetId, eliona.SignifyGroupAssetType, building.Name)
				if err != nil {
					return fmt.Errorf("create building asset first time: %w", err)
				}

				for _, storey := range building.Children {

					storeyAssetId, err := createAssetFirstTime(config, projectId, storey.Uuid, &buildingAssetId, eliona.SignifyGroupAssetType, storey.Name)
					if err != nil {
						return fmt.Errorf("create storey asset first time: %w", err)
					}

					for _, space := range storey.Children {

						_, err := createAssetFirstTime(config, projectId, space.Uuid, &storeyAssetId, eliona.SignifySpaceAssetType, space.Name)
						if err != nil {
							return fmt.Errorf("create space asset first time: %w", err)
						}

					}
				}
			}
		}
	}

	return nil
}

func createAssetFirstTime(config apiserver.Configuration, projectId string, identifier string, parentId *int32, assetType string, name string) (int32, error) {
	uniqueIdentifier := assetType + "_" + identifier
	ctx := context.Background()

	// check if asset already exists in app
	assetId, err := conf.GetAssetId(ctx, config, projectId, uniqueIdentifier)
	if err != nil {
		return 0, fmt.Errorf("get asset id for %s in app: %w", uniqueIdentifier, err)
	}

	// if not, create asset in Eliona also
	if assetId == nil {

		log.Debug("assets", "no asset id found for %s", uniqueIdentifier)
		assetId, err = eliona.UpsertAsset(projectId, uniqueIdentifier, parentId, assetType, name)
		if err != nil || assetId == nil {
			return 0, fmt.Errorf("upserting root asset %s in Eliona: %w", uniqueIdentifier, err)
		}

		err = conf.InsertAsset(ctx, config, projectId, uniqueIdentifier, *assetId)
		if err != nil {
			return 0, fmt.Errorf("insert asset %s in app: %w", uniqueIdentifier, err)
		}
		log.Debug("assets", "asset created for %s with id %d", uniqueIdentifier, *assetId)

	} else {
		log.Debug("assets", "asset already created for %s with id %d", uniqueIdentifier, *assetId)
	}

	return *assetId, nil
}

func collectSpaces(config apiserver.Configuration) ([]signify.Object, error) {

	// Sites
	sites, err := signify.GetSites(config)
	if err != nil {
		return nil, err
	}
	for siteIdx, site := range sites {
		log.Debug("collect", "site: %s", site.Name)

		// Buildings
		buildings, err := signify.GetBuildings(config, site)
		if err != nil {
			return nil, err
		}
		sites[siteIdx].Children = buildings
		for buildingIdx, building := range buildings {
			log.Debug("collect", "building: %s", building.Name)

			// Storeys
			storeys, err := signify.GetStoreys(config, building)
			if err != nil {
				return nil, err
			}
			sites[siteIdx].Children[buildingIdx].Children = storeys
			for storeyIdx, storey := range storeys {
				log.Debug("collect", "storey: %s", storey.Name)

				// Spaces
				spaces, err := signify.GetSensorSpaces(config, storey)
				if err != nil {
					return nil, err
				}
				sites[siteIdx].Children[buildingIdx].Children[storeyIdx].Children = spaces
				for _, space := range spaces {
					log.Debug("collect", "space: %s", space.Name)
				}
			}
		}
	}

	return sites, nil

}

// listenApi starts the API server and listen for requests
func listenApi() {
	err := http.ListenAndServe(":"+common.Getenv("API_SERVER_PORT", "3000"), utilshttp.NewCORSEnabledHandler(
		apiserver.NewRouter(
			apiserver.NewConfigurationAPIController(apiservices.NewConfigurationApiService()),
			apiserver.NewVersionAPIController(apiservices.NewVersionApiService()),
			apiserver.NewCustomizationAPIController(apiservices.NewCustomizationApiService()),
		)),
	)
	log.Fatal("main", "API server: %v", err)
}
