package dto

import "time"

type MetadataResponse struct {
	RequestId string `json:"request_id"`
	Timestamp string `json:"timestamp"`
}

type ErrorResponse struct {
	Code    string               `json:"code"`
	Message string               `json:"message"`
	Fields  *[]map[string]string `json:"fields,omitempty"`
}

type APIResponse[T any] struct {
	Data       T                 `json:"data"`
	Meta       *MetadataResponse `json:"meta"`
	Pagination *PaginationMeta   `json:"pagination,omitempty"`
	Error      *ErrorResponse    `json:"error,omitempty"`
}

func NewAPIResponse[T any](data T, err *ErrorResponse) *APIResponse[T] {
	return &APIResponse[T]{
		Data: data,
		Meta: &MetadataResponse{
			RequestId: "",
			Timestamp: time.Now().Format(time.RFC3339),
		},
		Error: err,
	}
}

func NewAPIResponseWithPagination[T any](data T, pagination *PaginationMeta) *APIResponse[T] {
	return &APIResponse[T]{
		Data: data,
		Meta: &MetadataResponse{
			RequestId: "",
			Timestamp: time.Now().Format(time.RFC3339),
		},
		Pagination: pagination,
		Error:      nil,
	}
}
