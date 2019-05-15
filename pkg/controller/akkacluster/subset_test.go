package akkacluster

import (
	"encoding/json"
	"io/ioutil"
	"reflect"
	"testing"
	"time"

	appsv1 "k8s.io/api/apps/v1"
)

// whitebox tests: A<=B, A<=A, B!<=A, B<=B
func test(a, b interface{}, t *testing.T) {
	t.Helper() // log test messages from caller instead of this helper function
	if !SubsetEqual(a, b) {
		t.Error("expecting subset, but found none", a)
	}
	if SubsetEqual(b, a) {
		t.Error("expected no subset, but found one", b)
	}
	if !SubsetEqual(a, a) {
		t.Error("expecting subset of self to be true, but found false", a)
	}
	if !SubsetEqual(b, b) {
		t.Error("expecting subset of self to be true, but found false", b)
	}
}

// blackbox tests: verify matching node count
func testMatchCount(a, b interface{}, expectedMatchCount int, t *testing.T) {
	t.Helper()
	// mirror inner workings of SubsetEqual so we can inspect the treeWalk metadata
	tree := newTreeWalk()
	if a == nil && expectedMatchCount != 0 {
		t.Fatalf("should expect no matches on nil subset, but wanted %d", expectedMatchCount)
	}
	if a != nil && b == nil && expectedMatchCount != 0 {
		t.Fatalf("should expect no matches against nil superset, but wanted %d", expectedMatchCount)
	}

	// A <= B
	if !tree.subsetValueEqual(reflect.ValueOf(a), reflect.ValueOf(b)) {
		t.Error("expecting subset, but found none", a)
	}
	if expectedMatchCount != tree.matches {
		t.Errorf("expecting %d matches, but got %d", expectedMatchCount, tree.matches)
	}

	// A <= A
	tree = newTreeWalk()
	if !tree.subsetValueEqual(reflect.ValueOf(a), reflect.ValueOf(a)) {
		t.Error("expecting subset of self, but found none", a)
	}
	if expectedMatchCount != tree.matches {
		t.Errorf("expecting %d matches against self, but got %d", expectedMatchCount, tree.matches)
	}
}

func testN(a, b interface{}, expectedMatches int, t *testing.T) {
	t.Helper()
	test(a, b, t)
	testMatchCount(a, b, expectedMatches, t)
}

func test0(a, b interface{}, t *testing.T) {
	t.Helper()
	testN(a, b, 0, t)
}

func test1(a, b interface{}, t *testing.T) {
	t.Helper()
	testN(a, b, 1, t)
}

func TestSubsetDeployment(t *testing.T) {
	deploymentSuperset, err := ioutil.ReadFile("testsubset/deployment_superset.json")
	if err != nil {
		t.Fatal(err)
	}
	deploymentSubset, err := ioutil.ReadFile("testsubset/deployment_subset.json")
	if err != nil {
		t.Fatal(err)
	}

	subset := &appsv1.Deployment{}
	superset := &appsv1.Deployment{}
	if err := json.Unmarshal(deploymentSubset, subset); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(deploymentSuperset, superset); err != nil {
		t.Fatal(err)
	}
	testN(subset, superset, 30, t)
}

func TestSubsetBasics(t *testing.T) {
	type p struct {
		Bool bool
		Code []byte
		Map  map[string]int
	}

	type f struct {
		Name       string
		Age        uint
		Percentage float64
		Timestamp  time.Time
		Nested     *p
	}
	superset := &f{
		Name:       "ohai",
		Age:        6,
		Percentage: 0.99,
		Timestamp:  time.Unix(1557787336, 1),
		Nested: &p{
			Bool: true,
			Code: []byte{'a', 'b', 'c'},
			Map:  map[string]int{"a": 1, "b": 2, "c": 3},
		},
	}
	var empty interface{}

	test0(nil, superset, t)
	test0(empty, superset, t)
	test0(&f{Timestamp: superset.Timestamp}, superset, t) // time.Time is opaque

	test1(&f{Name: superset.Name}, superset, t)
	test1(&f{Age: superset.Age}, superset, t)
	test1(&f{Percentage: superset.Percentage}, superset, t)
	test1(&f{Nested: &p{Bool: superset.Nested.Bool}}, superset, t)
	test1(&f{Nested: &p{Code: []byte{superset.Nested.Code[0]}}}, superset, t)
	test1(&f{Nested: &p{Map: map[string]int{"b": 2}}}, superset, t)

	testN(&f{Nested: &p{Code: superset.Nested.Code}}, superset, len(superset.Nested.Code), t)

	// invalid comparisons
	if SubsetEqual(&f{Name: "onoe"}, nil) {
		t.Error("expecting subset test to fail because nil superset")
	}
	if SubsetEqual(&f{Name: "onoe"}, "onoe") {
		t.Error("expecting subset test to fail because mismatched types")
	}
	if SubsetEqual(&f{Nested: &p{Bool: superset.Nested.Bool}}, &f{Nested: nil}) {
		t.Error("expecting subset test to fail because superset invalid")
	}

	// deeper mismatch
	if SubsetEqual(&f{Nested: &p{Code: []byte{'x'}}}, superset) {
		t.Error("expecting subset test to fail because slice mismatch")
	}
	if SubsetEqual(&f{Nested: &p{Map: map[string]int{"b": 1}}}, superset) {
		t.Error("expecting subset test to fail because map mismatch", superset.Nested.Map)
	}

	// recursive subset: embedded containers are themselves subsets
	// using json serde to get a deep copy of superset, so everything matches except one thing
	copy, _ := json.Marshal(superset)
	subset := &f{}
	json.Unmarshal(copy, subset)
	subset.Nested.Code = []byte{'a', 'b'}
	testN(subset, superset, 9, t)

	json.Unmarshal(copy, subset)
	subset.Nested.Map = map[string]int{"b": 2, "c": 3}
	testN(subset, superset, 9, t)
}

func TestSubsetShortCircuit(t *testing.T) {
	type recursive struct {
		Name     string
		Extra    bool
		Infinity *recursive
	}
	subset := &recursive{"ohai", false, nil}
	subset.Infinity = subset
	superset := &recursive{"ohai", true, nil}
	superset.Infinity = superset

	// without effective short circuit, this will overflow stack or timeout
	testN(subset, superset, 2, t)
}
