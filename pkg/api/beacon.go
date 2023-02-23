package api

type ErrorContainer struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}
