//
// DISCLAIMER
//
// Copyright 2023 ArangoDB GmbH, Cologne, Germany
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

package arangodb

import (
	"context"
	"io"
	"net/http"

	"github.com/arangodb/go-driver/v2/arangodb/shared"
	"github.com/arangodb/go-driver/v2/connection"
)

func newCollectionDocumentDelete(collection *collection) *collectionDocumentDelete {
	return &collectionDocumentDelete{
		collection: collection,
	}
}

var _ CollectionDocumentDelete = &collectionDocumentDelete{}

type collectionDocumentDelete struct {
	collection *collection
}

func (c collectionDocumentDelete) DeleteDocument(ctx context.Context, key string) (CollectionDocumentDeleteResponse, error) {
	return c.DeleteDocumentWithOptions(ctx, key, nil)
}

func (c collectionDocumentDelete) DeleteDocumentWithOptions(ctx context.Context, key string, opts *CollectionDocumentDeleteOptions) (CollectionDocumentDeleteResponse, error) {
	url := c.collection.url("document", key)

	var meta CollectionDocumentDeleteResponse

	resp, err := connection.CallDelete(ctx, c.collection.connection(), url, &meta, c.collection.withModifiers(opts.modifyRequest)...)
	if err != nil {
		return CollectionDocumentDeleteResponse{}, err
	}

	switch code := resp.Code(); code {
	case http.StatusOK, http.StatusAccepted:
		return meta, nil
	default:
		return CollectionDocumentDeleteResponse{}, meta.AsArangoErrorWithCode(code)
	}
}

func (c collectionDocumentDelete) DeleteDocuments(ctx context.Context, keys []string) (CollectionDocumentDeleteResponseReader, error) {
	return c.DeleteDocumentsWithOptions(ctx, keys, nil)
}

func (c collectionDocumentDelete) DeleteDocumentsWithOptions(ctx context.Context, keys []string, opts *CollectionDocumentDeleteOptions) (CollectionDocumentDeleteResponseReader, error) {
	url := c.collection.url("document")

	req, err := c.collection.connection().NewRequest(http.MethodDelete, url)
	if err != nil {
		return nil, err
	}

	for _, modifier := range c.collection.withModifiers(opts.modifyRequest, connection.WithBody(keys)) {
		if err = modifier(req); err != nil {
			return nil, err
		}
	}

	var arr connection.Array

	resp, err := c.collection.connection().Do(ctx, req, &arr)
	if err != nil {
		return nil, err
	}

	switch code := resp.Code(); code {
	case http.StatusOK, http.StatusAccepted:
		return newCollectionDocumentDeleteResponseReader(arr, opts), nil
	default:
		return nil, shared.NewResponseStruct().AsArangoErrorWithCode(code)
	}
}

func newCollectionDocumentDeleteResponseReader(array connection.Array, options *CollectionDocumentDeleteOptions) *collectionDocumentDeleteResponseReader {
	c := &collectionDocumentDeleteResponseReader{array: array, options: options}

	return c
}

var _ CollectionDocumentDeleteResponseReader = &collectionDocumentDeleteResponseReader{}

type collectionDocumentDeleteResponseReader struct {
	array   connection.Array
	options *CollectionDocumentDeleteOptions
}

func (c *collectionDocumentDeleteResponseReader) Read(i interface{}) (CollectionDocumentDeleteResponse, error) {
	if !c.array.More() {
		return CollectionDocumentDeleteResponse{}, shared.NoMoreDocumentsError{}
	}

	var meta CollectionDocumentDeleteResponse

	if err := c.array.Unmarshal(newMultiUnmarshaller(&meta, newUnmarshalInto(i))); err != nil {
		if err == io.EOF {
			return CollectionDocumentDeleteResponse{}, shared.NoMoreDocumentsError{}
		}
		return CollectionDocumentDeleteResponse{}, err
	}

	return meta, nil
}
