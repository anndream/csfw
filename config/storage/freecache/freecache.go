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

package freecache

import (
	"time"

	"github.com/coocood/freecache"
	"github.com/corestoreio/csfw/config/cfgpath"
	"github.com/corestoreio/csfw/config/storage"
	"github.com/corestoreio/csfw/util/conv"
)

// Storage wrapper around the freecache.Cache type
type Storage struct {
	// Expiration defines the global value when a key expires.
	// Expiration <= 0 means no expire, but it can be evicted when the
	// cache is full.
	Expiration time.Duration
	*freecache.Cache
}

// New creates a new cache with a minimum size set to 512KB.
// If the size is set relatively large, you should call
// `debug.SetGCPercent()`, set it to a much smaller value
// to limit the memory consumption and GC pause time.
func New(size int) *Storage {
	return &Storage{
		Cache: freecache.NewCache(size),
	}
}

// Set writes a key with its value into the storage. The value
// gets converted to a byte slice.
func (s *Storage) Set(key cfgpath.Path, value interface{}) error {
	fq, err := key.FQ()
	if err != nil {
		return err
	}
	b, err := conv.ToByteE(value)
	if err != nil {
		return err
	}
	return s.Cache.Set(fq.Bytes(), b, int(s.Expiration.Seconds()))
}

// Get may return a ErrKeyNotFound error
func (s *Storage) Get(key cfgpath.Path) (interface{}, error) {
	fq, err := key.FQ()
	if err != nil {
		return nil, err
	}
	val, err := s.Cache.Get(fq.Bytes())
	if err != nil && err != freecache.ErrNotFound {
		return nil, err
	}
	if err == freecache.ErrNotFound {
		return nil, storage.NotFound{}
	}

	return val, nil
}

// AllKeys returns always nil. Function not supported.
func (s *Storage) AllKeys() (cfgpath.PathSlice, error) {
	return nil, nil
}
