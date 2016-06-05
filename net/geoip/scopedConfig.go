// Copyright 2015-2016, Cyrill @ Schumacher.fm and the CoreStore contributors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package geoip

import (
	"net/http"

	"github.com/corestoreio/csfw/log"
	"github.com/corestoreio/csfw/store"
	"github.com/corestoreio/csfw/store/scope"
	"github.com/corestoreio/csfw/util"
	"github.com/corestoreio/csfw/util/errors"
)

// scopedConfig private internal scoped based configuration
type scopedConfig struct {
	// useDefault if true uses the default configuration and all other fields are
	// empty.
	useDefault bool
	// lastErr used during selecting the config from the scopeCache map and gets
	// filled if an entry cannot be found.
	lastErr error
	// scopeHash defines the scope to which this configuration is bound to.
	scopeHash scope.Hash

	// log only available in the scope of DefaultHash
	log log.Logger

	// AllowedCountries a model containing a path to the configuration which
	// countries are allowed within a scope. Current implementation triggers for
	// each HTTP request a configuration lookup which can be a bottle neck.
	allowedCountries []string
	// IsAllowedFunc checks in middleware WithIsCountryAllowedByIP if the country is
	// allowed to process the request.
	IsAllowedFunc // func(s *store.Store, c *Country, allowedCountries []string, r *http.Request) bool

	// alternativeHandler if ip/country is denied we call this handler
	alternativeHandler http.Handler
}

func defaultScopedConfig(h scope.Hash) (scopedConfig, error) {
	var l log.Logger
	if h == scope.DefaultHash {
		l = log.BlackHole{} // disabled info and debug logging
	}
	return scopedConfig{
		scopeHash: h,
		log:       l,
		IsAllowedFunc: func(_ *store.Store, c *Country, allowedCountries []string, _ *http.Request) bool {
			var ac util.StringSlice = allowedCountries
			return ac.Contains(c.Country.IsoCode)
		},
		alternativeHandler: DefaultAlternativeHandler,
	}, nil
}

// IsValid a configuration for a scope is only then valid when the Key has been
// supplied, a non-nil signing method and a non-nil Verifier.
func (sc scopedConfig) isValid() error {
	if sc.lastErr != nil {
		return sc.lastErr
	}
	if sc.scopeHash == 0 || sc.IsAllowedFunc == nil || sc.alternativeHandler == nil {
		return errors.NewNotValidf(errScopedConfigNotValid, sc.scopeHash)
	}
	return nil
}

func (sc scopedConfig) checkAllow(reqSt *store.Store, c *Country, r *http.Request) bool {
	if len(sc.allowedCountries) == 0 {
		return true
	}
	return sc.IsAllowedFunc(reqSt, c, sc.allowedCountries, r)
}
