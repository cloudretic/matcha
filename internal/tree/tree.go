// Package tree defines a method of recursively matching route paths by part.
//
// See [https://github.com/jnichols-git/matcha/v2/blob/main/docs/routers.md#matching].
package tree

import (
	"net/http"
	"sync/atomic"

	"github.com/jnichols-git/matcha/v2/internal/path"
	"github.com/jnichols-git/matcha/v2/require"
	"github.com/jnichols-git/matcha/v2/route"
)

const NO_LEAF_ID = int64(0)

type node struct {
	p             route.Part
	children      []*node
	leaf_id       int64
	leaf_required []require.Required
}

func (n *node) isLeaf() bool {
	return n.leaf_id != NO_LEAF_ID
}

func createNode(p route.Part) *node {
	return &node{
		p:        p,
		children: make([]*node, 0),
	}
}

func (n *node) resolveLeafForRequest(req *http.Request) int64 {
	if n.leaf_id == NO_LEAF_ID {
		return NO_LEAF_ID
	}
	if !require.Execute(req, n.leaf_required) {
		return NO_LEAF_ID
	}
	return n.leaf_id
}

// Propagate a set of parts through the tree, with this node as the root.
// If there are no parts left to propagate, the node will instead be set to leaf leaf_id.
func (n *node) propagate(r *route.Route, ps []route.Part, leaf_id int64) {
	if len(ps) == 0 {
		n.leaf_id = leaf_id
		n.leaf_required = r.Required()
		return
	}
	next := ps[0]
	if !n.isLeaf() && len(ps)-1 != 0 {
		for _, child := range n.children {
			if child.p.Eq(next) && !child.isLeaf() {
				child.propagate(r, ps[1:], leaf_id)
				return
			}
		}
	}
	child := createNode(next)
	child.propagate(r, ps[1:], leaf_id)
	n.children = append(n.children, child)
}

// match traverses a subtree of nodes to find the first matching route.
func (n *node) match(req *http.Request, expr string, last int) int64 {
	// If we've reached the end of the expression, return the leaf_id of the current node.
	// This encapsulates several edge cases where it's difficult to know if the routine should return early or not,
	// like with partial leaves.
	if last == -1 {
		return n.resolveLeafForRequest(req)
	}
	// Get the next token from the path and match it against the route.Part of the current node.
	token, next := path.Next(expr, last)
	ok := n.p.Match(token)
	if !ok {
		// If the part doesn't match, return NO_LEAF_ID.
		return NO_LEAF_ID
	} else if n.isLeaf() {
		// If the part matches...
		if n.p.Multi() {
			// ...and the leaf is partial, return the result of recursively matching until termination.
			return n.match(req, expr, next)
		} else if next == -1 {
			// ...and the route has been exhausted, return the id of the leaf as a successful match.
			return n.resolveLeafForRequest(req)
		} else {
			// ...and the route has not been exhausted, return NO_LEAF_ID.
			return NO_LEAF_ID
		}
	}
	// If next is -1, we've exhausted the path without matching.
	if next == -1 {
		return NO_LEAF_ID
	}
	// Iterate through the children of this node.
	for _, child := range n.children {
		if match_leaf_id := child.match(req, expr, next); match_leaf_id != NO_LEAF_ID {
			// If a child matches the entire remaining route, return its leaf_id.
			return match_leaf_id
		}
	}
	// If we reach this point, the entire subtree from this node has been traversed with no match.
	return NO_LEAF_ID
}

type RouteTree struct {
	methodRoot map[string]*node
	nextId     atomic.Int64
}

// Create a new RouteTree.
func New() *RouteTree {
	return &RouteTree{
		methodRoot: make(map[string]*node),
		nextId:     atomic.Int64{},
	}
}

// Add a route to the tree.
// Returns the leaf ID of the added route.
func (rtree *RouteTree) Add(r *route.Route) int64 {
	root, ok := rtree.methodRoot[r.Method()]
	if !ok || root == nil {
		root = createNode(route.Part{})
		rtree.methodRoot[r.Method()] = root
	}
	id := rtree.nextId.Add(1)
	root.propagate(r, r.Parts(), id)
	return id
}

// Match a request to the tree.
// Returns the leaf ID of the matched route, or NO_LEAF_ID if no match is found.
func (rtree *RouteTree) Match(req *http.Request) int64 {
	root, ok := rtree.methodRoot[req.Method]
	if !ok || root == nil {
		return 0
	}
	expr := req.URL.Path
	for _, r := range root.children {
		match_leaf_id := r.match(req, expr, 0)
		if match_leaf_id != NO_LEAF_ID {
			return match_leaf_id
		}
	}
	return 0
}
