/*
Copyright 2026 Pedro Cozinheiro.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package testhelper

import (
	"context"

	goredis "github.com/redis/go-redis/v9"
)

// FakeRedisClient is a test double for Redis client behavior.
//
// It implements only the subset of methods required by the application,
// allowing full control over Redis interactions in tests.
type FakeRedisClient struct {
	XGroupCreateMkStreamFunc func(ctx context.Context, stream, group, start string) *goredis.StatusCmd
	XReadGroupFunc           func(ctx context.Context, args *goredis.XReadGroupArgs) *goredis.XStreamSliceCmd
	XAckFunc                 func(ctx context.Context, stream, group string, ids ...string) *goredis.IntCmd
}

// XGroupCreateMkStream simulates consumer group creation.
func (f *FakeRedisClient) XGroupCreateMkStream(
	ctx context.Context,
	stream, group, start string,
) *goredis.StatusCmd {

	if f.XGroupCreateMkStreamFunc != nil {
		return f.XGroupCreateMkStreamFunc(ctx, stream, group, start)
	}

	return goredis.NewStatusCmd(ctx)
}

// XReadGroup simulates reading from a Redis stream.
func (f *FakeRedisClient) XReadGroup(
	ctx context.Context,
	args *goredis.XReadGroupArgs,
) *goredis.XStreamSliceCmd {

	if f.XReadGroupFunc != nil {
		return f.XReadGroupFunc(ctx, args)
	}

	return goredis.NewXStreamSliceCmd(ctx)
}

// XAck simulates acknowledging a message.
func (f *FakeRedisClient) XAck(
	ctx context.Context,
	stream, group string,
	ids ...string,
) *goredis.IntCmd {

	if f.XAckFunc != nil {
		return f.XAckFunc(ctx, stream, group, ids...)
	}

	return goredis.NewIntCmd(ctx)
}
