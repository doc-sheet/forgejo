// Copyright 2019 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package session

import (
	"log"
	"sync"

	"code.gitea.io/gitea/models"
	"code.gitea.io/gitea/modules/timeutil"
	"gitea.com/macaron/session"
)

// XormlStore represents a xorm session store implementation.
type XormStore struct {
	sid  string
	lock sync.RWMutex
	data map[interface{}]interface{}
}

// NewXormStore creates and returns a Xorm session store.
func NewXormStore(sid string, kv map[interface{}]interface{}) *XormStore {
	return &XormStore{
		sid:  sid,
		data: kv,
	}
}

// Set sets value to given key in session.
func (s *XormStore) Set(key, val interface{}) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.data[key] = val
	return nil
}

// Get gets value by given key in session.
func (s *XormStore) Get(key interface{}) interface{} {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return s.data[key]
}

// Delete delete a key from session.
func (s *XormStore) Delete(key interface{}) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	delete(s.data, key)
	return nil
}

// ID returns current session ID.
func (s *XormStore) ID() string {
	return s.sid
}

// Release releases resource and save data to provider.
func (s *XormStore) Release() error {
	// Skip encoding if the data is empty
	if len(s.data) == 0 {
		return nil
	}

	data, err := session.EncodeGob(s.data)
	if err != nil {
		return err
	}

	return models.UpdateSession(s.sid, data)
}

// Flush deletes all session data.
func (s *XormStore) Flush() error {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.data = make(map[interface{}]interface{})
	return nil
}

// XormProvider represents a Xorm session provider implementation.
type XormProvider struct {
	maxLifetime int64
}

// Init initializes Xorm session provider.
// connStr: username:password@protocol(address)/dbname?param=value
func (p *XormProvider) Init(maxLifetime int64, connStr string) error {
	p.maxLifetime = maxLifetime
	return nil
}

// Read returns raw session store by session ID.
func (p *XormProvider) Read(sid string) (session.RawStore, error) {
	s, err := models.ReadSession(sid)
	if err != nil {
		return nil, err
	}

	var kv map[interface{}]interface{}
	if len(s.Data) == 0 || s.Expiry.Add(p.maxLifetime) <= timeutil.TimeStampNow() {
		kv = make(map[interface{}]interface{})
	} else {
		kv, err = session.DecodeGob(s.Data)
		if err != nil {
			return nil, err
		}
	}

	return NewXormStore(sid, kv), nil
}

// Exist returns true if session with given ID exists.
func (p *XormProvider) Exist(sid string) bool {
	has, err := models.ExistSession(sid)
	if err != nil {
		panic("session/Xorm: error checking existence: " + err.Error())
	}
	return has
}

// Destroy deletes a session by session ID.
func (p *XormProvider) Destroy(sid string) error {
	return models.DestroySession(sid)
}

// Regenerate regenerates a session store from old session ID to new one.
func (p *XormProvider) Regenerate(oldsid, sid string) (_ session.RawStore, err error) {
	s, err := models.RegenerateSession(oldsid, sid)
	if err != nil {
		return nil, err

	}

	var kv map[interface{}]interface{}
	if len(s.Data) == 0 || s.Expiry.Add(p.maxLifetime) <= timeutil.TimeStampNow() {
		kv = make(map[interface{}]interface{})
	} else {
		kv, err = session.DecodeGob(s.Data)
		if err != nil {
			return nil, err
		}
	}

	return NewXormStore(sid, kv), nil
}

// Count counts and returns number of sessions.
func (p *XormProvider) Count() int {
	total, err := models.CountSessions()
	if err != nil {
		panic("session/Xorm: error counting records: " + err.Error())
	}
	return int(total)
}

// GC calls GC to clean expired sessions.
func (p *XormProvider) GC() {
	if err := models.CleanupSessions(p.maxLifetime); err != nil {
		log.Printf("session/Xorm: error garbage collecting: %v", err)
	}
}

func init() {
	session.Register("Xorm", &XormProvider{})
}
