package webapi

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/pzabolotniy/logging/pkg/logging"
)

type ResponseBody struct {
	Data  any    `json:"data,omitempty"`
	Error string `json:"error,omitempty"`
}

type Response struct {
	HTTPBody   *ResponseBody
	HTTPStatus int
}

func InternalServerError(ctx context.Context, w http.ResponseWriter, msg string) {
	respBody := &ResponseBody{
		Error: msg,
	}
	resp := &Response{
		HTTPStatus: http.StatusInternalServerError,
		HTTPBody:   respBody,
	}
	makeJSONResponse(ctx, w, resp)
}

func CreatedResponse(ctx context.Context, w http.ResponseWriter, data any) {
	respBody := &ResponseBody{
		Data: data,
	}
	resp := &Response{
		HTTPStatus: http.StatusCreated,
		HTTPBody:   respBody,
	}
	makeJSONResponse(ctx, w, resp)
}

func BadRequest(ctx context.Context, w http.ResponseWriter, msg string) {
	respBody := &ResponseBody{
		Error: msg,
	}
	resp := &Response{
		HTTPStatus: http.StatusBadRequest,
		HTTPBody:   respBody,
	}
	makeJSONResponse(ctx, w, resp)
}

func Forbidden(ctx context.Context, w http.ResponseWriter, msg string) {
	respBody := &ResponseBody{
		Error: msg,
	}
	resp := &Response{
		HTTPStatus: http.StatusForbidden,
		HTTPBody:   respBody,
	}
	makeJSONResponse(ctx, w, resp)
}

func OKResponse(ctx context.Context, w http.ResponseWriter, data any) {
	respBody := &ResponseBody{
		Data: data,
	}
	resp := &Response{
		HTTPStatus: http.StatusOK,
		HTTPBody:   respBody,
	}
	makeJSONResponse(ctx, w, resp)
}

func makeJSONResponse(ctx context.Context, w http.ResponseWriter, resp *Response) {
	logger := logging.FromContext(ctx)
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(resp.HTTPStatus)
	if encodeErr := json.NewEncoder(w).Encode(resp.HTTPBody); encodeErr != nil {
		logger.WithError(encodeErr).Error("encode response failed")
		w.WriteHeader(http.StatusInternalServerError)
	}
}
