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
	"github.com/eliona-smart-building-assistant/go-utils/common"
	utilshttp "github.com/eliona-smart-building-assistant/go-utils/http"
	"github.com/eliona-smart-building-assistant/go-utils/log"
	"github.com/gorilla/websocket"
	"signify/apiserver"
	"signify/eliona"
	"time"
)

type Object struct {
	ObjectType   ObjectType `eliona:"object_type,filterable"`
	Name         string     `json:"name" eliona:"name,filterable"`
	Uuid         string     `json:"uuid" eliona:"uuid,filterable"`
	FunctionType string     `json:"functionType" eliona:"function_type,filterable"`
	SpaceType    string     `json:"spaceType" eliona:"space_type,filterable"`
	Children     []Object
}

const (
	OccupancySpaceType   = "occupancy"
	PeopleCountSpaceType = "peoplecount"
	TemperatureSpaceType = "temperature"
	HumiditySpaceType    = "humidity"
)

type ObjectType string

const (
	SiteObjectType     ObjectType = "site"
	BuildingObjectType ObjectType = "building"
	StoreyObjectType   ObjectType = "storey"
	SpaceObjectType    ObjectType = "space"
)

type OccupancyState string

const (
	OccupiedOccupancyState   OccupancyState = "occupied"
	UnoccupiedOccupancyState OccupancyState = "unoccupied"
	UnknownOccupancyState    OccupancyState = "unknown"
)

type SubscriptionType string

const (
	OccupancySubscriptionType   SubscriptionType = "OCCUPANCY"
	HumiditySubscriptionType    SubscriptionType = "HUMIDITY"
	TemperatureSubscriptionType SubscriptionType = "TEMPERATURE"
	PeopleCountSubscriptionType SubscriptionType = "PEOPLE_COUNT"
)

type Message struct {
	SpaceId        string          `json:"spaceId"`
	Timestamp      int64           `json:"timestamp"`
	Count          *int            `json:"count" eliona:"people_count" subtype:"input"`
	Temperature    *int            `json:"temperature" eliona:"temperature" subtype:"input"`
	Humidity       *int            `json:"humidity" eliona:"humidity" subtype:"input"`
	Unit           *string         `json:"unit"`
	OccupancyState *OccupancyState `json:"occupancy"`
	Occupancy      *int            `json:"-" eliona:"occupancy" subtype:"input"`
}

type WebsocketUrl struct {
	Url       *string `json:"websocketUrl"`
	Timestamp int64   `json:"timestamp"`
	Errors    any     `json:"errors"`
}

var subscriptions []*websocket.Conn

func fetchObjects(config apiserver.Configuration, endpoint string, objectType ObjectType) ([]Object, error) {
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

func Subscribe(url string, messageHandler func(message Message)) {
	messages := make(chan Message)
	defer close(messages)

	// start listening
	go func() {
		err := utilshttp.ListenWebSocketWithReconnectAlways(subscriptionCreator(url), time.Second*10, messages)
		if err != nil {
			log.Error("Listening", "Error listening on %s: %v", url, err)
		}
	}()
	go func() {
		log.Debug("Listening", "Start listening on: %s", url)
		for message := range messages {
			log.Debug("Listening", "New message from %s: %v", url, message)
			if message.OccupancyState != nil {
				switch *message.OccupancyState {
				case OccupiedOccupancyState:
					message.Occupancy = common.Ptr(1)
				case UnoccupiedOccupancyState:
					message.Occupancy = common.Ptr(-1)
				case UnknownOccupancyState:
					message.Occupancy = common.Ptr(0)
				}
			}
			messageHandler(message)
		}
		log.Info("Listening", "Stop listening on %s", url)
	}()
}

func subscriptionCreator(url string) func() (*websocket.Conn, error) {
	return func() (*websocket.Conn, error) {
		log.Info("Listening", "Create subscription for %s", url)
		subscription, err := utilshttp.NewWebSocketConnectionWithApiKey(url, "", "")
		subscriptions = append(subscriptions, subscription)
		return subscription, err
	}
}

func CloseExistingSubscriptions() {
	for _, subscription := range subscriptions {
		if subscription != nil {
			_ = subscription.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		}
	}
}

func GetSubscriptionUrl(config apiserver.Configuration, buildingUUID string, subscriptionType SubscriptionType) (*string, error) {
	token, err := getBearerToken(config)
	if err != nil {
		return nil, err
	}
	endpoint := "/interact/api/officeCloud/v1/subscription/" + buildingUUID + "/" + string(subscriptionType)
	request, err := utilshttp.NewRequestWithBearer(config.BaseUrl+endpoint, token.Token)
	if err != nil {
		return nil, fmt.Errorf("request %s: %w", endpoint, err)
	}

	websocketUrl, err := utilshttp.Read[WebsocketUrl](request, time.Duration(*config.RequestTimeout)*time.Second, true)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", endpoint, err)
	}
	if websocketUrl.Url == nil {
		return nil, fmt.Errorf("read %s: %v", endpoint, websocketUrl.Errors)
	}
	return websocketUrl.Url, nil
}
