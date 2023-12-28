package postfeedcache

import (
	"context"
	"errors"
	"fmt"

	"github.com/redis/go-redis/v9"
	"myfacebook/internal/rdb"
)

const (
	maxListLen                         = 1000
	postFeedCachePrefix                = "postfeed:user_"
	postFeedLastRetrievedAtCachePrefix = "postfeed:last_retrieved_at:user_"
)

type Cache struct {
	redisDB *rdb.RedisDB
}

func New(redisDB *rdb.RedisDB) *Cache {
	return &Cache{
		redisDB: redisDB,
	}
}

func (c *Cache) AddPostID(ctx context.Context, key string, value string) error {
	_, err := c.redisDB.GetClient().LPush(ctx, postFeedCachePrefix+key, value).Result()
	if err != nil {
		return fmt.Errorf("postfeedcache failed to push value for key %q: %w", key, err)
	}

	_, err = c.redisDB.GetClient().LTrim(ctx, key, 0, maxListLen-1).Result()
	if err != nil {
		return fmt.Errorf("postfeedcache failed to trim list size: %w", err)
	}

	return nil
}

func (c *Cache) RemovePostID(ctx context.Context, key string, value string) error {
	_, err := c.redisDB.GetClient().LRem(ctx, postFeedCachePrefix+key, 0, value).Result()
	if err != nil {
		return fmt.Errorf("postfeedcache failed to remove value from the list for key %q: %w", key, err)
	}

	return nil
}

func (c *Cache) GetPostsIDs(ctx context.Context, key string) ([]string, error) {
	values, err := c.redisDB.GetClient().LRange(ctx, postFeedCachePrefix+key, 0, -1).Result()
	if err != nil {
		return nil, fmt.Errorf("postfeedcache failed to fetch elements for key %q: %w", key, err)
	}

	return values, nil
}

func (c *Cache) SetLastRetrievedAt(ctx context.Context, key string, lastRetrievedAtTimestamp int64) error {
	_, err := c.redisDB.GetClient().Set(ctx, postFeedLastRetrievedAtCachePrefix+key, lastRetrievedAtTimestamp, 0).Result()
	if err != nil {
		return fmt.Errorf("postfeedcache failed to set last retrieved timestamp: %w", err)
	}

	return nil
}

func (c *Cache) GetLastRetrievedAt(ctx context.Context, key string) (int64, error) {
	lastRetrievedAtTimestampMilli, err := c.redisDB.GetClient().Get(ctx, postFeedLastRetrievedAtCachePrefix+key).Int64()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return 0, nil
		}

		return 0, fmt.Errorf("postfeedcache failed to get last retrieved timestamp: %w", err)
	}

	return lastRetrievedAtTimestampMilli, nil
}
