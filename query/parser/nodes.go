package parser

import (
	"errors"
	"fmt"
	"sync"
)

//HandlerNotFoundMessage is a message returned when a parso handler is not found
const HandlerNotFoundMessage = "Error: Query Parser for '%s' not Found!"

//InspectionNotFoundMessage provides error for not found condition makers
const InspectionNotFoundMessage = ("Inspector '%s' Not Found!")

//ErrInspectionNotFound provides error for not found condition makers
var ErrInspectionNotFound = errors.New("Inspector Not Found!")

// ParseFx defines a op caller for a custom colletor
type ParseFx func(string, Collector) ([]string, error)

// Parso provides a parser for custom collectors
type Parso struct {
	tag string
	fx  ParseFx
}

// CreateQueryParser creates a new parsor instance for handling query types
func CreateQueryParser(tag string, fx ParseFx) *Parso {
	po := Parso{tag: tag, fx: fx}
	return &po
}

// OPFactory provides a factory of dealing with special query parameters in the parser
type OPFactory struct {
	factory map[string]*Parso
	rw      sync.RWMutex
}

// NewOPFactory returns a new OPFactory instance
func NewOPFactory() *OPFactory {
	op := OPFactory{factory: make(map[string]*Parso)}
	return &op
}

// Has returns true/false if a tag already exists
func (op *OPFactory) Has(tag string) bool {
	var res bool
	op.rw.RLock()
	{
		if _, ok := op.factory[tag]; ok {
			res = true
		}
	}
	op.rw.RUnlock()
	return res
}

// Process tags a tag and a Collector runs it against the specific Parso if it exists
func (op *OPFactory) Process(tag, id string, c Collector) ([]string, error) {
	po, err := op.Get(tag)
	if err != nil {
		return nil, err
	}
	return po.fx(id, c)
}

// Get returns the parso that belongs to the tag or an erro
func (op *OPFactory) Get(tag string) (*Parso, error) {
	if !op.Has(tag) {
		return nil, fmt.Errorf(HandlerNotFoundMessage, tag)
	}
	var po *Parso
	op.rw.RLock()
	{
		po = op.factory[tag]
	}
	op.rw.RUnlock()

	return po, nil
}

// Remove removes an existing special handler for query parsing
func (op *OPFactory) Remove(tag string) {
	if !op.Has(tag) {
		return
	}
	op.rw.Lock()
	{
		delete(op.factory, tag)
	}
	op.rw.Unlock()
}

// Add adds a new special handler for query parsing
func (op *OPFactory) Add(tag string, fx ParseFx) {
	if op.Has(tag) {
		return
	}
	op.rw.Lock()
	{
		op.factory[tag] = CreateQueryParser(tag, fx)
	}
	op.rw.Unlock()
}

//ValidFx defines a function type of function validators
type ValidFx func(data string) (Collector, error)

//InspectionFactory provides a factory of dealing with special query parameters in the parser
type InspectionFactory struct {
	factory map[string]*Inspector
	rw      sync.RWMutex
}

//NewInspectionFactory returns a new InspectionFactory
func NewInspectionFactory() *InspectionFactory {
	ios := InspectionFactory{
		factory: make(map[string]*Inspector),
	}
	return &ios
}

//Find lets you add a new condition maker
func (c *InspectionFactory) Find(tag string) (*Inspector, error) {
	c.rw.RLock()
	in, ok := c.factory[tag]
	c.rw.RUnlock()

	if !ok {
		return nil, fmt.Errorf(InspectionNotFoundMessage, tag)
	}

	return in, nil
}

//Has returns true if the inspector exists
func (c *InspectionFactory) Has(tag string) bool {
	c.rw.RLock()
	_, ok := c.factory[tag]
	c.rw.RUnlock()
	return ok
}

//Register lets you add a new condition maker
func (c *InspectionFactory) Register(tag string, cx ValidFx) {
	c.rw.Lock()
	defer c.rw.Unlock()
	c.factory[tag] = NewInspector(tag, cx)
}

//Deregister lets you add a new condition maker
func (c *InspectionFactory) Deregister(tag string) {
	if !c.Has(tag) {
		return
	}
	c.rw.Lock()
	defer c.rw.Unlock()
	delete(c.factory, tag)
}

//Inspector provides a baseline for other validators
type Inspector struct {
	tag string
	fx  ValidFx
}

//NewInspector returns a new validator instance
func NewInspector(tag string, fx ValidFx) *Inspector {
	return &Inspector{tag: tag, fx: fx}
}

//Create validates a set of data against a provided inspector
func (v *Inspector) Create(data string) (Collector, error) {
	return v.fx(data)
}

//Keyword provides the data set for the conditions
func (v *Inspector) Keyword() string {
	return v.tag
}

//NewCondition makes a new condition foruse
func NewCondition(ctype string) Collector {
	col := NewCollector()
	col.Set("type", ctype)
	return col
}
