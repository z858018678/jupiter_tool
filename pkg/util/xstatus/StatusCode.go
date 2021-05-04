package xstatus

import (
	"fmt"
	"sync"

	"github.com/golang/protobuf/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type StatusCodes sync.Map

const (
	UnknownCode = "Unknown Status Code"
)

func NewStatusCodes() *StatusCodes {
	return new(StatusCodes)
}

func (s *StatusCodes) Add(code int, desc string) {
	(*sync.Map)(s).Store(code, desc)
}

func (s *StatusCodes) Desc(code int) string {
	var desc, ok = (*sync.Map)(s).Load(code)
	if !ok {
		return fmt.Sprintf("%s:%d", UnknownCode, code)
	}

	return fmt.Sprintf("%d:%v", code, desc)
}

func (s *StatusCodes) NewError(statusCode int, code codes.Code, details ...proto.Message) error {
	var st, _ = status.New(code, s.Desc(statusCode)).WithDetails(details...)
	return st.Err()
}
