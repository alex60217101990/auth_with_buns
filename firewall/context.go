package firewall

import "context"

type key int

const IP key = 0

func newContextWithIP(ctx context.Context, ip string) context.Context {
	return context.WithValue(ctx, IP, ip)
}

func requestIPFromContext(ctx context.Context) string {
	return ctx.Value(IP).(string)
}
