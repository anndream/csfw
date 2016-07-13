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

package memstore

import (
	"github.com/corestoreio/csfw/config"
	"github.com/corestoreio/csfw/net/ratelimit"
	"github.com/corestoreio/csfw/net/ratelimit/backendratelimit"
	"github.com/corestoreio/csfw/util/errors"
)

// OptionName identifies this package within the register of the
// backendratelimit.Backend type.
const OptionName = `memstore`

// NewOptionFactory creates a new option factory function for the memstore in the
// backend package to be used for automatic scope based configuration
// initialization. Configuration values are read from argument `be`.
func NewOptionFactory(be *backendratelimit.Backend) (string, ratelimit.OptionFactoryFunc) {
	return OptionName, func(sg config.Scoped) []ratelimit.Option {

		burst, _, err := be.RateLimitBurst.Get(sg)
		if err != nil {
			return ratelimit.OptionsError(errors.Wrap(err, "[memstore] RateLimitBurst.Get"))
		}
		req, _, err := be.RateLimitRequests.Get(sg)
		if err != nil {
			return ratelimit.OptionsError(errors.Wrap(err, "[memstore] RateLimitRequests.Get"))
		}
		durRaw, _, err := be.RateLimitDuration.Get(sg)
		if err != nil {
			return ratelimit.OptionsError(errors.Wrap(err, "[memstore] RateLimitDuration.Get"))
		}

		if len(durRaw) != 1 {
			return ratelimit.OptionsError(errors.NewFatalf("[memstore] RateLimitDuration invalid character count: %q. Should be one character long.", durRaw))
		}

		dur := rune(durRaw[0])

		useInMemMaxKeys, scpHash, err := be.RateLimitStorageGcraMaxMemoryKeys.Get(sg)
		if err != nil {
			return ratelimit.OptionsError(errors.Wrap(err, "[memstore] RateLimitStorageGcraMaxMemoryKeys.Get"))
		} else if useInMemMaxKeys > 0 {
			scp, scpID := scpHash.Unpack()
			return []ratelimit.Option{
				WithGCRA(scp, scpID, useInMemMaxKeys, dur, req, burst),
			}
		}
		return ratelimit.OptionsError(errors.NewEmptyf("[memstore] Memstore not active because RateLimitStorageGcraMaxMemoryKeys is %d.", useInMemMaxKeys))
	}
}
