package serror

import "encoding/json"

//ErrCode is a type of error
type ErrCode string

const (
	ErrCodeInternalError = "internalError"
	ErrCodeBadRequest    = "badRequest"
	ErrCodeNotFound      = "notFound"
)

//Error business error
type Error struct {
	Code      ErrCode `json:"code"`
	Detail    string  `json:"detail"`
	ErrorList ErrList `json:"errorList"`
}

//ErrList is type for custom marshalling
type ErrList []error

func (el ErrList) MarshalJSON() ([]byte, error) {
	res := make([]string, 0)
	for _, e := range el {
		res = append(res, e.Error())
	}
	return json.Marshal(res)
}

//New error
func New(code ErrCode, detail string, errs ...error) *Error {
	return &Error{Code: code, Detail: detail, ErrorList: errs}
}

//NewMarshalRespErr if we have problem with marshalling response
func NewMarshalRespErr(errs ...error) *Error {
	return &Error{Code: ErrCodeInternalError, Detail: "Can't marshal response", ErrorList: errs}
}

//NewUnmarshalReqErr if we have problem with unmarshalling request
func NewUnmarshalReqErr(errs ...error) *Error {
	return &Error{Code: ErrCodeInternalError, Detail: "Can't unmarshal request", ErrorList: errs}
}

//GetMarshalErrorBytes static bytes with error for error in marshalling
func GetMarshalErrorBytes() []byte {
	return []byte(`{ 
	"code": "` + ErrCodeInternalError + `",
	"detail": "Can't marshal response err'",
	"errorList": []
}`)
}
