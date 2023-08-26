// Package radical implements a basic radix trie like structure for use in the
// fernet router.
package radical

import (
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

	for _, segment := range segments {
		if strings.HasPrefix(segment, ":") {
			child, ok := n.children[":named"]

			if ok {
				currentSegment = child
				continue
			}

			n.children[":named"] = &Node[T]{
				segment:  ":named",
				children: make(map[string]*Node[T], 0),
			}

			currentSegment = n.children[":named"]
			continue
		}

		child, ok := n.children[segment]
		if ok {
			currentSegment = child
			continue
		}

		n.children[segment] = &Node[T]{
			segment:  segment,
			children: make(map[string]*Node[T], 0),
		}

		currentSegment = n.children[segment]
	}

	currentSegment.value = value
	currentSegment.isSet = true
}

// Value searches the tree for a node matching the provided segments. If a match
// is found it returns true and the associated value T. If a match is not found
// it returns false and the zero value of T.
func (n *Node[T]) Value(segments []string) (bool, T) {
	currentNode := n
	for _, segment := range segments {
		child, ok := n.children[segment]
		if ok {
			currentNode = child
			continue
		}

		child, ok = n.children[":named"]
		if !ok {
			return false, reflect.Zero(reflect.TypeOf(n.value)).Interface().(T)
		}

		currentNode = child
	}

	if currentNode.isSet {
		return true, currentNode.value
	}

	return false, currentNode.value
}
