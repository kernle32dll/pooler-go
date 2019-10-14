package pooler

import (
	"context"
	"errors"
	"net/http"
	"sync"
)

var (
	// holderPool is the internal pool, used to initialize
	// poolHolder objects used per request.
	holderPool = &sync.Pool{New: func() interface{} {
		return &poolHolder{
			takenObjects: map[interface{}]struct{}{},
		}
	}}

	// ErrMissingContext is used as a panic object for when Get is
	// used with an context, which does not contain the requested
	// holder key.
	ErrMissingContext = errors.New("pooler middleware missing in request context")
)

// poolHolder is the internal structure used per request
// to keep track of objects which were retrieved from the
// underlying pool.
type poolHolder struct {
	pool         *sync.Pool
	takenObjects map[interface{}]struct{}
}

// cleanup returns all used objects back to the underlying pool.
func (poolHolder *poolHolder) cleanup() {
	for obj := range poolHolder.takenObjects {
		poolHolder.pool.Put(obj)
	}
}

// NewMiddleware initializes a new middleware for providing object pooling capabilities.
// This middleware injects an object pool into the request context via the given key. This pool can
// be accessed with this packages Get method, and the same key. Returning objects is not required,
// as they are returned after the request ended.
func NewMiddleware(key interface{}, factoryMethod func() interface{}) func(h http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		objPool := &sync.Pool{New: factoryMethod}

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			holder := holderPool.Get().(*poolHolder)
			holder.pool = objPool

			c := context.WithValue(r.Context(), key, holder)

			defer func() {
				// Recover any panic, if it happened
				r := recover()

				holder.cleanup()
				holderPool.Put(holder)

				// We are done here - pass the panic up
				if r != nil {
					panic(r)
				}
			}()

			h.ServeHTTP(w, r.WithContext(c))
		})
	}
}

// Get retrieves a new object, by internally retrieving a pooler holder from the context
// by its key, and getting an object from that holders pool.
func Get(ctx context.Context, key interface{}) interface{} {
	poolMiddleware, ok := ctx.Value(key).(*poolHolder)
	if !ok {
		panic(ErrMissingContext)
	}

	obj := poolMiddleware.pool.Get()
	poolMiddleware.takenObjects[obj] = struct{}{}

	return obj
}
