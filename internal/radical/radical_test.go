package radical_test

import (
	"testing"

	"github.com/blakewilliams/fernet/internal/radical"
	"github.com/stretchr/testify/require"
)

func TestNode(t *testing.T) {
	root := radical.New[int]()

	// Simulate /foo/bar/baz
	root.Add([]string{"foo", "bar", "baz"}, 1)

	ok, value := root.Value([]string{"foo"})
	require.False(t, ok)
	require.Zero(t, value)

	ok, value = root.Value([]string{"foo", "bar"})
	require.False(t, ok)
	require.Zero(t, value)

	ok, value = root.Value([]string{"foo", "bar", "baz"})
	require.True(t, ok)
	require.Equal(t, 1, value)
}

func TestNode_Dynamic(t *testing.T) {
	root := radical.New[int]()

	// Simulate /foo/bar/baz
	root.Add([]string{"foo", ":bar", "baz"}, 1)

	ok, value := root.Value([]string{"foo"})
	require.False(t, ok)
	require.Zero(t, value)

	ok, value = root.Value([]string{"foo", "bar"})
	require.False(t, ok)
	require.Zero(t, value)

	ok, value = root.Value([]string{"foo", "bar", "baz"})
	require.True(t, ok)
	require.Equal(t, 1, value)
}

func TestNode_Backpedal(t *testing.T) {
	root := radical.New[int]()

	// Simulate /foo/bar/baz
	root.Add([]string{"foo", "bar", "baz"}, 1)

	ok, value := root.Value([]string{"foo"})
	require.False(t, ok)
	require.Zero(t, value)

	ok, value = root.Value([]string{"foo", "bar"})
	require.False(t, ok)
	require.Zero(t, value)

	// Simulate /foo/bar/baz
	root.Add([]string{"foo", "bar"}, 2)

	ok, value = root.Value([]string{"foo", "bar"})
	require.True(t, ok)
	require.Equal(t, 2, value)

	ok, value = root.Value([]string{"foo", "bar", "baz"})
	require.True(t, ok)
	require.Equal(t, 1, value)

	// Simulate /foo/bar/baz
	root.Add([]string{}, 3)

	ok, value = root.Value([]string{})
	require.True(t, ok)
	require.Equal(t, 3, value)
}
