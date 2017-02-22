//
// DISCLAIMER
//
// Copyright 2017 ArangoDB GmbH, Cologne, Germany
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// Copyright holder is ArangoDB GmbH, Cologne, Germany
//
// Author Ewout Prangsma
//

package driver

import (
	"context"
	"strconv"
)

const (
	keyRevision    = "arangodb-revision"
	keyReturnNew   = "arangodb-returnNew"
	keyReturnOld   = "arangodb-returnOld"
	keySilent      = "arangodb-silent"
	keyWaitForSync = "arangodb-waitForSync"
	keyDetails     = "arangodb-details"
)

// WithRevision is used to configure a context to make document
// functions specify an explicit revision of the document using an `If-Match` condition.
func WithRevision(parent context.Context, revision string) context.Context {
	return context.WithValue(parent, keyRevision, revision)
}

// WithReturnNew is used to configure a context to make create, update & replace document
// functions return the new document into the given result.
func WithReturnNew(parent context.Context, result interface{}) context.Context {
	return context.WithValue(parent, keyReturnNew, result)
}

// WithReturnOld is used to configure a context to make update & replace document
// functions return the old document into the given result.
func WithReturnOld(parent context.Context, result interface{}) context.Context {
	return context.WithValue(parent, keyReturnOld, result)
}

// WithDetails is used to configure a context to make Client.Version return additional details.
// You can pass a single (optional) boolean. If that is set to false, you explicitly ask to not provide details.
func WithDetails(parent context.Context, value ...bool) context.Context {
	v := true
	if len(value) == 1 {
		v = value[0]
	}
	return context.WithValue(parent, keyDetails, v)
}

// WithSilent is used to configure a context to make functions return an empty result (silent==true),
// instead of a metadata result (silent==false, default).
// You can pass a single (optional) boolean. If that is set to false, you explicitly ask to return metadata result.
func WithSilent(parent context.Context, value ...bool) context.Context {
	v := true
	if len(value) == 1 {
		v = value[0]
	}
	return context.WithValue(parent, keySilent, v)
}

// WithWaitForSync is used to configure a context to make modification
// functions wait until the data has been synced to disk (or not).
// You can pass a single (optional) boolean. If that is set to false, you explicitly do not wait for
// data to be synced to disk.
func WithWaitForSync(parent context.Context, value ...bool) context.Context {
	v := true
	if len(value) == 1 {
		v = value[0]
	}
	return context.WithValue(parent, keyWaitForSync, v)
}

type contextSettings struct {
	Silent      bool
	WaitForSync bool
	ReturnOld   interface{}
	ReturnNew   interface{}
	Revision    string
}

// applyContextSettings returns the settings configured in the context in the given request.
// It then returns information about the applied settings that may be needed later in API implementation functions.
func applyContextSettings(ctx context.Context, req Request) contextSettings {
	result := contextSettings{}
	if ctx == nil {
		return result
	}
	// Details
	if v := ctx.Value(keyDetails); v != nil {
		if details, ok := v.(bool); ok {
			req.SetQuery("details", strconv.FormatBool(details))
		}
	}
	// Silent
	if v := ctx.Value(keySilent); v != nil {
		if silent, ok := v.(bool); ok {
			req.SetQuery("silent", strconv.FormatBool(silent))
			result.Silent = silent
		}
	}
	// WaitForSync
	if v := ctx.Value(keyWaitForSync); v != nil {
		if waitForSync, ok := v.(bool); ok {
			req.SetQuery("waitForSync", strconv.FormatBool(waitForSync))
			result.WaitForSync = waitForSync
		}
	}
	// ReturnOld
	if v := ctx.Value(keyReturnOld); v != nil {
		req.SetQuery("returnOld", "true")
		result.ReturnOld = v
	}
	// ReturnNew
	if v := ctx.Value(keyReturnNew); v != nil {
		req.SetQuery("returnNew", "true")
		result.ReturnNew = v
	}
	// If-Match
	if v := ctx.Value(keyRevision); v != nil {
		if rev, ok := v.(string); ok {
			req.SetHeader("If-Match", rev)
			result.Revision = rev
		}
	}
	return result
}

// okStatus returns one of the given status codes depending on the WaitForSync field value.
// If WaitForSync==true, statusWithWaitForSync is returned, otherwise statusWithoutWaitForSync is returned.
func (cs contextSettings) okStatus(statusWithWaitForSync, statusWithoutWaitForSync int) int {
	if cs.WaitForSync {
		return statusWithWaitForSync
	} else {
		return statusWithoutWaitForSync
	}
}
