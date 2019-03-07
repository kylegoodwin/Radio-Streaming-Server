package indexes

import (
	"errors"
	"sort"
	"strings"
	"sync"
	"unicode/utf8"
)

//TODO: implement a trie data structure that stores
//keys of type string and values of type int64

//Trie is the root of a tree of TrieNodes
type Trie struct {
	root TrieNode
	mx   *sync.RWMutex
}

//TrieNode is a struct with a value and a slice of TrieNodes, its children
type TrieNode struct {
	key      string
	values   *[]int64
	children *[]TrieNode
}

//NewTrie returns the root of a Trie
func NewTrie() *Trie {
	return &Trie{
		root: TrieNode{
			children: &[]TrieNode{},
		},
		mx: &sync.RWMutex{},
	}
}

//Add adds a new key value pair to the trie, creating new nodes as necessary
func (t Trie) Add(key string, value int64) error {

	var currentNode *TrieNode
	//Read Lock it
	t.mx.RLock()
	currentNode, err := t.FindNode(key, 0)
	t.mx.RUnlock()
	if err != nil {
		return err
	}
	//Make sure that this value does not already exist at this location
	alreadyExists := false
	if len(*currentNode.values) != 0 {
		for _, i := range *currentNode.values {
			if i == value {
				alreadyExists = true
			}
		}
	}

	//If it does not already exist here, put it there
	if !alreadyExists {
		//Write lock it
		t.mx.Lock()
		*currentNode.values = append(*currentNode.values, value)
		t.mx.Unlock()
	} else {
		return errors.New("That value already exists for this key")
	}

	return nil
}

//Remove removes the value at the key location and trims empty leaf nodes beneath it.
func (t Trie) Remove(key string, value int64) error {
	var currentNode *TrieNode
	//Readlocking it
	t.mx.RLock()
	currentNode, err := t.FindNode(key, 1)
	t.mx.RUnlock()
	if err != nil {
		return err
	}

	//Remove the value from the node
	removeSuccess := false
	for i, v := range *currentNode.values {
		if v == value {
			//Lock mutex, data is about to be edited
			t.mx.Lock()

			//replaces the current values slice with a combination of every value before and after the matched index
			if i != len(*currentNode.values)-1 {
				*currentNode.values = append((*currentNode.values)[:i], (*currentNode.values)[i+1:]...)
			}
			//If it did not meet the previous condition, it must be the last one in the slice so just chop it off
			*currentNode.values = (*currentNode.values)[:len(*currentNode.values)-1]

			//Confirm that the removal happened
			removeSuccess = true

			//Unlock the mutex
			t.mx.Unlock()

			//leave the loop in case it causes some kinda monkey business for messing with the pointer
			break
		}

	}

	if !removeSuccess {
		return errors.New("Unable to find that value at the given key")
	}

	//if its a leaf node && has no more values in it, begin the tedious process of removing empty leaves
	if len(*currentNode.children) == 0 && len(*currentNode.values) == 0 {
		//Write lock, about to edit the structure
		t.mx.Lock()
		currentNode, err := t.RemoveNode(key)
		t.mx.Unlock()
		possibleHeight := len(key)
		for {
			if possibleHeight != 1 {
				possibleHeight = possibleHeight - 1
			}
			//Grab the next node above
			t.mx.RLock()
			currentNode, err = t.FindNode(key[:possibleHeight], 1)
			t.mx.RUnlock()
			if err != nil {
				return err
			}

			//If the next node above this one has any other children or any values, leave and don't remove it
			if len(*currentNode.values) > 0 || len(*currentNode.children) > 0 || possibleHeight == 0 {
				break
			}

			//Remove if it is empty and a leaf
			t.mx.Lock()
			t.RemoveNode(key[:possibleHeight])
			t.mx.Unlock()

			//If the parent of this node we just removed is the root, stop
			if possibleHeight == 1 {
				break
			}

		}

	}

	return nil
}

//Find returns the first n values that match the provided prefix
func (t Trie) Find(searchResultCount int, prefix string) ([]int64, error) {
	t.mx.RLock()
	prefixNode, err := t.FindNode(prefix, 1)
	t.mx.RUnlock()
	if err != nil {
		return nil, err
	}

	//Declare the final slice object and results counter
	var searchResults []int64
	current := 0

	//Run the recursive search
	t.mx.RLock()
	searchResults = append(searchResults, search(prefixNode, searchResultCount, &current)...)
	t.mx.RUnlock()
	//Return the search results
	return searchResults, nil

}

//search is the recursive function used by Find()
func search(node *TrieNode, goal int, current *int) []int64 {
	var returnSlice []int64
	for _, v := range *node.values {
		if goal != *current {
			returnSlice = append(returnSlice, v)
			*current = *current + 1
		} else {
			break
		}

	}

	//Goal met, begin recursive return
	if *current == goal {
		return returnSlice
	}

	//Other return condition: reached the end of that branch
	if len(*node.children) == 0 {
		return returnSlice
	}

	//Sort the children of the current node for alphabetical search
	nextNodes := SortChildren(node)

	//Begin a recursive call that returns an append of the next alphabetically sorted node
	for _, n := range *nextNodes {
		returnSlice = append(returnSlice, search(&n, goal, current)...)
	}

	//Mandatory return to get the compiler to shut up
	return returnSlice

}

//SortChildren returns a sorted slice of a given node's children
func SortChildren(givenNode *TrieNode) *[]TrieNode {
	var sortedNodes []TrieNode

	//If there is only one child then it is not necessary to sort
	if len(*givenNode.children) == 1 {
		sortedNodes = append(sortedNodes, (*givenNode.children)[0])
		return &sortedNodes
	}

	//Populate a map to better track the utf8 values of each node
	mappy := make(map[int]TrieNode)
	for _, n := range *givenNode.children {
		r, _ := utf8.DecodeLastRuneInString(n.key)
		mappy[int(r)] = n
	}

	//Make a slice of the keys
	mapKeys := make([]int, 0, len(mappy))
	for k := range mappy {
		mapKeys = append(mapKeys, k)
	}

	//Sort the key slice
	sort.Ints(mapKeys)

	//Populate the sortedNodes slice in the correct order
	for _, k := range mapKeys {
		sortedNodes = append(sortedNodes, mappy[k])
	}

	//Return the sorted slice of sortedNodes
	return &sortedNodes
}

//RemoveNode removes a given node and returns the parent node
func (t Trie) RemoveNode(key string) (*TrieNode, error) {
	if len(key) != 1 {
		//Find the node above the given node
		parentNode, err := t.FindNode(key[:len(key)-1], 1)
		if err != nil {
			return nil, err
		}
		//extract the exact character which is the key that needs to be removed
		keyToRemove := key[len(key)-1:]

		//Look though it's children and remove the correct one
		for i, c := range *parentNode.children {
			if c.key == keyToRemove {
				//Make sure that the index of the matching child is not the last one in the slice
				if i != len(*parentNode.children)-1 {
					*parentNode.children = append((*parentNode.children)[:i], (*parentNode.children)[i+1:]...)
				}
				//If it did not meet the previous condition, it must be the last one in the slice so just chop it off
				*parentNode.children = (*parentNode.children)[:len((*parentNode.children))-1]
			}
		}
		return parentNode, nil
	}

	//If the key's length is only 1 character long, it must be a child of the root
	for i, c := range *t.root.children {
		//if the child's key(single letter) matches the last letter of the given key (full string)
		//remove it
		if c.key == key {
			//Make sure that the index of the matching child is not the last one in the slice
			if i != len(*t.root.children)-1 {
				*t.root.children = append((*t.root.children)[:i], (*t.root.children)[i+1:]...)
			}
			//If it did not meet the previous condition, it must be the last one in the slice so just chop it off
			*t.root.children = (*t.root.children)[:len((*t.root.children))-1]
		}
	}
	return &t.root, nil

}

//FindNode searches the tree and returns the node at the end of the key. Mode indicates whether it will be used in add (0) or remove(1)
func (t Trie) FindNode(key string, mode int) (*TrieNode, error) {
	var currentNode *TrieNode
	currentNode = &t.root
	keySlice := strings.Split(key, "")
	//Look through the letters of the key string
	for _, i := range keySlice {
		changedNode := false
		//For each child of the current node, check to see if we can go to the next node that matches the key letter
		for _, j := range *currentNode.children {
			if j.key == i {
				currentNode = &j
				changedNode = true
				//Leave the loop now that it was found
				break
			}
		}
		//If the current node was not updated then it didnt exist in that node's children, so create it
		if !changedNode {
			if mode == 0 {
				newNode := &TrieNode{
					key:      i,
					values:   &[]int64{},
					children: &[]TrieNode{},
				}
				*currentNode.children = append(*currentNode.children, *newNode)
				currentNode = newNode
			} else if mode == 1 {
				return nil, errors.New("Did not find node at given key: " + key)
			}
		}
	}
	return currentNode, nil
}
