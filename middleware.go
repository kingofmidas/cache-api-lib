package cache

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
)

var ctx = context.Background()

type Tagger interface {
	Tag() string
}

type Middleware func(http.Handler, ...Tagger) http.Handler

func (c CacheClient) CacheMiddleware() Middleware {
	return c.cacheMiddleware
}

func (c CacheClient) cacheMiddleware(next http.Handler, tags ...Tagger) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		val, err := c.checkCache(r.URL.Path)
		if err == redis.Nil {
			logrus.Warnf("Key %s - not found", r.URL.Path)

			rec := httptest.NewRecorder()
			next.ServeHTTP(rec, r)
			val = rec.Body.Bytes()
			c.saveCache(r.URL.Path, val, tags...)
			w.Write(val)
			return
		} else if err != nil {
			logrus.Errorf("Redis error: %s", err.Error())
			next.ServeHTTP(w, r)
			return
		}
		logrus.Infof("Found %s key", r.URL.Path)
		w.Write(val)
	}

	return http.HandlerFunc(fn)
}

func (c CacheClient) checkCache(url string) ([]byte, error) {
	cacheKey := fmt.Sprintf("%s:cache:%s", c.appName, url)
	return c.redis.Get(ctx, cacheKey).Bytes()
}

func (c CacheClient) saveCache(url string, content []byte, tags ...Tagger) {
	cacheKey := fmt.Sprintf("%s:cache:%s", c.appName, url)
	c.redis.Set(ctx, cacheKey, content, c.expiration)

	for _, tag := range tags {
		tagKey := fmt.Sprintf("%s:tags:%s", c.appName, tag.Tag())
		c.redis.SAdd(ctx, tagKey, cacheKey)
	}
}

func (c CacheClient) InvalidateCache(tags ...Tagger) {
	for _, tag := range tags {
		tagKey := fmt.Sprintf("%s:tags:%s", c.appName, tag)
		keys, _ := c.redis.SMembers(ctx, tagKey).Result()
		for _, key := range keys {
			c.redis.Del(ctx, key)
		}
	}
}
