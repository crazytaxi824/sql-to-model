package util

import (
	"context"
	"fmt"
	"time"

	"github.com/uptrace/bun"
)

// databse debug 用
type QueryHook struct{}

func (*QueryHook) BeforeQuery(ctx context.Context, _ *bun.QueryEvent) context.Context {
	return ctx
}

func (*QueryHook) AfterQuery(_ context.Context, event *bun.QueryEvent) {
	fmt.Println(time.Since(event.StartTime), event.Query)
}
