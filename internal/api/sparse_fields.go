package api

import (
	"context"
	"net/url"
	"strings"
)

type sparseFieldsKey struct{}

type SparseFieldOverrides struct {
	FieldsSet  bool
	Primary    []string
	Typed      map[string][]string
	IncludeSet bool
	Include    []string
}

func WithSparseFieldOverrides(ctx context.Context, overrides SparseFieldOverrides) context.Context {
	return context.WithValue(ctx, sparseFieldsKey{}, overrides)
}

func SparseFieldOverridesFromContext(ctx context.Context) (SparseFieldOverrides, bool) {
	value := ctx.Value(sparseFieldsKey{})
	if value == nil {
		return SparseFieldOverrides{}, false
	}
	overrides, ok := value.(SparseFieldOverrides)
	return overrides, ok
}

func ApplySparseFieldOverrides(ctx context.Context, path string, query url.Values) {
	overrides, ok := SparseFieldOverridesFromContext(ctx)
	if !ok {
		return
	}

	if overrides.FieldsSet {
		primaryType := resourceTypeFromPath(path)
		if primaryType != "" && len(overrides.Primary) > 0 {
			if overrides.Typed == nil || len(overrides.Typed[primaryType]) == 0 {
				query.Set("fields["+primaryType+"]", strings.Join(overrides.Primary, ","))
			}
		}
		for resourceType, fields := range overrides.Typed {
			if len(fields) == 0 {
				continue
			}
			query.Set("fields["+resourceType+"]", strings.Join(fields, ","))
		}
	}

	if overrides.IncludeSet {
		if len(overrides.Include) == 0 {
			query.Del("include")
		} else {
			query.Set("include", strings.Join(overrides.Include, ","))
		}
	}
}

func resourceTypeFromPath(path string) string {
	trimmed := strings.Trim(path, "/")
	if trimmed == "" {
		return ""
	}
	parts := strings.Split(trimmed, "/")
	if len(parts) >= 2 && parts[0] == "v1" {
		return parts[1]
	}
	return parts[0]
}
