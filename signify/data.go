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
	"time"
)

type Object struct {
	Name         string `json:"name"`
	Uuid         string `json:"uuid"`
	FunctionType string `json:"functionType"`
	Children     []Object
}

func fetchData(config apiserver.Configuration, endpoint string) ([]Object, error) {
	token, err := getBearerToken(config)
	if err != nil {
		return nil, err
	}

	request, err := utilshttp.NewRequestWithBearer(config.BaseUrl+endpoint, token.Token)
	if err != nil {
		return nil, fmt.Errorf("request %s: %w", endpoint, err)
	}

	data, err := utilshttp.Read[[]Object](request, time.Duration(*config.RequestTimeout)*time.Second, true)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", endpoint, err)
	}

	return data, nil
}

func GetSites(config apiserver.Configuration) ([]Object, error) {
	return fetchData(config, "/interact/api/officeCloud/v1/sites")
}

func GetBuildings(config apiserver.Configuration, site Object) ([]Object, error) {
	return fetchData(config, "/interact/api/officeCloud/v1/sites/"+site.Uuid+"/buildings")
}

func GetStoreys(config apiserver.Configuration, building Object) ([]Object, error) {
	return fetchData(config, "/interact/api/officeCloud/v1/buildings/"+building.Uuid+"/buildingStoreys")
}

func GetSensorSpaces(config apiserver.Configuration, storey Object) ([]Object, error) {
	return fetchData(config, "/interact/api/officeCloud/v1/buildingStoreys/"+storey.Uuid+"/sensorSpaces")
}
