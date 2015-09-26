package sql

import (
	"errors"
	"fmt"
	"strings"

	"github.com/influx6/dataquery/parser"
)

//SQLSimpleSelect defines a standard sql query format
const SQLSimpleSelect = `SELECT {{columns}} FROM {{tables}} WHERE {{clauses}};`

// ErrNoValue returns when a collector has no value
var ErrNoValue = errors.New("Collector has no value")

//TemplatesQueries provides a global handler for sql query templates
var TemplatesQueries = parser.NewOPFactory()

// RelQueries provides query formatters for relation tags or keys (key: id), providing a custom but simple of passing specifc special keys that provide context
var RelQueries = parser.NewOPFactory()

// AddSQLQueryHandlers adds handlers for sql query parameters to a OPFactory
func AddSQLQueryHandlers(op *parser.OPFactory) {
	//these are used to generate the where clause statement section
	op.Add("id", func(name string, c parser.Collector) ([]string, error) {

		if !c.Has("value") {
			return nil, ErrNoValue
		}

		val := c.Get("value")
		return []string{fmt.Sprintf("{{table}}.%s = %s", name, val)}, nil
	})

	op.Add("gte", func(name string, c parser.Collector) ([]string, error) {

		if !c.Has("value") {
			return nil, ErrNoValue
		}

		val := c.Get("value").(int)
		return []string{fmt.Sprintf("{{table}}.%s => %d", name, val)}, nil
	})

	op.Add("gt", func(name string, c parser.Collector) ([]string, error) {

		if !c.Has("value") {
			return nil, ErrNoValue
		}

		val := c.Get("value").(int)
		return []string{fmt.Sprintf("{{table}}.%s > %d", name, val)}, nil
	})

	op.Add("lte", func(name string, c parser.Collector) ([]string, error) {

		if !c.Has("value") {
			return nil, ErrNoValue
		}

		val := c.Get("value").(int)
		return []string{fmt.Sprintf("{{table}}.%s <= %d", name, val)}, nil
	})

	op.Add("lt", func(name string, c parser.Collector) ([]string, error) {

		if !c.Has("value") {
			return nil, ErrNoValue
		}

		val := c.Get("value").(int)
		return []string{fmt.Sprintf("{{table}}.%s < %d", name, val)}, nil
	})

	op.Add("in", func(name string, c parser.Collector) ([]string, error) {

		if !c.Has("range") {
			return nil, ErrNoValue
		}

		var inwords []string
		ranges := c.Get("range").([]string)

		for _, ins := range ranges {
			inwords = append(inwords, fmt.Sprintf("{{table}}.%s = %s", name, ins))
		}

		return []string{strings.Join(inwords, "\nOR ")}, nil
	})

	op.Add("is", func(name string, c parser.Collector) ([]string, error) {

		if !c.Has("value") {
			return nil, ErrNoValue
		}

		val := c.Get("value")

		return []string{fmt.Sprintf("{{table}}.%s = %s", name, val)}, nil
	})

	op.Add("isnot", func(name string, c parser.Collector) ([]string, error) {

		if !c.Has("value") {
			return nil, ErrNoValue
		}

		val := c.Get("value")
		return []string{fmt.Sprintf("{{table}}.%s != %s", name, val)}, nil
	})

	op.Add("range", func(name string, c parser.Collector) ([]string, error) {

		if !c.Has("max") && !c.Has("min") {
			return nil, ErrNoValue
		}

		max := c.Get("max")
		maxso := fmt.Sprintf("{{table}}.%s => %d", name, max)

		min := c.Get("min")
		minso := fmt.Sprintf("{{table}}.%s <= %d", name, min)

		// orange := strings.Join([]string{minso, maxso}, "\nOR\n")

		return []string{minso, maxso}, nil
	})
}

// AddSQLRelHandlers provides handlers for sql special keys tags
func AddSQLRelHandlers(op *parser.OPFactory) {
	op.Add("with", func(name string, c parser.Collector) ([]string, error) {
		if !c.Has("value") {
			return nil, ErrNoValue
		}

		val := c.Get("value").([]string)

		return []string{fmt.Sprintf("{{table}}.%s = {{parentTable}}.%s", val[0], val[1])}, nil
	})
}

func init() {
	AddSQLQueryHandlers(TemplatesQueries)
	AddSQLRelHandlers(RelQueries)
}
