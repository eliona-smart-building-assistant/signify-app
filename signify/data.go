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

package signify

import (
	"fmt"
	utilshttp "github.com/eliona-smart-building-assistant/go-utils/http"
	"signify/apiserver"
	"signify/eliona"
	"time"
)

type Object struct {
	ObjectType   string `eliona:"object_type,filterable"`
	Name         string `json:"name" eliona:"name,filterable"`
	Uuid         string `json:"uuid" eliona:"uuid,filterable"`
	FunctionType string `json:"functionType" eliona:"function_type,filterable"`
	SpaceType    string `json:"spaceType" eliona:"space_type,filterable"`
	Children     []Object
}

const (
	OccupancySpaceType   = "occupancy"
	TemperatureSpaceType = "temperature"
	HumiditySpaceType    = "humidity"
)

const (
	SiteObjectType     = "site"
	BuildingObjectType = "building"
	StoreyObjectType   = "storey"
	SpaceObjectType    = "space"
)

func fetchObjects(config apiserver.Configuration, endpoint string, objectType string) ([]Object, error) {
	token, err := getBearerToken(config)
	if err != nil {
		return nil, err
	}

	request, err := utilshttp.NewRequestWithBearer(config.BaseUrl+endpoint, token.Token)
	if err != nil {
		return nil, fmt.Errorf("request %s: %w", endpoint, err)
	}

	objects, err := utilshttp.Read[[]Object](request, time.Duration(*config.RequestTimeout)*time.Second, true)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", endpoint, err)
	}
	var filteredObjects = make([]Object, 0)
	for _, object := range objects {
		object.ObjectType = objectType

		shouldUse, err := eliona.AdheresToFilter(object, config.AssetFilter)
		if err != nil {
			return nil, fmt.Errorf("filtering object %s: %w", object.Name, err)
		}
		if !shouldUse {
			continue
		}

		filteredObjects = append(filteredObjects, object)
	}
	return filteredObjects, nil
}

func GetSites(config apiserver.Configuration) ([]Object, error) {
	return fetchObjects(config, "/interact/api/officeCloud/v1/sites", SiteObjectType)
}

func GetBuildings(config apiserver.Configuration, site Object) ([]Object, error) {
	return fetchObjects(config, "/interact/api/officeCloud/v1/sites/"+site.Uuid+"/buildings", BuildingObjectType)
}

func GetStoreys(config apiserver.Configuration, building Object) ([]Object, error) {
	return fetchObjects(config, "/interact/api/officeCloud/v1/buildings/"+building.Uuid+"/buildingStoreys", StoreyObjectType)
}

func GetSensorSpaces(config apiserver.Configuration, storey Object) ([]Object, error) {
	return fetchObjects(config, "/interact/api/officeCloud/v1/buildingStoreys/"+storey.Uuid+"/sensorSpaces", SpaceObjectType)
}
