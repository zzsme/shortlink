package main

import (
	"crypto/sha1"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/go-redis/redis"
	"github.com/mattheath/base62"
)

const (
	// URLIDKEY is global counter
	URLIDKEY = "next.url.id"
	// ShortlinkKey mapping the shortlink to the url
	ShortlinkKey = "shortlink:%s:url"
	// URLHashKey mapping the hash of the url to shortlink
	URLHashKey = "urlhash:%s:url"
	// ShortlinkDetailKey mapping the shortlink to the detail of url
	ShortlinkDetailKey = "shortlink:%s:detail"
)

// RedisCli contains a redis Client
type RedisCli struct {
	Cli *redis.Client
}

// URLDetail contains the detail of the shortlink
type URLDetail struct {
	URL                 string        `json:"url"`
	CreatedAt           string        `json:"created_at"`
	ExpirationInMinutes time.Duration `json:"expiration_in_minutes"`
}

// NewRedisCli create a redis client
func NewRedisCli(addr string, passwd string, db int) *RedisCli {
	c := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: passwd,
		DB:       db,
	})

	if _, err := c.Ping().Result(); err != nil {
		panic(err)
	}

	return &RedisCli{Cli: c}
}

// Shorten convert url to shortlink
func (r *RedisCli) Shorten(url string, exp int64) (string, error) {
	h := toSha1(url)
	d, err := r.Cli.Get(fmt.Sprintf(URLHashKey, h)).Result()
	if err != nil {
		return "", err
	}
	if d == "{}" {
		// expiration, nothing to do
	} else {
		return d, nil
	}

	// increase the global counter
	err = r.Cli.Incr(URLIDKEY).Err()
	if err != nil {
		return "", err
	}

	// encode global counter to base62
	id, err := r.Cli.Get(URLIDKEY).Int64()
	if err != nil {
		return "", err
	}

	eid := base62.EncodeInt64(id)
	//store the url against this encoded id
	err = r.Cli.Set(fmt.Sprintf(ShortlinkKey, eid), url, time.Minute*time.Duration(exp)).Err()
	if err != nil {
		return "", err
	}

	// store the url against the hash url
	err = r.Cli.Set(fmt.Sprintf(URLHashKey, h), eid, time.Minute*time.Duration(exp)).Err()
	if err != nil {
		return "", err
	}
	detail, err := json.Marshal(
		&URLDetail{
			URL:                 url,
			CreatedAt:           time.Now().String(),
			ExpirationInMinutes: time.Duration(exp)})

	if err != nil {
		return "", err
	}

	err = r.Cli.Set(fmt.Sprintf(ShortlinkDetailKey, eid), detail, time.Minute*time.Duration(exp)).Err()

	if err != nil {
		return "", err
	}

	return eid, nil
}

// ShortlinkInfo returns the detail of the shortlink
func (r *RedisCli) ShortlinkInfo(eid string) (interface{}, error) {
	d, err := r.Cli.Get(fmt.Sprintf(ShortlinkDetailKey, eid)).Result()
	if err == redis.Nil {
		return "", StatusError{404, errors.New("Unknow short URL")}
	} else if err != nil {
		return "", err
	} else {
		return d, nil
	}
}

// Unshorten convert shortlink to url
func (r *RedisCli) Unshorten(eid string) (string, error) {
	url, err := r.Cli.Get(fmt.Sprintf(ShortlinkKey, eid)).Result()
	if err == redis.Nil {
		return "", StatusError{404, err}
	} else if err != nil {
		return "", err
	} else {
		return url, nil
	}
}

func toSha1(s string) string {

	h := sha1.New()
	h.Write([]byte(s))

	bs := h.Sum(nil)

	return string(bs)
}
