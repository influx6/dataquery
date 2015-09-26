package sql

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/influx6/data/query/adaptors"
	"github.com/influx6/data/query/parser"
	"github.com/influx6/ds"
	"github.com/influx6/flux"
)

// specialKeys are import keywords that provide context for sql tables within queries
var specialKeys = []string{"with", "rel"}
var relationKey = "with"

// Table represent a table of a single record to be queried
type Table struct {
	Name       string
	Key        string
	Parent     string
	PKey       string
	Attrs      []string
	Columns    []string
	Conditions []string
	Orders     []string
	Node       *parser.ParseNode
	Graph      ds.Graphs
}

// Tables represent an array of SQLTable
type Tables []*Table

// TableBuilder provides a simple sql parser
func TableBuilder(op, specs *parser.OPFactory) flux.Reactor {
	return adaptors.QueryAdaptor(func(r flux.Reactor, gs ds.Graphs) {
		mo, err := adaptors.DFGraph(gs)

		if err != nil {
			r.ReplyError(err)
			return
		}

		recordSize := gs.Length()
		_ = recordSize

		var tables Tables

		for mo.Next() == nil {
			uo := mo.Node().(*parser.ParseNode)
			//create a table for this record
			table := &Table{
				Key:    uo.Key,
				Name:   uo.Name(),
				Parent: uo.Parent,
				PKey:   uo.PKey,
				Attrs:  uo.Attr.All(),
				Node:   uo,
				Graph:  gs,
			}

			tables = append(tables, table)

			rules := uo.Rules

			if table.Parent != "" {
				if !rules.Has(relationKey) {
					r.ReplyError(fmt.Errorf("Query for '%s' is a subroot/child of '%s' and needs a '%s(with: [childkey parentkey])' in its conditions for proper evaluation e.g '%s(with: [user_id id])'", table.Name, table.Parent, table.Name, table.Name))
					// col := parser.NewCondition("with")
					// col.Set("value", []string{fmt.Sprintf("%s_id", strings.ToLower(table.Parent)), "id"})
					// rules.Set(relationKey, []parser.Collector{col})
					return
				}
			}

			for _, val := range specialKeys {
				if !rules.Has(val) {
					continue
				}

				co, err := rules.Get(val)

				if err != nil && len(co) <= 0 {
					return
				}

				cod := co[0]

				// table.Keys[val] = co[0].Get("value")
				if kos, err := specs.Process(val, val, cod); err == nil {
					table.Conditions = append(table.Conditions, kos...)
				}
				rules.Remove(val)

			}

			rules.EachCondition(func(name string, c parser.Collector, stop func()) {
				co, err := op.Process(c.Get("type").(string), name, c)

				if err != nil {
					r.ReplyError(err)
					stop()
				}

				table.Conditions = append(table.Conditions, co...)
			})

			records := uo.Records
			//add the record to the column list
			table.Columns = append(table.Columns, records.Keys()...)

			//process the records constraints
			records.EachCondition(func(name string, c parser.Collector, stop func()) {
				co, err := op.Process(c.Get("type").(string), name, c)

				if err != nil {
					r.ReplyError(err)
					stop()
				}
				table.Conditions = append(table.Conditions, co...)
			})
		}
		//deliver the table for building
		r.Reply(tables)
	})
}

//ErrInvalidTableData represent the error when the data type does not match the Tables type
var ErrInvalidTableData = errors.New("Data type not []*Tables")

// TableInfo defines the range of rows and columns that a table respectfully has when decifying the result of a query
type TableInfo struct {
	Alias       string
	ParentAlias string
	Name        string
	Parent      string
	Begin, End  int
	Columns     []string
	Node        *parser.ParseNode
	Graph       ds.Graphs
}

// TableMeta defines a map of TableInfo
type TableMeta map[string]*TableInfo

// Statement represent a properly passed sql SqlStatement
type Statement struct {
	Query   string
	Tables  TableMeta
	Columns int
	Data    [][]interface{}
	Graph   ds.Graphs
}

//TableParser is a reactor that takes a array of *Tables and generates the corresponding sql statement
func TableParser() flux.Reactor {
	return flux.Reactive(func(r flux.Reactor, err error, data interface{}) {
		if err != nil {
			r.ReplyError(err)
			return
		}

		var tables Tables
		var ok bool

		tables, ok = data.(Tables)

		if !ok {
			r.ReplyError(ErrInvalidTableData)
			return
		}

		var tableNames []string
		var tableColumns []string
		var tableWheres []string
		var tableMeta = make(TableMeta)
		var lastColumSize = 0
		var graph ds.Graphs

		for _, table := range tables {

			if graph == nil {
				graph = table.Graph
			}

			//add the tables names into the array and ensure to use aliases format "TALBENAME tablename"
			tableNames = append(tableNames, fmt.Sprintf("%s %s", strings.ToUpper(table.Name), table.Key))

			//loop through each column name and append talbe alias,add the column names for the 'from' clause
			for _, coname := range table.Columns {
				tableColumns = append(tableColumns, fmt.Sprintf("%s.%s", table.Key, coname))
			}

			//collect table info for particular table
			tableMeta[table.Name] = &TableInfo{
				Alias:       table.Key,
				ParentAlias: table.PKey,
				Name:        table.Name,
				Parent:      table.Parent,
				Columns:     table.Columns,
				Begin:       lastColumSize,
				End:         (lastColumSize + (len(table.Columns) - 1)),
				Node:        table.Node,
				Graph:       table.Graph,
			}

			lastColumSize = len(tableColumns)

			// for _,clo := range table.Conditions {
			// }
			//join the conditions of this table with a AND
			clos := strings.Join(table.Conditions, "\nAND ")

			//replace both {{table}} and {{parentTable}} with the appropriate names/tags
			// log.Printf("setting alias:", table.Key, table.PKey, table.Name)
			clos = strings.Replace(clos, "{{table}}", table.Key, -1)
			clos = strings.Replace(clos, "{{parentTable}}", table.PKey, -1)

			//add this condition to the global where list
			tableWheres = append(tableWheres, clos)
		}

		var sqlst = SQLSimpleSelect

		sqlst = strings.Replace(sqlst, "{{columns}}", strings.Join(tableColumns, ", "), -1)
		sqlst = strings.Replace(sqlst, "{{tables}}", strings.Join(tableNames, ", "), -1)

		//clean where clauses of an empty strings or only spaces
		tableWheres = adaptors.CleanHouse(tableWheres)

		if len(tableWheres) < 2 {
			sqlst = strings.Replace(sqlst, "{{clauses}}", strings.Join(tableWheres, " "), -1)
		} else {
			sqlst = strings.Replace(sqlst, "{{clauses}}", strings.Join(tableWheres, "\nAND "), -1)
		}

		// log.Printf("SQL: %s", sqlst)
		r.Reply(&Statement{
			Query:   sqlst,
			Tables:  tableMeta,
			Columns: len(tableColumns),
			Graph:   graph,
		})
	})
}

//ErrInvalidTableData represent the error when the data type does not match the Tables type
var ErrInvalidStatementType = errors.New("Data type not *Statement")

//DbExecutor returns a reactor that takes a sql.Db for execution of queries
func DbExecutor(db *sql.DB) flux.Reactor {
	return flux.Reactive(func(r flux.Reactor, err error, d interface{}) {
		if err != nil {
			r.ReplyError(err)
			return
		}

		var stl *Statement
		var ok bool

		if stl, ok = d.(*Statement); !ok {
			r.ReplyError(ErrInvalidStatementType)
			return
		}

		rows, err := db.Query(stl.Query)

		if err != nil {
			r.ReplyError(err)
			return
		}

		var datarows [][]interface{}

		defer rows.Close()

		for rows.Next() {
			bu := adaptors.BuildInterfacePoints(stl.Columns)

			err := rows.Scan(bu...)

			if err != nil {
				r.ReplyError(err)
				return
			}

			datarows = append(datarows, bu)
		}

		stl.Data = adaptors.UnbuildInterfaceList(datarows)
		r.Reply(stl)
	})
}

//TableSection represents a single data composition tree per sql record row representing the retrieved data
type TableSection map[string]interface{}

// RecordBlock represent a tree of records of a single row
type RecordBlock map[string]TableSection

// TableBlock represents a list of TableSection
type TableBlock []RecordBlock

// JSONBuilder produces a flux.Reactor for turning a list of sql data with corresponding tableinfo's to build a json structure
func JSONBuilder() flux.Reactor {
	return flux.Reactive(func(r flux.Reactor, err error, d interface{}) {
		if err != nil {
			r.ReplyError(err)
			return
		}

		var stl *Statement
		var ok bool

		if stl, ok = d.(*Statement); !ok {
			r.ReplyError(ErrInvalidStatementType)
			return
		}

		for _, blck := range stl.Data {
			func(block []interface{}) {
				// orecord := make(RecordBlock)
				// records = append(records, orecord)
				for _, ifo := range stl.Tables {
					func(info *TableInfo) {
						section := make(TableSection)
						max := info.End - info.Begin

						for j := 0; j <= max; j++ {
							func(ind int) {
								col := info.Columns[ind]
								section[col] = block[info.Begin+ind]
							}(j)
						}

						// info.Node.Result = section
						info.Node.Result = append(info.Node.Result, section)
						// orecord[info.Alias] = section
					}(ifo)
				}

			}(blck)
		}

		mo, err := adaptors.BFGraph(stl.Graph)

		if err != nil {
			r.ReplyError(err)
			return
		}

		var roots = make(map[string]*parser.ParseNode)
		var tree = make(map[string]interface{})
		var root *parser.ParseNode

		for mo.Next() == nil {
			no := mo.Node().(*parser.ParseNode)

			if root == nil {
				root = no
			}

			if _, ok := roots[no.Key]; !ok {
				roots[no.Key] = no
			}

			rod, ok := roots[no.PKey]

			if !ok {
				continue
			}

			for n, rorec := range rod.Result {
				rorec[no.Name()] = no.Result[n]
			}
		}

		// res := root.Result
		tree[root.Name()] = root.Result
		stl = nil
		roots = nil
		root = nil

		r.Reply(tree)
	})
}

// BuildPreQuero generates a sql parser without any attachement to the sql record query Reactor
func BuildPreQuero(op, sp *parser.OPFactory, ds *parser.InspectionFactory) flux.Reactor {
	co := adaptors.ChunkParser(ds)
	co.Bind(TableBuilder(op, sp), true)
	co.Bind(TableParser(), true)
	return co
}

// BuildQuero generates a full sql query parser and table parser for instance use
func BuildQuero(db *sql.DB, op, sp *parser.OPFactory, ds *parser.InspectionFactory) flux.Reactor {
	co := BuildPreQuero(op, sp, ds)
	co.Bind(DbExecutor(db), true)
	co.Bind(JSONBuilder(), true)
	return co
}

// BasicQueroEngine produces an engine with the default query handlers
func BasicQueroEngine() flux.Reactor {
	return BuildPreQuero(TemplatesQueries, RelQueries, parser.DefaultInspectionFactory)
}

// Quero returns a new instance build of a complete sql query handler and parser using the defaultly provide query formatters and collectors
func Quero(db *sql.DB) flux.Reactor {
	return BuildQuero(db, TemplatesQueries, RelQueries, parser.DefaultInspectionFactory)
}

// QueroJSON attaches a json reactor that marshalls all response out as a json string
func QueroJSON(db *sql.DB) flux.Reactor {
	qo := Quero(db)
	qo.Bind(flux.JSONReactor(), true)
	return qo
}
