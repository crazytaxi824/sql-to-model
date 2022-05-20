package db

import (
	"context"
	"fmt"
	"time"

	"github.com/uptrace/bun"
)

// databse debug 用
type queryHook struct{}

func (*queryHook) BeforeQuery(ctx context.Context, _ *bun.QueryEvent) context.Context {
	return ctx
}

func (*queryHook) AfterQuery(_ context.Context, event *bun.QueryEvent) {
	fmt.Println(time.Since(event.StartTime), event.Query)
}
