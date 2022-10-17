package common

import (
	"fmt"
)

type ErrorCode struct {
	code ReturnCode
	err  error
}

func (e *ErrorCode) Code() ReturnCode {
	return e.code
}

func (e *ErrorCode) Error() string {
	return fmt.Sprintf("%s:0x%x", ReturnCode_name[int32(e.code)], e.code.Number())
}

func NewErrorCode(code ReturnCode, err error) *ErrorCode {
	return &ErrorCode{code: code, err: err}
}

func (e *ErrorCode) Unwap() error {
	return e.err
}
