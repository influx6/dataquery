package datamodel

import "database/sql"

//Sql provides a basic Model handler for generating a model save /update or delete of the record

//SQLSave takes a Model and performs a save operation on every request coming to it
type SQLSave struct {
	model *Models
	db    *sql.DB
}

// NewSQLSave returns a new sql saver
func NewSQLSave(db *sql.DB, model *Models) *SQLSave {
	return &SQLSave{
		model: model,
		db:    db,
	}
}

// Save performs the save operation on the model
func (s *SQLSave) Save(data ModelData) {

}
