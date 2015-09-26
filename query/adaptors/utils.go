package adaptors

import (
	"regexp"

	"github.com/influx6/dataquery/parser"
	"github.com/influx6/ds"
)

// ParserNodeEvaluator provides a function type of ParserNode evluation
type ParserNodeEvaluator func(*parser.ParseNode) bool

// WrapNodeEvaluator wraps a ds.NodeEvaluator into a ParserNodeEvaluator
func WrapNodeEvaluator(fx ParserNodeEvaluator) ds.EvaluateNode {
	return func(n ds.Nodes) bool {
		ps, ok := n.(*parser.ParseNode)
		if ok {
			return fx(ps)
		}
		return false
	}
}

// GetRoot returns the root node of a parser.graph
func GetRoot(gs ds.Graphs) (ds.Nodes, error) {
	// create a linear search so we can find the root Node using the NodeType parser.MODELROOT
	goo := ds.NewLinearGraphSearch(gs)

	//using the FindOne to get the node that is the only root of all the node paths using its NType
	return goo.FindOne(WrapNodeEvaluator(func(q *parser.ParseNode) bool {
		if q.NType == parser.MODELROOT {
			return true
		}
		return false
	}))

}

// DFGraph aka (Depth-First Graph) searches a parser graph for the root node for transversing and returns a depth-first iterator or an error if it failed
func DFGraph(gs ds.Graphs) (*ds.Transversor, error) {
	// create a linear search so we can find the root Node using the NodeType parser.MODELROOT
	uon, err := GetRoot(gs)
	// goo = nil
	//if there is an error,return it immediately
	if err != nil {
		return nil, err
	}

	//we create a depth-first iterator for the graph from the root node
	mo, err := ds.DepthFirstPreOrderIterator(nil, nil)

	if err != nil {
		return nil, err
	}

	//set the iterator to use the root node we found
	mo.Use(uon)

	return mo, nil
}

// BFGraph aka (Depth-First Graph) searches a parser graph for the root node for transversing and returns a depth-first iterator or an error if it failed
func BFGraph(gs ds.Graphs) (*ds.Transversor, error) {
	// create a linear search so we can find the root Node using the NodeType parser.MODELROOT
	uon, err := GetRoot(gs)
	// goo = nil
	//if there is an error,return it immediately
	if err != nil {
		return nil, err
	}

	//we create a depth-first iterator for the graph from the root node
	mo, err := ds.BreadthFirstPreOrderIterator(nil, nil)

	if err != nil {
		return nil, err
	}

	//set the iterator to use the root node we found
	mo.Use(uon)

	return mo, nil
}

// FindMatch matches against all items in an array returning a true/false
func FindMatch(set []string, key string) (int, bool) {
	var found bool
	var pos int
	//search through the array
	for po, so := range set {
		if key != so {
			continue
		}
		pos = po
		found = true
		break
	}

	return pos, found
}

// BuildInterfacePoints creates an array of interface pointers
func BuildInterfacePoints(size int) []interface{} {
	mu := make([]interface{}, size)

	for n := range mu {
		mu[n] = new(interface{})
	}

	return mu
}

// UnbuildInterfacePoints unwraps an array of interface pointers
func UnbuildInterfacePoints(mu []interface{}) []interface{} {
	mus := make([]interface{}, len(mu))

	for n := range mu {
		mus[n] = *(mu[n].(*interface{}))
	}

	return mus
}

// UnbuildInterfaceList unwraps a list of interfaces pointers into inteface values
func UnbuildInterfaceList(mu [][]interface{}) [][]interface{} {
	var bus [][]interface{}
	for _, mo := range mu {
		bus = append(bus, UnbuildInterfacePoints(mo))
	}
	return bus
}

var onlySpaces = regexp.MustCompile(`^\s+$`)

//CleanHouse cleans all spaces out of list of strings
func CleanHouse(ls []string) []string {
	var clean []string
	for _, v := range ls {
		if v != "" && !onlySpaces.MatchString(v) {
			clean = append(clean, v)
		}
	}

	return clean
}
