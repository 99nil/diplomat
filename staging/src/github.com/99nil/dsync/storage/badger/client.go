// Copyright © 2021 zc2638 <zc2638@qq.com>.
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

package badger

import (
	"bytes"
	"context"
	"errors"

	"github.com/99nil/dsync/storage"

	"github.com/dgraph-io/badger/v3"
)

var _ storage.Interface = (*Client)(nil)

type Config struct {
	Path string `json:"path"`
}

type Client struct {
	db *badger.DB
}

func New(cfg *Config) (*Client, error) {
	var err error
	options := badger.DefaultOptions(cfg.Path)
	options.Logger = nil
	options.BypassLockGuard = true
	db, err := badger.Open(options)
	if err != nil {
		return nil, err
	}
	client := &Client{db: db}
	return client, nil
}

func NewWithDB(db *badger.DB) (*Client, error) {
	if db == nil {
		return nil, errors.New("db unavailable")
	}
	client := &Client{db: db}
	return client, nil
}

func (c *Client) test() {
	c.db.DropPrefix()
}

func (c *Client) Close() error {
	return c.db.Close()
}

func buildPrefix(space string) []byte {
	return append([]byte(space), '-')
}

func (c *Client) Get(_ context.Context, space, key string) ([]byte, error) {
	var res []byte
	err := c.db.View(func(txn *badger.Txn) error {
		prefix := buildPrefix(space)
		fmtKey := append(prefix, key...)
		item, err := txn.Get(fmtKey)
		if err != nil {
			if err == badger.ErrKeyNotFound {
				return nil
			}
			return err
		}
		return item.Value(func(v []byte) error {
			res = v
			return nil
		})
	})
	return res, err
}

func (c *Client) Add(_ context.Context, space, key string, value []byte) error {
	return c.db.Update(func(txn *badger.Txn) error {
		fmtKey := append(buildPrefix(space), key...)
		err := txn.Set(fmtKey, value)
		if err == badger.ErrKeyNotFound {
			return nil
		}
		return err
	})
}

func (c *Client) Del(_ context.Context, space, key string) error {
	return c.db.Update(func(txn *badger.Txn) error {
		fmtKey := append(buildPrefix(space), key...)
		err := txn.Delete(fmtKey)
		if err == badger.ErrKeyNotFound {
			return nil
		}
		return err
	})
}

func (c *Client) Clear(_ context.Context, space string) error {
	return c.db.DropPrefix([]byte(space))
}

func (c *Client) Range(_ context.Context, space string, fn func(key, value []byte) error) error {
	return c.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()

		prefix := buildPrefix(space)
		for it.Rewind(); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			k := bytes.TrimPrefix(item.Key(), prefix)
			err := item.Value(func(v []byte) error {
				return fn(k, v)
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
}
