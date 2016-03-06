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

package storage_test

import (
	"testing"

	"github.com/corestoreio/csfw/config/path"
	"github.com/corestoreio/csfw/config/storage"
	"github.com/corestoreio/csfw/store/scope"
	"github.com/stretchr/testify/assert"
)

var _ storage.Storager = storage.NewKV()

func TestSimpleStorage(t *testing.T) {
	t.Parallel()
	sp := storage.NewKV()

	p1 := path.MustNewByParts("aa/bb/cc")

	assert.NoError(t, sp.Set(p1, 19.99))
	f, err := sp.Get(p1)
	assert.NoError(t, err)
	assert.Exactly(t, 19.99, f.(float64))

	p2 := path.MustNewByParts("xx/yy/zz").Bind(scope.StoreID, 2)

	assert.NoError(t, sp.Set(p2, 4711))
	i, err := sp.Get(p2)
	assert.NoError(t, err)
	assert.Exactly(t, 4711, i.(int))

	ni, err := sp.Get(path.Path{})
	assert.EqualError(t, err, path.ErrIncorrectPath.Error())
	assert.Nil(t, ni)

	keys, err := sp.AllKeys()
	assert.NoError(t, err)
	keys.Sort()

	wantKeys := path.PathSlice{path.Path{Route: path.NewRoute(`aa/bb/cc`), Scope: 1, ID: 0}, path.Path{Route: path.NewRoute(`xx/yy/zz`), Scope: 4, ID: 2}}
	assert.Exactly(t, wantKeys, keys)

	p3 := path.MustNewByParts("rr/ss/tt").Bind(scope.StoreID, 1)
	ni, err = sp.Get(p3)
	assert.EqualError(t, err, storage.ErrKeyNotFound.Error())
	assert.Nil(t, ni)
}