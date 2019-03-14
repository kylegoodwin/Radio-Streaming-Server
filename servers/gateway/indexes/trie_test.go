package indexes

//TODO: implement automated tests for your trie data structure

import (
	"fmt"
	"reflect"
	"testing"
)

type TestTuple struct {
	key string
	id  int64
}

func Test_Trie(t *testing.T) {

	cases := []struct {
		name                 string
		addValues            []TestTuple
		findLimit            int64
		findKey              string
		removeValues         []TestTuple
		expectError          bool
		expectedBeforeDelete []int64
		expectedAfterDelete  []int64
	}{
		{
			"Adding values only",
			[]TestTuple{TestTuple{"a", 1}, TestTuple{"ab", 2}},
			10,
			"a",
			nil,
			false,
			[]int64{1, 2},
			[]int64{1, 2},
		},
		{
			"Adding values with Unicode",
			[]TestTuple{TestTuple{"a", 1}, TestTuple{"ab", 2}, TestTuple{"a😘", 3}},
			10,
			"a",
			nil,
			false,
			[]int64{1, 2, 3},
			[]int64{1, 2, 3},
		},
		{
			"Adding values only with root node",
			[]TestTuple{TestTuple{"a", 1}, TestTuple{"ab", 2}},
			10,
			"",
			nil,
			false,
			[]int64{1, 2},
			[]int64{1, 2},
		},
		{
			"Adding values and only getting a limited number backc",
			[]TestTuple{TestTuple{"a", 1}, TestTuple{"ab", 2}},
			1,
			"a",
			nil,
			false,
			[]int64{1},
			[]int64{1},
		},
		{
			"Adding multi child values",
			[]TestTuple{TestTuple{"a", 1}, TestTuple{"ab", 2}, TestTuple{"ac", 3}},
			10,
			"a",
			nil,
			false,
			[]int64{1, 2, 3},
			[]int64{1, 2, 3},
		},
		{
			"Removing all values",
			[]TestTuple{TestTuple{"a", 1}, TestTuple{"ab", 2}, TestTuple{"ab", 3}},
			10,
			"a",
			[]TestTuple{TestTuple{"a", 1}, TestTuple{"ab", 2}, TestTuple{"ab", 3}},
			false,
			[]int64{1, 2, 3},
			[]int64{},
		},
		{
			"Removing first value of branch only",
			[]TestTuple{TestTuple{"a", 1}, TestTuple{"ab", 2}, TestTuple{"ab", 3}, TestTuple{"foo", 5}},
			10,
			"a",
			[]TestTuple{TestTuple{"a", 1}},
			false,
			[]int64{1, 2, 3},
			[]int64{2, 3},
		},
		{
			"Removing recursivly",
			[]TestTuple{TestTuple{"a", 1}, TestTuple{"ab", 2}, TestTuple{"abc", 3}, TestTuple{"foo", 5}},
			10,
			"f",
			[]TestTuple{TestTuple{"foo", 5}},
			false,
			[]int64{5},
			[]int64{},
		},
		{
			"Removing recursivly with node above that has data",
			[]TestTuple{TestTuple{"a", 1}, TestTuple{"ab", 2}, TestTuple{"abc", 3}, TestTuple{"foo", 5}},
			10,
			"a",
			[]TestTuple{TestTuple{"abc", 3}},
			false,
			[]int64{1, 2, 3},
			[]int64{1, 2},
		},
		{
			"Removing key that doesnt exist",
			[]TestTuple{TestTuple{"a", 1}, TestTuple{"ab", 2}, TestTuple{"abc", 3}, TestTuple{"foo", 5}},
			10,
			"f",
			[]TestTuple{TestTuple{"darn", 55}},
			true,
			[]int64{5},
			[]int64{},
		},

		//Need more unicode cases
		//Also need cases where nothing should be returned
	}

	//Add values to trie
	for _, c := range cases {
		fmt.Println(c.name)
		tree := NewTrie()

		for _, tuple := range c.addValues {
			tree.Add(tuple.key, tuple.id)
		}

		//After add before delete tests
		if !reflect.DeepEqual(tree.Find(c.findKey, c.findLimit), c.expectedBeforeDelete) {
			t.Errorf("Error with find operation on trie after adding on case: %s", c.name)
		}
		for _, tuple := range c.removeValues {
			fmt.Println("*** Removing " + tuple.key + string(tuple.id))
			err := tree.Remove(tuple.key, tuple.id)
			if !c.expectError && err != nil {
				t.Errorf("Unexpected error removing values from trie for case: %s Error message: %s", c.name, err.Error())
			}
		}

		//After add before delete tests
		if !c.expectError && (len(tree.Find(c.findKey, c.findLimit)) != len(c.expectedAfterDelete) && !reflect.DeepEqual(tree.Find(c.findKey, c.findLimit), c.expectedAfterDelete)) {
			t.Errorf("Error with find operation on trie after round of deletions on case %s", c.name)
		}

	}

}
