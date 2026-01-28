package api

import (
	"context"
	"net/url"
)

type metaOverridesKey struct{}

type MetaOverrides struct {
	Params map[string]string
}

func WithMetaOverrides(ctx context.Context, params map[string]string) context.Context {
	if len(params) == 0 {
		return ctx
	}
	return context.WithValue(ctx, metaOverridesKey{}, MetaOverrides{Params: params})
}

func MetaOverridesFromContext(ctx context.Context) (MetaOverrides, bool) {
	value := ctx.Value(metaOverridesKey{})
	if value == nil {
		return MetaOverrides{}, false
	}
	overrides, ok := value.(MetaOverrides)
	return overrides, ok
}

func ApplyMetaOverrides(ctx context.Context, query url.Values) {
	overrides, ok := MetaOverridesFromContext(ctx)
	if !ok || len(overrides.Params) == 0 {
		return
	}
	for key, value := range overrides.Params {
		query.Set(key, value)
	}
}
