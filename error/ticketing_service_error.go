package error

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/magiconair/properties"
	"github.com/pkg/errors"
)

type CustomError struct {
	code int
	msg  string
	err  error
}

var (
	propertyFile []string
	props        *properties.Properties
)

func init() {
	messages := "./configs/error_messages.properties"
	ReadPropertyfile(messages)
}

func ReadPropertyfile(filepath string) {
	propertyFile = []string{filepath}
	props, _ = properties.LoadFiles(propertyFile, properties.UTF8, true)

}

// GetMessage is used to get the error messages from property file
func GetMessage(msgName string, params ...string) string {
	msg, ok := props.Get(msgName)
	if !ok {
		return props.MustGet("Message is not available in property file for key : " + msgName)
	}
	placehdrCnt := strings.Count(msg, "{")
	if placehdrCnt == len(params) {
		for i, val := range params {
			repalcestr := fmt.Sprintf("%s%d%s", "{", i, "}")
			msg = strings.Replace(msg, repalcestr, val, -1)
		}
	}
	return msg
}

func (e *CustomError) Error() string {
	return fmt.Sprintf(e.err.Error())
	//return fmt.Sprintf(e.msg)
}

//It returns the new custom error struct with wrapped error
func newCustomError(message string, code int) *CustomError {
	er := errors.New("")
	er = errors.Wrapf(er, message)
	err := &CustomError{code: code, msg: message, err: er}
	return err
}

func NewEmptyCustomError() *CustomError {
	err := &CustomError{}
	return err
}

func (e *CustomError) GetErrorCode() int {
	return e.code
}

func AddMessageToStack(er error, stackMsg string) *CustomError {
	//WithMessage is used to add contextual text information to underlying
	//error without attaching call stack
	//Apply this method for “wrapped error”only
	//(example: use this function after NewCustomError or WrapError fuction)
	wrappedError := errors.WithMessagef(er, stackMsg)
	if customErr, ok := er.(*CustomError); ok {
		return &CustomError{
			code: customErr.code,
			msg:  customErr.msg,
			err:  wrappedError,
		}
	}
	return NewEmptyCustomError()
}

// This function is to return error code and error message
func (e *CustomError) GetErMsgAndCode() (string, int) {
	if e != nil {
		return e.msg, e.code
	}
	return "", -1
}

//This method is used to convert error code in the form of string to int
func convertStringToInt(str string) (int, error) {
	err_code, err := strconv.Atoi(str)
	if err != nil {
		return -1, err //returning -1 for undefined error code
	}
	return err_code, nil
}

// This method is used when a new custom error has to be created
// It returns the new custom error
func NewCustomErr(errconst string, stackMessage string) *CustomError {
	err_code, _ := convertStringToInt(errconst)
	err := newCustomError(GetMessage(errconst), err_code)
	err = AddMessageToStack(err, stackMessage)
	return err
}

// This method is used to create a new custom err with dynamic message
func NewCustomErrWithMsg(customMsg string, customErrCode string, stackMessage string) *CustomError {
	err_code, _ := convertStringToInt(customErrCode)
	err := newCustomError(customMsg, err_code)
	err = AddMessageToStack(err, stackMessage)
	return err
}

// This method is used to add the error code and stack message to the existing error
// It returns the custom error with added error code and stack message
func AddErrCodeAndStackMsg(errconst string, er error, stackMessage string) *CustomError {
	err_code, _ := convertStringToInt(errconst)
	er = errors.WithMessagef(er, stackMessage)
	return &CustomError{err_code, GetMessage(errconst), er}
}

// This method is used to wrap the error
// It returns the custom error with wrapped error
func WrapError(errconst string, er error, stackMessage string) *CustomError {
	err_code, _ := convertStringToInt(errconst)
	wrappedError := errors.Wrapf(er, GetMessage(errconst))
	wrappedError = errors.WithMessagef(er, stackMessage)
	return &CustomError{code: err_code, msg: GetMessage(errconst), err: wrappedError}
}

// This method is used to create a new custom err with dynamic message
func NewCustomErrorWithCustomMsg(customMsg string, customErrCode string) *CustomError {
	err_code, _ := convertStringToInt(customErrCode)
	err := newCustomError(customMsg, err_code)
	return err
}
