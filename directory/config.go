// Copyright 2015 CoreStore Authors
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

package directory

import (
	"github.com/corestoreio/csfw/config"
	"github.com/corestoreio/csfw/storage/csdb"
)

const (
	PathSystemCurrencyInstalled = "system/currency/installed"
	PathCurrencyAllow           = "currency/options/allow"
	PathCurrencyDefault         = "currency/options/default"
	PathCurrencyBase            = "currency/options/base"
)

var (
	// configReader stores the reader. Should not be used. Access it via mustConfig()
	configReader config.ScopeReader
	// TableCollection handles all tables and its columns. init() in generated Go file will set the value.
	TableCollection csdb.TableStructureSlice
)

// SetConfig sets the internal variable to the current scope config reader.
// ScopeReader will be used across all functions in this package.
func SetConfigReader(c config.ScopeReader) {
	if c == nil {
		panic("config.ScopeReader cannot be nil")
	}
	configReader = c
}

// mustReadConfig internally used
func mustReadConfig() config.ScopeReader {
	if configReader == nil {
		panic("config.ScopeReader cannot be nil")
	}
	return configReader
}

// GetDefaultConfiguration in conjunction with config.Scope.ApplyDefaults function to
// set the default configuration value for a package.
func GetDefaultConfiguration() config.DefaultMap {
	return config.DefaultMap{
		PathSystemCurrencyInstalled: "AZN,AZM,AFN,ALL,DZD,AOA,ARS,AMD,AWG,AUD,BSD,BHD,BDT,BBD,BYR,BZD,BMD,BTN,BOB,BAM,BWP,BRL,GBP,BND,BGN,BUK,BIF,KHR,CAD,CVE,CZK,KYD,CLP,CNY,COP,KMF,CDF,CRC,HRK,CUP,DKK,DJF,DOP,XCD,EGP,SVC,GQE,ERN,EEK,ETB,EUR,FKP,FJD,GMD,GEK,GEL,GHS,GIP,GTQ,GNF,GYD,HTG,HNL,HKD,HUF,ISK,INR,IDR,IRR,IQD,ILS,JMD,JPY,JOD,KZT,KES,KWD,KGS,LAK,LVL,LBP,LSL,LRD,LYD,LTL,MOP,MKD,MGA,MWK,MYR,MVR,LSM,MRO,MUR,MXN,MDL,MNT,MAD,MZN,MMK,NAD,NPR,ANG,TRL,TRY,NZD,NIC,NGN,KPW,NOK,OMR,PKR,PAB,PGK,PYG,PEN,PHP,PLN,QAR,RHD,RON,ROL,RUB,RWF,SHP,STD,SAR,RSD,SCR,SLL,SGD,SKK,SBD,SOS,ZAR,KRW,LKR,SDG,SRD,SZL,SEK,CHF,SYP,TWD,TJS,TZS,THB,TOP,TTD,TND,TMM,USD,UGX,UAH,AED,UYU,UZS,VUV,VEB,VEF,VND,CHE,CHW,XOF,XPF,WST,YER,ZMK,ZWD",
		PathCurrencyAllow:           "USD,EUR",
		PathCurrencyBase:            "EUR",
		PathCurrencyDefault:         "EUR",
	}
}