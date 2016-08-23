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

package cfgmock

import (
	"fmt"
	"reflect"
	"sort"
	"sync/atomic"
	"time"

	"github.com/corestoreio/csfw/config"
	"github.com/corestoreio/csfw/config/cfgpath"
	"github.com/corestoreio/csfw/config/storage"
	"github.com/corestoreio/csfw/util/bufferpool"
	"github.com/corestoreio/csfw/util/conv"
	"github.com/corestoreio/csfw/util/errors"
)

// keyNotFound for performance and allocs reasons in benchmarks to test properly
// the cfg* code and not the configuration Service. The NotFound error has been
// hard coded which does not record the position where the error happens. We can
// maybe add the path which was not found but that will trigger 2 allocs because
// of the sprintf ... which could be bypassed with a bufferpool ;-)
type keyNotFound struct{}

func (a keyNotFound) Error() string  { return "[cfgmock] Get() Path" }
func (a keyNotFound) NotFound() bool { return true }

// Write used for testing when writing configuration values.
type Write struct {
	// WriteError gets always returned by Write
	WriteError error
	// ArgPath will be set after calling write to export the config path.
	// Values you enter here will be overwritten when calling Write
	ArgPath string
	// ArgValue contains the written data
	ArgValue interface{}
}

// Write writes to a black hole, may return an error
func (w *Write) Write(p cfgpath.Path, v interface{}) error {
	w.ArgPath = p.String()
	w.ArgValue = v
	return w.WriteError
}

// Service used for testing. Contains functions which will be called in the
// appropriate methods of interface config.Getter. Field DB has precedence over
// the applied functions.
type Service struct {
	DB               storage.Storager
	ByteFn           func(path string) ([]byte, error)
	byteFnInvokes    int32
	StringFn         func(path string) (string, error)
	stringFnInvokes  int32
	BoolFn           func(path string) (bool, error)
	boolFnInvokes    int32
	Float64Fn        func(path string) (float64, error)
	float64FnInvokes int32
	IntFn            func(path string) (int, error)
	intFnInvokes     int32
	TimeFn           func(path string) (time.Time, error)
	timeFnInvokes    int32
	SubscriptionID   int
	SubscriptionErr  error
}

// PathValue is a required type for an option function. PV = path => value. This
// map[string]interface{} is protected by a mutex.
type PathValue map[string]interface{}

func (pv PathValue) set(db storage.Storager) {
	for fq, v := range pv {
		p, err := cfgpath.SplitFQ(fq)
		if err != nil {
			panic(err)
		}
		if err := db.Set(p, v); err != nil {
			panic(err)
		}
	}
}

// GoString creates a sorted Go syntax valid map representation. This function
// panics if it fails to write to the internal buffer. Panicing permitted here
// because this function is only used in testing.
func (pv PathValue) GoString() string {
	keys := make(sort.StringSlice, len(pv))
	i := 0
	for k := range pv {
		keys[i] = k
		i++
	}
	keys.Sort()

	buf := bufferpool.Get()
	defer bufferpool.Put(buf)
	if _, err := buf.WriteString("cfgmock.PathValue{\n"); err != nil {
		panic(err)
	}

	for _, p := range keys {
		if _, err := fmt.Fprintf(buf, "%q: %#v,\n", p, pv[p]); err != nil {
			panic(err)
		}
	}
	if _, err := buf.WriteRune('}'); err != nil {
		panic(err)
	}
	return buf.String()
}

// NewService creates a new mocked Service for testing usage. Initializes a
// simple in memory key/value storage.
func NewService(pvs ...PathValue) *Service {
	mr := &Service{
		DB: storage.NewKV(),
	}
	if len(pvs) > 0 {
		for _, pv := range pvs {
			pv.set(mr.DB)
		}
	}
	return mr
}

// UpdateValues adds or overwrites the internal path => value map.
func (mr *Service) UpdateValues(pathValues PathValue) {
	pathValues.set(mr.DB)
}

func (mr *Service) hasVal(p cfgpath.Path) bool {
	if mr.DB == nil {
		return false
	}
	v, err := mr.DB.Get(p)
	if err != nil && !errors.IsNotFound(err) {
		println("Mock.Service.hasVal error:", err.Error(), "path", p.String())
	}
	return v != nil && err == nil
}

func (mr *Service) getVal(p cfgpath.Path) interface{} {
	v, err := mr.DB.Get(p)
	if err != nil && !errors.IsNotFound(err) {
		println("Mock.Service.getVal error:", err.Error(), "path", p.String())
		return nil
	}
	v = indirect(v)
	return v
}

// Byte returns a byte slice value
func (mr *Service) Byte(p cfgpath.Path) ([]byte, error) {
	switch {
	case mr.hasVal(p):
		return conv.ToByteE(mr.getVal(p))
	case mr.ByteFn != nil:
		atomic.AddInt32(&mr.byteFnInvokes, 1)
		return mr.ByteFn(p.String())
	default:
		return nil, keyNotFound{}
	}
}

// ByteFnInvokes returns the number of Byte() invocations.
func (mr *Service) ByteFnInvokes() int {
	return int(atomic.LoadInt32(&mr.byteFnInvokes))
}

// String returns a string value
func (mr *Service) String(p cfgpath.Path) (string, error) {
	switch {
	case mr.hasVal(p):
		return conv.ToStringE(mr.getVal(p))
	case mr.StringFn != nil:
		atomic.AddInt32(&mr.stringFnInvokes, 1)
		return mr.StringFn(p.String())
	default:
		return "", keyNotFound{}
	}
}

// StringFnInvokes returns the number of String() invocations.
func (mr *Service) StringFnInvokes() int {
	return int(atomic.LoadInt32(&mr.stringFnInvokes))
}

// Bool returns a bool value
func (mr *Service) Bool(p cfgpath.Path) (bool, error) {
	switch {
	case mr.hasVal(p):
		return conv.ToBoolE(mr.getVal(p))
	case mr.BoolFn != nil:
		atomic.AddInt32(&mr.boolFnInvokes, 1)
		return mr.BoolFn(p.String())
	default:
		return false, keyNotFound{}
	}
}

// BoolFnInvokes returns the number of Bool() invocations.
func (mr *Service) BoolFnInvokes() int {
	return int(atomic.LoadInt32(&mr.boolFnInvokes))
}

// Float64 returns a float64 value
func (mr *Service) Float64(p cfgpath.Path) (float64, error) {
	switch {
	case mr.hasVal(p):
		return conv.ToFloat64E(mr.getVal(p))
	case mr.Float64Fn != nil:
		atomic.AddInt32(&mr.float64FnInvokes, 1)
		return mr.Float64Fn(p.String())
	default:
		return 0.0, keyNotFound{}
	}
}

// Float64FnInvokes returns the number of Float64() invocations.
func (mr *Service) Float64FnInvokes() int {
	return int(atomic.LoadInt32(&mr.float64FnInvokes))
}

// Int returns an integer value
func (mr *Service) Int(p cfgpath.Path) (int, error) {
	switch {
	case mr.hasVal(p):
		return conv.ToIntE(mr.getVal(p))
	case mr.IntFn != nil:
		atomic.AddInt32(&mr.intFnInvokes, 1)
		return mr.IntFn(p.String())
	default:
		return 0, keyNotFound{}
	}
}

// IntFnInvokes returns the number of Int() invocations.
func (mr *Service) IntFnInvokes() int {
	return int(atomic.LoadInt32(&mr.intFnInvokes))
}

// Time returns a time value
func (mr *Service) Time(p cfgpath.Path) (time.Time, error) {
	switch {
	case mr.hasVal(p):
		return conv.ToTimeE(mr.getVal(p))
	case mr.TimeFn != nil:
		atomic.AddInt32(&mr.timeFnInvokes, 1)
		return mr.TimeFn(p.String())
	default:
		return time.Time{}, keyNotFound{}
	}
}

// TimeFnInvokes returns the number of Time() invocations.
func (mr *Service) TimeFnInvokes() int {
	return int(atomic.LoadInt32(&mr.timeFnInvokes))
}

// Subscribe returns the before applied SubscriptionID and SubscriptionErr
// Does not start any underlying Goroutines.
func (mr *Service) Subscribe(_ cfgpath.Route, s config.MessageReceiver) (subscriptionID int, err error) {
	return mr.SubscriptionID, mr.SubscriptionErr
}

// NewScoped creates a new config.ScopedReader which uses the underlying
// mocked paths and values.
func (mr *Service) NewScoped(websiteID, storeID int64) config.Scoped {
	return config.NewScoped(mr, websiteID, storeID)
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
