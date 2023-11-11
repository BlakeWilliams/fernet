package fernet

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

type MyRequestContext struct {
	*RootRequestContext
}

type PostDataWrongSignature struct{}

func (pd *PostDataWrongSignature) FromRequest(i int) bool {
	return true
}

func Test_FailureCases(t *testing.T) {
	testCases := map[string]struct {
		fn           any
		panicMessage string
	}{
		"non-function handler": {
			fn:           1,
			panicMessage: "handlers must be a function",
		},
		"pointer receiver with non-pointer": {
			fn:           func(p PostData) {},
			panicMessage: "fernet.PostData of func(fernet.PostData), does not implement FromRequest[fernet.RequestContext]. FromRequest has pointer receiver, but fernet.PostData is not a pointer",
		},
		"wrong FromRequest signature": {
			fn:           func(p *PostDataWrongSignature) {},
			panicMessage: "FromRequest method on *fernet.PostDataWrongSignature of func(*fernet.PostDataWrongSignature), must have the signature `func(context.Context, fernet.RequestContext) bool. Got `*fernet.PostDataWrongSignature`",
		},
		"invalid type": {
			fn:           func(i int) {},
			panicMessage: "paramter 1 (int) in function func(int) is not a valid type, must be context.Context, fernet.RequestContext, or implement FromRequest[fernet.RequestContext]",
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			require.PanicsWithValue(t, tc.panicMessage, func() {
				createHandler[RequestContext](tc.fn)
			})
		})
	}
}

func Test_FailureWithWrongRequestContext(t *testing.T) {
	require.PanicsWithValue(t, "received RequestContext type *fernet.RootRequestContext, but expected *fernet.MyRequestContext", func() {
		createHandler[*MyRequestContext](func(rc *RootRequestContext) {})
	})
}

type ShortCircuitFromRequest struct{}

func (s ShortCircuitFromRequest) FromRequest(ctx context.Context, r RequestContext) bool {
	return false
}

func Test_FromRequestFalseShortCircuits(t *testing.T) {
	called := false
	h := createHandler[RequestContext](func(s ShortCircuitFromRequest) {
		called = true
	})

	h(context.Background(), &RootRequestContext{})

	require.False(t, called, "expected ShortCircuitFromRequest to short circuit handler")
}
