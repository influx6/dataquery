#query
query is all about how we get data, it builds on the core ideals of simplicity and meeting a single goal by providing a flexibile and extensible overlay for querying databases. It's about simplify the way in which we get our data from these database endpoints.

##Ideas
  These are simple ideas that power query underneath and provide the basis for its operations

  - Adaptors

  Adaptors are custom graph consumers, that take the parsed graph of the giving query and generate the corresponding query syntax of the corresponding database and process the necessary data to be retrieved into a usable format as output (eg json)

  - ParseGraph:

  A parsegraph is a depth graph that contains a more defined set of data requirements needed by the database adaptor which allow simple generation of the corresponding database query syntax


# Queries

  - Basic Query

  A basic example of what a query looks like is really a simplified json structure with the values removed, it provides a simple and express hierarchy of the data and the constraint and conditions which each applies. Underneath, each adaptor provides these constraint processors that create the corresponding database syntax to match the query. Queries are dealt with by the parser in a singular approach i.e the query information within this format 'record_name(..){...}'  is a single graph containing the root record with the connected child records,any addition queries are single out and processed as single entities for retrieval, so the parser graph is small,compact and expresses a single data collection and its related parts

	      ```go

        //A single query
				query := `
					user(id: 4000){
					  name,
					  state,
					  address,
					  skills(range: 30..100),
					  age(lt:30, gte:40),
					  day(isnot: wednesday),
					  photos(with: [user_id id]){
					    day,
					    fax,
					  },
					  face(is: 20),
					}
				`

			```

  - Compound Query

  Since query treats all information between the 'record_name(){...}' format as a single query, this allows compound queries be build and treated as seperate query graphs by the parser, it pieces out each sub-query according to the governing format rules then sends each query graph for processes to the connected adaptor, basically batching query requests

	      ```go

        //A compound query, where 'user(){}' and 'comments(){}' are seperate queries
				query := `{
					user(){
					  name,
					  state,
					  address,
					  photos(with: [user_id id]){
					    url,
					  },
					},
					comments(){
					  date,
					  user_id,
					  email,
					  photo(with: [comment_id id]){
              comment_id,
					    url,
					  },
					},
        }
				`
			```

# Example

  - MySql (Standard SQL Adaptor)

   Included with the base library is the adaptor for a sql related database (generally those using mysql govererned syntax rules)

   ```go

   package sql

   import (
   	"database/sql"
   	"sync"
   	"testing"

   	_ "github.com/go-sql-driver/mysql"

   	sqlap "github.com/influx6/data/query/adaptors/sql"
   	"github.com/influx6/data/query/adaptors"
   	"github.com/influx6/data/query/parser"
   	"github.com/influx6/flux"
   )

   	db, err := sql.Open("mysql", "root:@tcp(localhost:3306)/test")

   	defer db.Close()

   	qo := sqlap.Quero(db)

   	qo.React(func(r flux.Reactor, err error, result interface{}) {
   	  // result is actually a map[string]interface{} of the giving query
      log.Println("data:",result)
   	}, true)

   	qo.Send(`
      users(){
        id,
        name,
        age,
        street,
        stamp,
        photos(with: [user_id id]){
          url,
          user_id,
          id,
        },
      }
    `)

   ```
#License

    .  MIT License
