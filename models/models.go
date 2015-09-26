package datamodel

import (
	"errors"
	"fmt"
	"reflect"
	"sync"
)

// ErrTypeMisMatch is returned when the type fails to match
var ErrTypeMisMatch = errors.New("Value does not match type")

// ModelAttr defines a singular attribute
type ModelAttr struct {
	Name  string
	Tag   string
	Tags  string
	Base  interface{}
	dtype reflect.Type
}

// Type returns the ModelAttr type value
func (m *ModelAttr) Type() reflect.Type {
	return m.dtype
}

// NewModelAttrWith defines a model attribute and its type
func NewModelAttrWith(name, tag string, bo interface{}, to reflect.Type) *ModelAttr {
	ma := ModelAttr{
		Name:  name,
		Tag:   tag,
		Base:  to,
		dtype: to,
	}
	// mot := reflect.TypeOf(to)
	//
	// if mot.Kind() == reflect.Ptr {
	// 	mot = mot.Elem()
	// }
	// ma.dtype = to

	return &ma
}

// NewModelAttr defines a model attribute and its type
func NewModelAttr(name, tag string, to interface{}) *ModelAttr {
	ma := ModelAttr{Name: name, Tag: tag, Base: to}

	mot := reflect.TypeOf(to)

	if mot.Kind() == reflect.Ptr {
		mot = mot.Elem()
	}

	ma.dtype = mot

	return &ma
}

// Validate validates if the value matches the model attribute type. We cant match individual content especially for slices,arrays or maps, so we can only do a one to one interface to interface type matching and not if the values are equal.
func (m *ModelAttr) Validate(val interface{}) error {
	mot := reflect.TypeOf(val)

	//is it a pointer,get the element
	if mot.Kind() == reflect.Ptr {
		mot = mot.Elem()
	}

	// switch m.dtype.Kind() {
	// case reflect.Array:
	// case reflect.Slice:
	// case reflect.Map:
	// 	mapo := reflect.ValueOf(val)
	//
	// 	if mapo.Kind() == reflect.Ptr {
	// 		mapo = mapo.Elem()
	// 	}
	//
	//
	//
	// default:
	//we must first if we are not dealing with a pointer or interface if so,use normal == equality checks and AssignableTo else we can do more by checking if it implements the type and the standard checks
	if mot.Kind() != reflect.Interface && mot.Kind() != reflect.Ptr {
		//are the types the same or can the type be interused
		if mot != m.dtype && !mot.AssignableTo(m.dtype) {
			return ErrTypeMisMatch
		}
	} else {
		//is this a implements relationship? or Assignable or equality?
		if !mot.Implements(m.dtype) && !mot.AssignableTo(m.dtype) && mot != m.dtype {
			return ErrTypeMisMatch
		}
	}
	// }

	return nil
}

// ModelData defines the data and values of a model
type ModelData map[string]interface{}

// ModelAttrs defines a map defining a models attributes
type ModelAttrs map[string]*ModelAttr

// NewModelAttrs returns a new ModelAttrs with the schema loaded
func NewModelAttrs(schema ModelData) ModelAttrs {
	attrs := make(ModelAttrs)

	//loop through the map and generate the ModelAttrs
	for id, to := range schema {
		attrs[id] = NewModelAttr(id, id, to)
	}

	return attrs
}

// ErrNotStruct is returned when the value given is nota struct
var ErrNotStruct = errors.New("Value is not a struct type")

// NewModelStruct  create a attribute map of all the fields of a struct and if provided filters out the fields by the tag if the tag contains a non-empty string value, attributes that have '-' will be skipped automatically
func NewModelStruct(scheme interface{}, tag string) (ModelAttrs, error) {

	mod := reflect.ValueOf(scheme)

	//reset to the real value if its a pointer
	if mod.Kind() == reflect.Ptr {
		mod = mod.Elem()
	}

	//if its not a struct exit,we only deal with structs here
	if mod.Kind() != reflect.Struct {
		// panic("Value is not a struct")
		return nil, ErrNotStruct
	}

	// if mod.Kind() != reflect.Invalid {
	// 	panic("Value is invalid")
	// }

	podt := mod.Type()

	// if podt.Kind() != reflect.Struct {
	// 	panic("Value is not a struct")
	// }

	attrs := make(ModelAttrs)

	//loop through the types fields and collect the appropritate data
	for i := 0; i < podt.NumField(); i++ {
		fl := podt.Field(i)

		var tagval string

		//get the tag if its not an empty string else use the name
		if tag == "" {
			tagval = fl.Name
		} else {
			tagval = fl.Tag.Get(tag)

			//ok we are required to use model tag filtering, is the get a empty string ? is so skip
			if tagval == "" || tagval == "-" {
				continue
			}

			// // split it so we can get the correct value
			// to := strings.Split(tagval, ":")
			//
			// //if the length is inadequate lets use the Name instead
			// if len(to) <= 1 {
			// 	tagval = fl.Name
			// } else {
			// 	tagval = to[2]
			// }

		}

		//get the value by the name
		moval := mod.FieldByName(fl.Name)

		//if its we cant collect the interface value,it must be a hidden/unexported value so we skip
		if !moval.CanInterface() {
			continue
		}

		// var rv interface{}
		//
		// //can we get the
		// if moval.CanInterface() {
		// 	rv = moval.Interface()
		// }

		//modat  is the new modelattribute object
		modat := NewModelAttrWith(fl.Name, tagval, moval.Interface(), moval.Type())

		//assign the total tags also
		modat.Tags = string(fl.Tag)

		//add the modelAttr
		attrs[modat.Tag] = modat
	}

	return attrs, nil
}

// NewModelStructType  create a attribute map of all the fields of a struct type and if provided filters out the fields by the tag if the tag contains a non-empty string value, attributes that have '-' will be skipped automatically
func NewModelStructType(mod reflect.Type, tag string) (ModelAttrs, error) {

	//if its not a struct exit,we only deal with structs here
	if mod.Kind() != reflect.Struct {
		// panic("Value is not a struct")
		return nil, ErrNotStruct
	}

	attrs := make(ModelAttrs)

	//loop through the types fields and collect the appropritate data
	for i := 0; i < mod.NumField(); i++ {
		fl := mod.Field(i)

		var tagval string

		//get the tag if its not an empty string else use the name
		if tag == "" {
			tagval = fl.Name
		} else {
			tagval = fl.Tag.Get(tag)

			//ok we are required to use model tag filtering, is the get a empty string ? is so skip
			if tagval == "" || tagval == "-" {
				continue
			}

			// // split it so we can get the correct value
			// to := strings.Split(tagval, ":")
			//
			// //if the length is inadequate lets use the Name instead
			// if len(to) <= 1 {
			// 	tagval = fl.Name
			// } else {
			// 	tagval = to[2]
			// }

		}

		//get the value by the name

		// moval := mod.FieldByName(fl.Name)

		//if its we cant collect the interface value,it must be a hidden/unexported value so we skip
		// if !moval.CanInterface() {
		// 	continue
		// }

		// var rv interface{}
		//
		// //can we get the
		// if moval.CanInterface() {
		// 	rv = moval.Interface()
		// }

		//modat  is the new modelattribute object
		modat := NewModelAttrWith(fl.Name, tagval, nil, fl.Type)

		//assign the total tags also
		modat.Tags = string(fl.Tag)

		//add the modelAttr
		attrs[modat.Tag] = modat
	}

	return attrs, nil
}

// Models basic idea is to create a model that allows the use of adaptors to
//adding saving and updating of models into db
type Models struct {
	attrs  ModelAttrs
	schema ModelData
	data   interface{}
	ro     sync.Mutex
}

// NewModels return a new model with dynamic attributes provider as a map of default types
func NewModels(format ModelData) *Models {
	mo := Models{
		schema: format,
		attrs:  NewModelAttrs(format),
	}

	return &mo
}

// NewStructModels return a new model with its field generated into a dynamic attribute map for validating a map[string]interface{}
func NewStructModels(format interface{}, tag string) (*Models, error) {
	bo, err := NewModelStruct(format, tag)

	if err != nil {
		return nil, err
	}

	return &Models{
		data:  format,
		attrs: bo,
	}, nil
}

// ErrEmptyData is returned when the data map is empty
var ErrEmptyData = errors.New("Data supplied is an empty map")

// ErrDataOverload is returned when the data Data is more than expected
var ErrDataOverload = errors.New("Data is more than models attribute")

// Validate validates a set of data against the data and if any of the values fail it returns the value that failed and an error
func (m *Models) Validate(data ModelData) (string, error) {
	if len(data) <= 0 {
		return "", ErrEmptyData
	}

	if len(data) > len(m.attrs) {
		return "", ErrDataOverload
	}

	m.ro.Lock()
	defer m.ro.Unlock()

	for m, mo := range m.attrs {
		var tag string

		//since the content may use the tag,we check if its that tagname instead of the real Name
		_, dok := data[mo.Name]
		//collect incase its the tag name
		_, sok := data[mo.Tag]

		//is both invalid, then this value is not existent,skip
		//TODO: should we break and return error or just skip
		//i think best to skip to allow zero values
		if !dok && !sok {
			continue
		}

		//if its the real name use the real Name
		if dok {
			tag = mo.Name
		}

		//if its the tag name use the tag
		if sok {
			tag = mo.Tag
		}

		// if _, ok := data[m]; !ok {
		// 	if _,ok := data[mo.Tag]; !ok{
		// 		continue
		// 	}else{
		// 		tag = mo.Tag
		// 	}
		// }else{
		// 	tag =
		// }

		if err := mo.Validate(data[tag]); err != nil {
			return m, err
		}
	}

	return "", nil
}

//OperationError provides a custom error for operations
type OperationError struct {
	Tag     string
	Name    string
	Message string
}

// Error returns a string that match the error interface{}
func (o *OperationError) Error() string {
	return o.String()
}

// String returns the message of the error
func (o *OperationError) String() string {
	return fmt.Sprintf("Model(%s): Property: %s -> Message: %s", o.Tag, o.Name, o.Message)
}
