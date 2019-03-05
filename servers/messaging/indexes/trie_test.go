package indexes

import (
	"log"
	"testing"
)

//TODO: implement automated tests for your trie data structure

func TestAdd(t *testing.T) {
	//testTrie := NewTrie()

	// testTrie.Add("a", int64(1))
	// testTrie.Add("ab", int64(2))
	// testTrie.Add("ab", int64(2))

	cases := []struct {
		name        string
		key         string
		value       int64
		expectError bool
	}{
		{
			"Single length key",
			"A",
			int64(1),
			false,
		},
		{
			"Multi length key",
			"BC",
			int64(1),
			false,
		},
		{
			"Multi length key - Repeat value",
			"BC",
			int64(1),
			true,
		},
		{
			"Multi length key - uses same branch",
			"ABC",
			int64(1),
			false,
		},
	}

	tt := NewTrie()

	for _, c := range cases {
		log.Printf("Running case: %s", c.name)
		err := tt.Add(c.key, c.value)
		if err != nil && !c.expectError {
			t.Errorf("Case %v returned an unexpected error: %v", c.name, err.Error())
		}
		if c.expectError && err == nil {
			t.Errorf("Expected error but did not receive one. Case: %v", c.name)
		}
	}

}

func TestRemove(t *testing.T) {
	testData := []struct {
		key   string
		value int64
	}{
		{
			"H",
			int64(1),
		},
		{
			"HI",
			int64(2),
		},
		{
			"HEY",
			int64(3),
		},
		{
			"HELL",
			int64(4),
		},
		{
			"HE",
			int64(5),
		},
		{
			"HE",
			int64(2),
		},
		{
			"HE",
			int64(3),
		},
		{
			"HE",
			int64(4),
		},
		{
			"HEAR",
			int64(2),
		},
	}

	cases := []struct {
		name        string
		key         string
		value       int64
		expectError bool
	}{
		{
			"Remove and trim",
			"HELL",
			int64(4),
			false,
		},
		{
			"Remove 1st child of root when there are children beneath it",
			"H",
			int64(1),
			false,
		},
		{
			"Remove one child from multi value node",
			"HE",
			int64(2),
			false,
		},
		{
			"Remove and avoid trim",
			"HE",
			int64(5),
			false,
		},
		{
			"Remove when there are two children",
			"HEY",
			int64(3),
			false,
		},
		{
			"Fail to find node",
			"HAR",
			int64(1),
			true,
		},
		{
			"Fail to find value at key",
			"H",
			int64(5),
			true,
		},
	}

	tt := NewTrie()

	for _, d := range testData {
		err := tt.Add(d.key, d.value)
		if err != nil {
			t.Errorf("Adding returned an unexpected error: %v", err.Error())
		}

	}
	for _, c := range cases {
		log.Printf("\n\n\nRunning case: %s", c.name)

		err := tt.Remove(c.key, c.value)

		if !c.expectError && err != nil {
			t.Errorf("Case %v returned an unexpected error: %v", c.name, err.Error())
		}
		if c.expectError && err == nil {
			t.Errorf("Expected error but did not receive one. Case: %v", c.name)
		}
	}

	//Test specific if statement concerning removal of leaf node that is child of root
	tt = NewTrie()
	err := tt.Add("HI", int64(1))
	if err != nil {
		t.Error("Broke on Add, special case")
	}
	err = tt.Remove("HI", int64(1))
	if err != nil {
		t.Error("Broke on Remove, special case")
	}

}

func TestFind(t *testing.T) {
	testData := []struct {
		key   string
		value int64
	}{
		{
			"HI",
			int64(3),
		},
		{
			"HI",
			int64(1),
		},
		{
			"HELLO",
			int64(2),
		},
	}

	cases := []struct {
		name        string
		valueCount  int
		prefix      string
		expectError bool
	}{
		{
			"Find at child of root",
			2,
			"H",
			false,
		},
		{
			"Find large number",
			20,
			"H",
			false,
		},
		{
			"Find with longer prefix",
			2,
			"HEL",
			false,
		},
	}

	tt := NewTrie()

	for _, d := range testData {
		err := tt.Add(d.key, d.value)
		if err != nil {
			t.Errorf("Adding returned an unexpected error: %v", err.Error())
		}

	}
	for _, c := range cases {
		log.Printf("\n\n\nRunning case: %s", c.name)

		foundValues, err := tt.Find(c.valueCount, c.prefix)

		log.Println("\nPrinting result slice:")
		for i, x := range foundValues {
			log.Printf("Result set index %v has value %v", i, x)
		}
		if !c.expectError && err != nil {
			t.Errorf("Case %v returned an unexpected error: %v", c.name, err.Error())
		}
		if c.expectError && err == nil {
			t.Errorf("Expected error but did not receive one. Case: %v", c.name)
		}
	}
}
