package context

import (
	"context"
)

type Context interface {
	context.Context
}

type instancedContext struct {
	Context
}

var background = &instancedContext{
	Context: context.Background(),
}

func Background() Context {
	return background
}

type stringMapContext struct {
	context.Context
	vals map[string]interface{}
}

func WithValues(ctx context.Context, vals map[string]interface{}) context.Context {
	nvals := make(map[string]interface{}, len(vals))
	for k, v := range vals {
		nvals[k] = v
	}

	return stringMapContext{
		Context: ctx,
		vals:    nvals,
	}
}

func WithValue(parent Context, key interface{}, val interface{}) Context {
	return context.WithValue(parent, key, val)
}

func (ctx stringMapContext) Value(key interface{}) interface{} {
	if ks, ok := key.(string); ok {
		if v, ok := ctx.vals[ks]; ok {
			return v
		}
	}

	return ctx.Context.Value(key)
}
