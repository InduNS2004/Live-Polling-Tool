package realtime

import (
	"context"
	"encoding/json"
	"github.com/redis/go-redis/v9"
	"live-polling-tool/backend/models"
	"net/url"
	"strings"
	"time"
)

type Redis struct{ Client *redis.Client }

func NewRedis(raw string) (*Redis, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return nil, err
	}
	opts := &redis.Options{Addr: u.Host, Password: "", DB: 0}
	if u.User != nil {
		opts.Username = u.User.Username()
		opts.Password, _ = u.User.Password()
	}
	if strings.TrimPrefix(u.Path, "/") != "" { /* DB intentionally stays 0 for predictable managed-Redis defaults. */
	}
	return &Redis{Client: redis.NewClient(opts)}, nil
}
func (r *Redis) Ping(ctx context.Context) error { return r.Client.Ping(ctx).Err() }
func (r *Redis) PublishResult(ctx context.Context, res models.Result) error {
	b, err := json.Marshal(res)
	if err != nil {
		return err
	}
	key := "poll:results:" + res.PollID
	pipe := r.Client.Pipeline()
	pipe.Set(ctx, key, b, 30*time.Minute)
	pipe.Publish(ctx, "poll:"+res.PollID, b)
	_, err = pipe.Exec(ctx)
	return err
}
func (r *Redis) CachedResult(ctx context.Context, pollID string) (*models.Result, error) {
	b, err := r.Client.Get(ctx, "poll:results:"+pollID).Bytes()
	if err != nil {
		return nil, err
	}
	var res models.Result
	if err = json.Unmarshal(b, &res); err != nil {
		return nil, err
	}
	return &res, nil
}
func (r *Redis) Subscribe(ctx context.Context, pollID string) *redis.PubSub {
	return r.Client.Subscribe(ctx, "poll:"+pollID)
}
