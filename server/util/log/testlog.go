package log

import (
	"bytes"
	"errors"
	"fmt"
)

const (
	_none = iota
	_error
)

func Panic(reason interface{}, val ...interface{}) {
	panic(errorToString(_none, reason, val))

}

func Error(reason interface{}, val ...interface{}) error {
	err := errorToString(_error, reason, val)
	return errors.New(err)
}

func ErrorP(reason interface{}, val ...interface{}) {
	err := errorToString(_none, reason, val)
	fmt.Println(err)
}

func errorToString(isPanic uint, reason interface{}, val ...interface{}) string {
	var buffer bytes.Buffer
	switch isPanic {
	case _none:
		buffer.WriteString("")
	case _error:
		buffer.WriteString("Error : ")
	default:

	}
	buffer.WriteString(reason.(string))
	buffer.WriteString(" - ")
	return fmt.Sprint(buffer.String(), val)
}

//todo : error types, file stream.
