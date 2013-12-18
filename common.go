package common

import (
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"reflect"
)

// init creates the interface map to allow packages to add to the register
func init() {
	interfaceMap = make(map[string]interface{})
	//	isPtr = make(map[string]bool)
}

// DoesNotMatch is an error returned if the test marshal and unmarshal
// does not pass
var DoesNotMatch error = errors.New("Does not match")

// InterfaceTestMarshalAndUnmarshal tests that the type in the interface
// is correctly marshaled and unmarshaled. Returns an error if they do
// not match
func InterfaceTestMarshalAndUnmarshal(l interface{}) error {
	l1 := &InterfaceMarshaler{I: l}
	b, err := l1.MarshalJSON()
	if err != nil {
		return err
	}
	l2 := &InterfaceMarshaler{}
	err = l2.UnmarshalJSON(b)
	if err != nil {
		return err
	}
	if !reflect.DeepEqual(l, l2.I) {
		return DoesNotMatch
	}
	return nil
}

// interfaceMap helps encoding and decoding interfaces
var interfaceMap map[string]interface{}

// registerString converts the input type to the string
// key in interfaceMap
func registerString(i interface{}) string {
	str := interfaceFullTypename(i)
	// See if it's a pointer, and if so append a * at the end
	if reflect.ValueOf(i).Kind() == reflect.Ptr {
		str += "*"
	}

	return str
}

// Register logs an interface type to allow interaction with JSON.
func Register(i interface{}) {
	str := registerString(i)
	_, ok := interfaceMap[str]
	if ok {
		panic("nnet/common interface type already registered")
	}

	isPtr := reflect.ValueOf(i).Kind() == reflect.Ptr

	var newVal interface{}
	var tmp interface{}
	if isPtr {
		tmp = reflect.ValueOf(i).Elem().Interface()
	} else {
		tmp = i
	}
	newVal = reflect.New(reflect.TypeOf(tmp)).Elem().Interface()

	// Either way, save a real value

	//TODO: Add in something where the types aren't copied for the *
	interfaceMap[str] = newVal

	//interfaceMap[str] = i
}

// NotRegistered is retured if the type is not registered
type NotRegistered struct {
	Type string
}

func (n *NotRegistered) Error() string {
	return fmt.Sprintf("common: type %s not registered", n.Type)
}

func ptrFromString(str string) (interface{}, bool, error) {
	val, ok := interfaceMap[str]
	if !ok {
		return nil, false, &NotRegistered{Type: str}
	}
	isPtr := str[len(str)-1:len(str)] == "*"

	return reflect.New(reflect.TypeOf(val)).Interface(), isPtr, nil
}

// FromString returns a copy of the registered type
func FromString(str string) (interface{}, error) {
	val, isPtr, err := ptrFromString(str)
	if err != nil {
		return nil, err
	}
	/*
		// Make a copy of that type, and then dereference it
		var newVal interface{}
		if isPtr {
			newVal = reflect.New(reflect.TypeOf(val)).Interface()
		} else {
			newVal = reflect.New(reflect.TypeOf(val)).Elem().Interface()
		}
	*/
	var newVal interface{}
	newVal = val
	if !isPtr {
		newVal = reflect.ValueOf(val).Elem().Interface()
	}
	return newVal, nil
}

// NotInPackage is an error which signifies the type is
// not in the package. Normally used with marshaling and
// unmarshaling
var NotInPackage = errors.New("NotInPackage")

// UnmarshalMismatch is an error type used when unmarshaling the specific
// activators in this package.
type UnmarshalMismatch struct {
	Expected string
	Received string
}

// Error is so UnmarshalMismatch implements the error interface
func (u UnmarshalMismatch) Error() string {
	return "Unmarshal string mismatch. Expected: " + u.Expected + " Received: " + u.Received
}

// InterfaceMarshaler is a type to help the marshaling and unmarshaling
// of interface values. Types marshaled and unmarshaled with InterfaceMarshaler
// must be first be registered using Register(). It uses a similar idea
// to gob.
type InterfaceMarshaler struct {
	I interface{}
}

type lossMarshaler struct {
	Type  string
	Value interface{}
}

type typeUnmarshaler struct {
	Type  string
	Value json.RawMessage
}

func (l *InterfaceMarshaler) MarshalJSON() ([]byte, error) {
	// Check that the type has been registered
	name := registerString(l.I)
	_, ok := interfaceMap[name]
	if !ok {
		return nil, &NotRegistered{Type: name}
	}
	marshaler := &lossMarshaler{
		Type:  name,
		Value: l.I,
	}
	return json.Marshal(marshaler)
}
func (l *InterfaceMarshaler) UnmarshalJSON(data []byte) error {
	// First, unmarshal the type
	t := &typeUnmarshaler{}
	err := json.Unmarshal(data, t)
	if err != nil {
		return err
	}
	// Get a pointer to the right type
	val, isPtr, err := ptrFromString(t.Type)
	if err != nil {
		return errors.New("nnet/common error unmarshaling: " + err.Error())
	}

	nextdata := []byte(t.Value)
	// Unmarshal the (pointer to the) value
	err = json.Unmarshal(nextdata, val)
	if err != nil {
		return err
	}

	// If we don't want an interface, return a pointer to it
	if !isPtr {
		val = reflect.ValueOf(val).Elem().Interface()
	}

	l.I = val
	return nil
}

// interfaceFullTypename returns the full name of the provided type
// as packagename/typename
func interfaceFullTypename(i interface{}) string {
	pkgpath, pkgname := interfaceLocation(i)
	return filepath.Join(pkgpath, pkgname)
}

// interfaceLocation finds the package path and typename of the given
// interface
func interfaceLocation(i interface{}) (pkgpath string, name string) {
	if reflect.ValueOf(i).Kind() == reflect.Ptr {
		pkgpath = reflect.ValueOf(i).Elem().Type().PkgPath()
		name = reflect.ValueOf(i).Elem().Type().Name()
	} else {
		pkgpath = reflect.TypeOf(i).PkgPath()
		name = reflect.TypeOf(i).Name()
	}
	return
}
