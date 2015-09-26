package parser

import "errors"

// ErrNotFound a general error for when a state or item is not found
var ErrNotFound = errors.New("Item not Found")

// ErrInvalidCollector returns when a collector has no value
var ErrInvalidCollector = errors.New("Collector Invalid: Either lacks required keys or value types")

//Collector defines a typ of map string
type Collector map[string]interface{}

//NewCollector returns a new collector instance
func NewCollector() Collector {
	return make(Collector)
}

//Clone makes a new clone of this collector
func (c Collector) Clone() Collector {
	col := make(Collector)
	col.Copy(c)
	return col
}

//Remove deletes a key:value pair
func (c Collector) Remove(k string) {
	if c.Has(k) {
		delete(c, k)
	}
}

//Set puts a specific key:value into the collector
func (c Collector) Set(k string, v interface{}) {
	c[k] = v
}

//Copy copies the map into the collector
func (c Collector) Copy(m map[string]interface{}) {
	for v, k := range m {
		c.Set(v, k)
	}
}

// StringEachFunc defines the each function type for a collector
type StringEachFunc func(interface{}, string, func())

//Each iterates through all items in the collector
func (c Collector) Each(fx StringEachFunc) {
	var state bool
	for k, v := range c {
		if state {
			break
		}

		fx(v, k, func() {
			state = true
		})
	}
}

//Keys return the keys of the Collector
func (c Collector) Keys() []string {
	var keys []string
	c.Each(func(_ interface{}, k string, _ func()) {
		keys = append(keys, k)
	})
	return keys
}

//Get returns the value with the key
func (c Collector) Get(k string) interface{} {
	return c[k]
}

//Has returns if a key exists
func (c Collector) Has(k string) bool {
	_, ok := c[k]
	return ok
}

//HasMatch checks if key and value exists and are matching
func (c Collector) HasMatch(k string, v interface{}) bool {
	if c.Has(k) {
		return c.Get(k) == v
	}
	return false
}

//Clear clears the collector
func (c Collector) Clear() {
	for k := range c {
		delete(c, k)
	}
}

// Collectors define sets of application options used by a parser node
type Collectors struct {
	Collector
}

//NewCollectors returns a new instance of collectors
func NewCollectors() *Collectors {
	co := Collectors{NewCollector()}
	return &co
}

// Set sets a Collector to a key
func (c *Collectors) Set(k string, co []Collector) {
	c.Collector.Set(k, co)
}

// Get returns a Collector or an error if not found by a key
func (c *Collectors) Get(k string) ([]Collector, error) {
	if c.Collector.Has(k) {
		return c.Collector.Get(k).([]Collector), nil
	}
	return nil, ErrNotFound
}

//EachCollector provides a iterator function signature for iterator through a map of collectors
type EachCollector func([]Collector, string, func())

//Each iterates through all items in the collector
func (c *Collectors) Each(fx EachCollector) {
	c.Collector.Each(func(v interface{}, k string, stop func()) {
		fx(v.([]Collector), k, stop)
	})
}

//EachConditionHandler provides a type for rules iteration
type EachConditionHandler func(string, Collector, func())

//EachCondition builds a custom function for iterator a Collector used for rules
func (c *Collectors) EachCondition(fx EachConditionHandler) {
	c.Each(func(vcol []Collector, k string, so func()) {
		for _, vc := range vcol {
			fx(k, vc, so)
		}
	})
}
