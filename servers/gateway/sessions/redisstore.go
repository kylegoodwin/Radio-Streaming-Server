package sessions

import (
	"encoding/json"
	"log"
	"time"

	"github.com/go-redis/redis"
)

//RedisStore represents a session.Store backed by redis.
type RedisStore struct {
	//Redis client used to talk to redis server.
	Client *redis.Client
	//Used for key expiry time on redis.
	SessionDuration time.Duration
}

//NewRedisStore constructs a new RedisStore
func NewRedisStore(client *redis.Client, sessionDuration time.Duration) *RedisStore {
	//initialize and return a new RedisStore struct
	if client != nil {
		return &RedisStore{
			Client:          client,
			SessionDuration: sessionDuration,
		}
	}
	return nil
}

//Store implementation

//Save saves the provided `sessionState` and associated SessionID to the store.
//The `sessionState` parameter is typically a pointer to a struct containing
//all the data you want to associated with the given SessionID.
func (rs *RedisStore) Save(sid SessionID, sessionState interface{}) error {
	//TODO: marshal the `sessionState` to JSON and save it in the redis database,
	//using `sid.getRedisKey()` for the key.
	//return any errors that occur along the way.
	marshaled, err := json.Marshal(sessionState)
	if err != nil {
		return err
	}
	log.Println("Trying to Save sid: " + sid.getRedisKey())
	err = rs.Client.Set(sid.getRedisKey(), marshaled, rs.SessionDuration).Err()
	if err != nil {
		if err == redis.Nil {
			return nil
		}
		return err
	}

	return nil
}

//Get populates `sessionState` with the data previously saved
//for the given SessionID
func (rs *RedisStore) Get(sid SessionID, sessionState interface{}) error {
	//TODO: get the previously-saved session state data from redis,
	//unmarshal it back into the `sessionState` parameter
	//and reset the expiry time, so that it doesn't get deleted until
	//the SessionDuration has elapsed.
	got := rs.Client.Get(sid.getRedisKey())
	if got.Val() == "" {
		log.Println("Did not find that sid")
		return ErrStateNotFound
	}

	err := json.Unmarshal([]byte(got.Val()), sessionState)
	if err != nil {
		return err
	}

	rs.Client.Expire(sid.getRedisKey(), rs.SessionDuration)
	//for extra-credit using the Pipeline feature of the redis
	//package to do both the get and the reset of the expiry time
	//in just one network round trip!

	//pipe := rs.Pipeline()

	return nil
}

//Delete deletes all state data associated with the SessionID from the store.
func (rs *RedisStore) Delete(sid SessionID) error {
	//TODO: delete the data stored in redis for the provided SessionID
	deleted := rs.Client.Del(sid.getRedisKey())
	if deleted.Val() == 0 {
		log.Println("Unable to delete given sid")
		return ErrStateNotFound
	}
	return nil
}

//getRedisKey() returns the redis key to use for the SessionID
func (sid SessionID) getRedisKey() string {
	//convert the SessionID to a string and add the prefix "sid:" to keep
	//SessionID keys separate from other keys that might end up in this
	//redis instance
	return "sid:" + sid.String()
}
