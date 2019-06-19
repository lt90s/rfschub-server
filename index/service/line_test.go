package service

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestLiner_getLine(t *testing.T) {
	content := "hello\nworld\nfoo\nbar"

	liner := newLiner([]byte(content))
	_, err := liner.getLine(0)
	require.Error(t, err)

	line, err := liner.getLine(1)
	require.NoError(t, err)
	require.Equal(t, "hello", line)
	require.Equal(t, 2, liner.lineNumber)

	line, err = liner.getLine(2)
	require.NoError(t, err)
	require.Equal(t, "world", line)
	require.Equal(t, 3, liner.lineNumber)

	line, err = liner.getLine(3)
	require.NoError(t, err)
	require.Equal(t, "foo", line)
	require.Equal(t, 4, liner.lineNumber)

	line, err = liner.getLine(4)
	require.NoError(t, err)
	require.Equal(t, "bar", line)
	require.Equal(t, 4, liner.lineNumber)
	//require.Equal(t, true, liner.done)

	line, err = liner.getLine(5)
	require.Equal(t, errLineNotExist, err)
	require.Equal(t, 4, liner.lineNumber)
}

var code = `package circuitbreaker

import (
	"context"

	"github.com/sony/gobreaker"

	"github.com/go-kit/kit/endpoint"
)

// Gobreaker returns an endpoint.Middleware that implements the circuit
// breaker pattern using the sony/gobreaker package. Only errors returned by
// the wrapped endpoint count against the circuit breaker's error count.
//
// See http://godoc.org/github.com/sony/gobreaker for more information.
func Gobreaker(cb *gobreaker.CircuitBreaker) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (interface{}, error) {
			return cb.Execute(func() (interface{}, error) { return next(ctx, request) })
		}
	}
}
`

func TestLiner(t *testing.T) {
	liner := newLiner([]byte(code))

	t.Log(liner.getLine(15))
	t.Log(liner.getLine(16))
}
