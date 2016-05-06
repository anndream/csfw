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

package ctxjwt

import (
	"testing"

	"context"

	"github.com/corestoreio/csfw/util/csjwt"
	"github.com/corestoreio/csfw/util/errors"
	"github.com/stretchr/testify/assert"
)

func TestContextWithError(t *testing.T) {

	const wantErr = errors.UserNotFound("User Contiki not found")
	ctx := withContextError(context.Background(), wantErr)
	assert.NotNil(t, ctx)

	haveToken, haveErr := FromContext(ctx)
	assert.NotNil(t, haveToken)
	assert.False(t, haveToken.Valid)
	assert.True(t, errors.IsUserNotFound(haveErr))
}

func TestFromContext(t *testing.T) {

	ctx := withContext(context.Background(), csjwt.Token{})
	assert.NotNil(t, ctx)

	haveToken, haveErr := FromContext(ctx)
	assert.NotNil(t, haveToken)
	assert.False(t, haveToken.Valid)
	assert.True(t, errors.IsNotFound(haveErr))
}
