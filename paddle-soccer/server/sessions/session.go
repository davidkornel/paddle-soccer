// Copyright 2017 Google Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package sessions

import (
	"log"

	"github.com/garyburd/redigo/redis"
	"github.com/pkg/errors"
)

// Session represents a game session
type Session struct {
	ID   string `json:"id" redis:"id"`
	Port int    `json:"port,omitempty" redis:"port"`
	IP   string `json:"ip,omitempty" redis:"ip"`
}

const redisSessionPrefix = "Session:"

// ErrSessionNotFound is returned when the Session can't be
// found in redis
var ErrSessionNotFound = errors.New("Could not find the requested session")

// storeSession store the session in redis
func (s *Server) storeSession(sess Session) error {
	log.Print("[Info][session] Storing session in redis")
	con := s.pool.Get()
	defer con.Close()

	key := redisSessionPrefix + sess.ID

	err := con.Send("MULTI")
	if err != nil {
		return errors.Wrap(err, "Could not Send MULTI")
	}

	err = con.Send("HMSET", key, "id", sess.ID,
		"port", sess.Port,
		"ip", sess.IP)
	if err != nil {
		return errors.Wrap(err, "Could not Send HMSET")
	}

	err = con.Send("EXPIRE", key, 60*60)
	if err != nil {
		return errors.Wrap(err, "Could not Send EXPIRE")
	}

	_, err = con.Do("EXEC")
	return errors.Wrap(err, "Could not save session to redis")
}

// getSession returns a Session from redis
func (s *Server) getSession(id string) (Session, error) {
	con := s.pool.Get()
	defer con.Close()

	key := redisSessionPrefix + id

	log.Printf("[Info][session_type] Getting data for key: %v", key)

	var result Session
	values, err := redis.Values(con.Do("HGETALL", key))
	if err != nil {
		return result, errors.Wrapf(err, "Error getting hash for key %v", key)
	}

	if len(values) == 0 {
		log.Printf("[Warn][session] Could not find record for key: %v", key)
		return result, ErrSessionNotFound
	}

	err = redis.ScanStruct(values, &result)
	return result, errors.Wrap(err, "Error Scanning Struct")
}
