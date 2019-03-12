package indexes

import (
	"fmt"
	"sort"
	"sync"
)

type Trie struct {
	root *Node
	mx   *sync.RWMutex
}

type Node struct {
	val      rune
	parent   *Node
	children map[rune]*Node
	data     []int64
	visited  bool
}

// Creates a new Trie with an initialized root Node.
func NewTrie() *Trie {
	return &Trie{
		root: &Node{children: make(map[rune]*Node)},
		mx:   &sync.RWMutex{},
	}
}

// Returns the root node for the Trie.
func (t *Trie) Root() *Node {
	return t.root
}

// Adds the key to the Trie, attaching an id to the data store
func (t *Trie) Add(key string, id int64) *Node {

	t.mx.Lock()
	defer t.mx.Unlock()

	//Get the main node of the tree
	node := t.root

	//Convert the string key into a Unicode character
	runes := []rune(key)

	for i := range runes {
		r := runes[i]

		if node.HasChildren() {

			_, ok := node.children[r]

			//If there node has a child with rune, go to that one
			if ok {
				node = node.children[r]

				//And if we are at the end of the string, add the id to this node
				//make sure that it is unique in the set
				if i == len(runes)-1 && !contains(node.data, id) {
					node.data = append(node.data, id)
				}
				//There is children but not the right ones
			} else {
				if i == len(runes)-1 {
					node = node.AddChild(r, id)
				} else {
					node = node.AddChild(r, -1)
				}
			}
			//If the node has no children
		} else {
			//At end of string
			if i == len(runes)-1 {
				node = node.AddChild(r, id)
			} else {
				node = node.AddChild(r, -1)
			}
		}
	}

	//t.mx.Unlock()
	return node
}

//Removes a given key from the Trie
func (t *Trie) Remove(key string, value int64) error {
	//Lock for concurency
	t.mx.Lock()
	defer t.mx.Unlock()

	//Convert the string key into a Unicode character
	runes := []rune(key)
	node := FindNode(runes, t.root)

	//If it has values, just remove
	if node != nil && contains(node.data, value) {
		node.data = remove(node.data, value)

		if len(node.data) == 0 {
			node.data = nil

			if !node.HasChildren() {
				t.RemoveHelp(node)
			}
		}
		return nil
	}

	return fmt.Errorf("Key Value Pair: %s %s Not found", key, string(value))

}

//RemoveHelp removes a dead node from the trie
func (t *Trie) RemoveHelp(currentNode *Node) {
	if currentNode == t.root {
		return
	}
	old := currentNode
	currentNode = old.parent

	//Removes reference in map to the child that is the dead node
	delete(currentNode.children, old.val)

	//Remove reference to the current node,
	//its now not linked to anything else
	old.parent = nil

	if currentNode.data != nil {
		return
	}
	t.RemoveHelp(currentNode)

}

//Find Retrives a list of UserID's that match a given prefix
func (t *Trie) Find(name string, num int64) []int64 {
	t.mx.RLock()
	defer t.mx.RUnlock()

	node := FindNode([]rune(name), t.Root())
	if node == nil {
		return []int64{}
	}
	var nums []int64
	val := FindHelper(node, num, nums)
	return val

}

//FindHelper recurses through the tree and collects userids that match a prefix,
//it only collects up to a certain number of ID's
func FindHelper(node *Node, num int64, results []int64) []int64 {

	//If the length of the data array is greater than zero
	//loop through it and add the values to results
	//Else just keep going down...
	if len(results) == int(num) {
		return results
	}

	//Sort for "alphabetical" order
	var childRunes []*Node
	for k := range node.children {
		childRunes = append(childRunes, node.children[k])
	}
	sort.Slice(childRunes, func(i, j int) bool {
		return childRunes[i].val < childRunes[j].val
	})

	//Add each datapoint
	for _, data := range node.data {
		results = append(results, data)
	}

	for _, k := range childRunes {
		results = FindHelper(k, num, results)

	}

	return results
}

//FindNode navigates the tree for a given prefix and returns the last
// node for a prefix
func FindNode(name []rune, n *Node) *Node {

	ok := true
	var node *Node

	if len(name) > 0 {
		node, ok = n.children[name[0]]
	} else {
		node = n
	}

	var nrunes []rune
	if len(name) > 1 {
		nrunes = name[1:]
	} else {
		nrunes = name[0:0]
	}

	if !ok {
		//This means that the node has no child that is in the prefix
		//therefore it should be nil, because there can be no values benith it
		return nil
	}

	//the string is empty, return the last node
	if len(nrunes) == 0 {
		return node
	}

	return FindNode(nrunes, node)
}

//HasChildren returns true if the node is not a leaf node
func (n *Node) HasChildren() bool {
	return len(n.children) != 0
}

//AddChild Adds a child node to a given node
func (n *Node) AddChild(val rune, data int64) *Node {
	node := &Node{
		val:      val,
		parent:   n,
		children: make(map[rune]*Node),
	}

	if data >= 0 {
		node.data = []int64{data}
	}

	//Add to the parents list of children the new node
	n.children[val] = node
	return node
}

//Contains scans a slice of Userids to see if a given id is in the slice
func contains(s []int64, e int64) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

//Remove removes a given UserID from a given slice
func remove(s []int64, e int64) []int64 {

	//Make old value -1
	for i, val := range s {
		if val == e {
			s[i] = -1
		}
	}

	//Sort array so that old value is in the begining
	sort.Slice(s, func(i, j int) bool {
		return s[i] < s[j]
	})

	return s[1:]

}
