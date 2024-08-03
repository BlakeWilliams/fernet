// Package radical implements a basic radix trie like structure for use in the
// fernet router.
package radical

import (
	"fmt"
	"reflect"
	"strings"
)

type (
	// Node represents a node in the tree
	Node[T any] struct {
		// The path segment for this node.
		segment string
		// The handler for this node.
		value T
		// Is there a value set?
		isSet bool
		// The children of this node.
		children map[string]*Node[T]
	}
)

// New returns a new root Radix tree node
func New[T any]() *Node[T] {
	return &Node[T]{
		segment:  "",
		children: make(map[string]*Node[T], 0),
	}
}

// Add adds a new node to the tree.
func (n *Node[T]) Add(segments []string, value T) {
	currentSegment := n

	for i, segment := range segments {
		if strings.HasPrefix(segment, ":") {
			child, ok := currentSegment.children[":named"]

			if ok {
				currentSegment = child
				continue
			}

			currentSegment.children[":named"] = &Node[T]{
				segment:  ":named",
				children: make(map[string]*Node[T], 0),
			}

			currentSegment = currentSegment.children[":named"]
			continue
		}

		if strings.HasPrefix(segment, "*") {
			if i != len(segments)-1 {
				panic("wildcard segments must be the last segment in a path")
			}

			_, ok := currentSegment.children["*"]
			if ok {
				panic("wildcard segments can only be used once in a path")
			}

			currentSegment.children["*"] = &Node[T]{
				segment:  "*",
				children: make(map[string]*Node[T], 0),
			}

			currentSegment = currentSegment.children["*"]

			break
		}

		child, ok := currentSegment.children[segment]
		if ok {
			currentSegment = child
			continue
		}

		currentSegment.children[segment] = &Node[T]{
			segment:  segment,
			children: make(map[string]*Node[T], 0),
		}

		currentSegment = currentSegment.children[segment]
	}

	if currentSegment.isSet {
		panic(fmt.Sprintf("duplicate route detected: %v", segments))
	}

	currentSegment.value = value
	currentSegment.isSet = true
}

// Value searches the tree for a node matching the provided segments. If a match
// is found it returns true and the associated value T. If a match is not found
// it returns false and the zero value of T.
func (n *Node[T]) Value(segments []string) (bool, T) {
	currentNode := n
	var lastWildcard *Node[T]

	for _, segment := range segments {
		// hold onto the last wildcard node we've seen in case we need it later
		// if routes don't match
		if wildcard, ok := currentNode.children["*"]; ok {
			lastWildcard = wildcard
		}

		child, ok := currentNode.children[segment]
		if ok {
			currentNode = child
			continue
		}

		child, ok = currentNode.children[":named"]
		if !ok {
			if lastWildcard != nil {
				return true, lastWildcard.value
			}

			return false, reflect.Zero(reflect.TypeOf(n.value)).Interface().(T)
		}

		currentNode = child
	}

	if currentNode.isSet {
		return true, currentNode.value
	}

	if lastWildcard != nil {
		return true, lastWildcard.value
	}

	return false, currentNode.value
}
