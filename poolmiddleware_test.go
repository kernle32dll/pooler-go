package pooler_test

import (
	"github.com/kernle32dll/pooler-go"

	"context"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

type sampleObj struct{}

const someKey = iota

func TestGet(t *testing.T) {

	t.Run("get-retrieval", func(t *testing.T) {
		t.Parallel()

		sampleObj := &sampleObj{}

		middleware := pooler.NewMiddleware(someKey, func() interface{} {
			return sampleObj
		})

		server := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			got := pooler.Get(r.Context(), someKey)

			if !reflect.DeepEqual(got, sampleObj) {
				t.Errorf("Get() = %v, want %v", got, sampleObj)
			}
		}))

		server.ServeHTTP(httptest.NewRecorder(), &http.Request{})
	})

	t.Run("request-panic", func(t *testing.T) {
		t.Parallel()

		panicVal := "some-error"

		middleware := pooler.NewMiddleware(someKey, nil)

		server := middleware(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
			panic(panicVal)
		}))

		defer func() {
			if got := recover(); got != panicVal {
				t.Errorf("Request panic () = %v, want %v", got, panicVal)
			}
		}()

		server.ServeHTTP(httptest.NewRecorder(), &http.Request{})
	})

	t.Run("get-panic", func(t *testing.T) {
		t.Parallel()

		defer func() {
			if got := recover(); got != pooler.ErrMissingContext {
				t.Errorf("Get panic () = %v, want %v", got, pooler.ErrMissingContext)
			}
		}()

		pooler.Get(context.Background(), "some-key")
	})
}
