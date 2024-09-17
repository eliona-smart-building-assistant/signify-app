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
	"github.com/eliona-smart-building-assistant/go-utils/log"
	"signify/apiserver"
	"time"
)

type BearerToken struct {
	Token     string         `json:"token"`
	ExpiresIn int            `json:"expires_in"`
	Fault     map[string]any `json:"fault"`
	Issued    int64
}

func getBearerToken(config apiserver.Configuration) (*BearerToken, error) {
	if bearerTokenValid(config) {
		log.Debug("auth", "Reuse bearer bearerToken: %.10s...", bearerTokens[*config.Id].Token)
		return bearerTokens[*config.Id], nil
	}
	request, err := utilshttp.NewPostFormRequestWithBasicAuth(config.BaseUrl+"/oauth/accesstoken", map[string][]string{
		"app_key":    {config.AppKey},
		"app_secret": {config.AppSecret},
		"service":    {config.Service},
	}, config.ServiceId, config.ServiceSecret)
	if err != nil {
		return nil, fmt.Errorf("request /oauth/accesstoken: %w", err)
	}
	token, err := utilshttp.Read[BearerToken](request, time.Duration(*config.RequestTimeout)*time.Second, true)
	if err != nil {
		return nil, fmt.Errorf("read /oauth/accesstoken: %w", err)
	}
	if token.Fault != nil {
		return nil, fmt.Errorf("read /oauth/accesstoken: %v", token.Fault["faultstring"])
	}
	token.Issued = time.Now().Unix()
	bearerTokens[*config.Id] = &token
	log.Info("auth", "Created new Bearer Token for %d: %.10s...", *config.Id, token.Token)
	return &token, nil
}

var bearerTokens = make(map[int64]*BearerToken)

func resetBearerToken(config apiserver.Configuration) {
	log.Info("auth", "Reset Bearer Token for %d", *config.Id)
	delete(bearerTokens, *config.Id)
}

func bearerTokenValid(config apiserver.Configuration) bool {
	token, found := bearerTokens[*config.Id]
	return found && token.Token != "" && token.Issued+int64(token.ExpiresIn) >= time.Now().Unix()-300
}
