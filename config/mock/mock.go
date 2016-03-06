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

package mock

import (
	"reflect"
	"time"

	"github.com/corestoreio/csfw/config"
	"github.com/corestoreio/csfw/config/internal/cfgctx"
	"github.com/corestoreio/csfw/config/path"
	"github.com/corestoreio/csfw/config/storage"
	"github.com/corestoreio/csfw/util/cast"
	"golang.org/x/net/context"
)

// MockWrite used for testing when writing configuration values.
type MockWrite struct {
	// WriteError gets always returned by Write
	WriteError error
	// ArgPath will be set after calling write to export the config path.
	// Values you enter here will be overwritten when calling Write
	ArgPath string
	// ArgValue contains the written data
	ArgValue interface{}
}

// Write writes to a black hole, may return an error
func (w *MockWrite) Write(p path.Path, v interface{}) error {
	w.ArgPath = p.String()
	w.ArgValue = v
	return w.WriteError
}

// mockOptionFunc to initialize the NewMockGetter
type mockOptionFunc func(*MockGet)

// MockGet used for testing. Contains functions which will be called in the
// appropriate methods of interface config.Getter.
// Using WithMockValues() has precedence over the applied functions.
type MockGet struct {
	db              storage.Storager
	FString         func(path string) (string, error)
	FBool           func(path string) (bool, error)
	FFloat64        func(path string) (float64, error)
	FInt            func(path string) (int, error)
	FTime           func(path string) (time.Time, error)
	SubscriptionID  int
	SubscriptionErr error
}

// MockPV is a required type for an option function. PV = path => value.
// This map[string]interface{} is protected by a mutex.
type MockPV map[string]interface{}

func (m MockPV) set(db storage.Storager) {
	for fq, v := range m {
		p, err := path.SplitFQ(fq)
		if err != nil {
			panic(err)
		}
		if err := db.Set(p, v); err != nil {
			panic(err)
		}
	}
}

// WithMockString returns a function which can be used in the NewMockGetter().
// Your function returns a string value from a given path.
// Call priority 2.
func WithMockString(f func(path string) (string, error)) mockOptionFunc {
	return func(mr *MockGet) { mr.FString = f }
}

// WithMockBool returns a function which can be used in the NewMockGetter().
// Your function returns a bool value from a given path.
// Call priority 2.
func WithMockBool(f func(path string) (bool, error)) mockOptionFunc {
	return func(mr *MockGet) { mr.FBool = f }
}

// WithMockFloat64 returns a function which can be used in the NewMockGetter().
// Your function returns a float64 value from a given path.
// Call priority 2.
func WithMockFloat64(f func(path string) (float64, error)) mockOptionFunc {
	return func(mr *MockGet) { mr.FFloat64 = f }
}

// WithMockInt returns a function which can be used in the NewMockGetter().
// Your function returns an int value from a given path.
// Call priority 2.
func WithMockInt(f func(path string) (int, error)) mockOptionFunc {
	return func(mr *MockGet) { mr.FInt = f }
}

// WithMockTime returns a function which can be used in the NewMockGetter().
// Your function returns a Time value from a given path.
// Call priority 2.
func WithMockTime(f func(path string) (time.Time, error)) mockOptionFunc {
	return func(mr *MockGet) {
		mr.FTime = f
	}
}

// WithMockValues lets you define a map of path and its values.
// Key is the fully qualified configuration path and value is the value.
// Value must be of the same type as returned by the functions.
// Panics on error.
// Call priority 1.
func WithMockValues(pathValues MockPV) mockOptionFunc {
	return func(mr *MockGet) {
		pathValues.set(mr.db)
	}
}

// WithContextMockGetter adds a MockGetter to a context.
func WithContextMockGetter(ctx context.Context, opts ...mockOptionFunc) context.Context {
	return context.WithValue(ctx, cfgctx.KeyGetter{}, NewMockGetter(opts...))
}

// WithContextMockScopedGetter adds a scoped MockGetter to a context.
func WithContextMockScopedGetter(websiteID, storeID int64, ctx context.Context, opts ...mockOptionFunc) context.Context {
	return context.WithValue(ctx, cfgctx.KeyScopedGetter{}, NewMockGetter(opts...).NewScoped(websiteID, storeID))
}

// NewMockGetter creates a new MockGetter used in testing.
// Allows you to set different options duration creation or you can
// set the struct fields afterwards.
// WithMockValues() option has priority over WithMock<T>() functions.
func NewMockGetter(opts ...mockOptionFunc) *MockGet {
	mr := &MockGet{
		db: storage.NewKV(),
	}
	for _, opt := range opts {
		opt(mr)
	}
	return mr
}

// UpdateValues adds or overwrites the internal path => value map.
func (mr *MockGet) UpdateValues(pathValues MockPV) {
	pathValues.set(mr.db)
}

func (mr *MockGet) hasVal(p path.Path) bool {
	v, err := mr.db.Get(p)
	if err != nil && config.NotKeyNotFoundError(err) {
		println("Mock.hasVal error:", err.Error(), "path", p.String())
	}
	return v != nil && err == nil
}

func (mr *MockGet) getVal(p path.Path) interface{} {
	v, err := mr.db.Get(p)
	if err != nil && config.NotKeyNotFoundError(err) {
		println("Mock.getVal error:", err.Error(), "path", p.String())
		return nil
	}
	v = indirect(v)
	return v
}

// String returns a string value
func (mr *MockGet) String(p path.Path) (string, error) {
	switch {
	case mr.hasVal(p):
		return cast.ToStringE(mr.getVal(p)) //mr.valString(p)
	case mr.FString != nil:
		return mr.FString(p.String())
	default:
		return "", storage.ErrKeyNotFound
	}
}

// Bool returns a bool value
func (mr *MockGet) Bool(p path.Path) (bool, error) {
	switch {
	case mr.hasVal(p):
		return cast.ToBoolE(mr.getVal(p))
	case mr.FBool != nil:
		return mr.FBool(p.String())
	default:
		return false, storage.ErrKeyNotFound
	}
}

// Float64 returns a float64 value
func (mr *MockGet) Float64(p path.Path) (float64, error) {
	switch {
	case mr.hasVal(p):
		return cast.ToFloat64E(mr.getVal(p))
	case mr.FFloat64 != nil:
		return mr.FFloat64(p.String())
	default:
		return 0.0, storage.ErrKeyNotFound
	}
}

// Int returns an integer value
func (mr *MockGet) Int(p path.Path) (int, error) {
	switch {
	case mr.hasVal(p):
		return cast.ToIntE(mr.getVal(p))
	case mr.FInt != nil:
		return mr.FInt(p.String())
	default:
		return 0, storage.ErrKeyNotFound
	}
}

// Time returns a time value
func (mr *MockGet) Time(p path.Path) (time.Time, error) {
	switch {
	case mr.hasVal(p):
		return cast.ToTimeE(mr.getVal(p))
	case mr.FTime != nil:
		return mr.FTime(p.String())
	default:
		return time.Time{}, storage.ErrKeyNotFound
	}
}

// Subscribe returns the before applied SubscriptionID and SubscriptionErr
// Does not start any underlying Goroutines.
func (mr *MockGet) Subscribe(_ path.Route, s config.MessageReceiver) (subscriptionID int, err error) {
	return mr.SubscriptionID, mr.SubscriptionErr
}

// NewScoped creates a new config.ScopedReader which uses the underlying
// mocked paths and values.
func (mr *MockGet) NewScoped(websiteID, storeID int64) config.ScopedGetter {
	return config.NewScopedService(mr, websiteID, storeID)
}

// From html/template/content.go
// Copyright 2011 The Go Authors. All rights reserved.
// indirect returns the value, after dereferencing as many times
// as necessary to reach the base type (or nil).
func indirect(a interface{}) interface{} {
	if a == nil {
		return nil
	}
	if t := reflect.TypeOf(a); t.Kind() != reflect.Ptr {
		// Avoid creating a reflect.Value if it's not a pointer.
		return a
	}
	v := reflect.ValueOf(a)
	for v.Kind() == reflect.Ptr && !v.IsNil() {
		v = v.Elem()
	}
	return v.Interface()
}