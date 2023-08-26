package fernet

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRequest_WithContext(t *testing.T) {
	httpReq, _ := http.NewRequest(http.MethodGet, "/", nil)
	req := &Request[int]{
		params: map[string]string{"hello": "world"},
		req:    httpReq,
		ctx:    context.TODO(),
		Data:   1,
	}

	// Ensure all relevant data is cloned
	require.Equal(t, req, req.WithContext(req.Context()))
}
