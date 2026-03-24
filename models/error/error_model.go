package errors

type ErrorModel struct {
	Message string `json:"message"`
	IsSuccess bool `json:"isSuccess"`
	Error error `json:"error"`
}