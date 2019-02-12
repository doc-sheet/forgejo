// Copyright 2019 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package lastcommit

import (
	"encoding/json"
	"fmt"

	"code.gitea.io/git"

	bolt "go.etcd.io/bbolt"
)

var (
	_ git.LastCommitCache = &LastCommitBoltDBCache{}
)

// LastCommitBoltDBCache implements git.LastCommitCache interface to save the last commits on leveldb
type LastCommitBoltDBCache struct {
	cacheDir string
	bucket   []byte
	db       *bolt.DB
}

// NewLastCommitBoltDBCache creates a boltdb cache
func NewLastCommitBoltDBCache(cacheDir string) (*LastCommitBoltDBCache, error) {
	db, err := bolt.Open(cacheDir, 0600, nil)
	if err != nil {
		return nil, err
	}

	var bucket = []byte("default")
	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(bucket)
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &LastCommitBoltDBCache{
		cacheDir: cacheDir,
		bucket:   bucket,
		db:       db,
	}, nil
}

func (c *LastCommitBoltDBCache) Get(repoPath, ref, entryPath string) (*git.Commit, error) {
	var commit git.Commit
	var found bool
	err := c.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(c.bucket)
		v := b.Get([]byte(getKey(repoPath, ref, entryPath)))
		if v == nil || len(v) <= 0 {
			return nil
		}
		found = true
		return json.Unmarshal(v, &commit)
	})
	if err != nil {
		return nil, err
	}
	if found {
		return &commit, nil
	}
	return nil, nil
}

func (c *LastCommitBoltDBCache) Put(repoPath, ref, entryPath string, commit *git.Commit) error {
	err := c.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(c.bucket)
		v, err := json.Marshal(commit)
		if err != nil {
			return err
		}
		return b.Put([]byte(getKey(repoPath, ref, entryPath)), v)
	})
	return err
}
