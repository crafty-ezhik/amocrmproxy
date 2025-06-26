package utils

import (
	"bytes"
	"io"
	"net/http"
)

// MakeRequest - Делает запрос по указанному url. Добавляет заголовок Authorization, значение которого
// берется из пришедшего запроса. Добавление body происходит в зависимости от его наличии в параметрах.
//
// Возвращает указатель на структуру Response
func MakeRequest(r *http.Request, c *http.Client, method string, url string, body []byte) *Response {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return &Response{
			Body:   nil,
			Status: http.StatusInternalServerError,
			Error:  err,
		}
	}

	if body != nil {
		req.Body = io.NopCloser(bytes.NewReader(body))
	}

	req.Header.Set(HeaderContentType, "application/json")
	if token := r.Header.Get(HeaderAuthorization); token != "" {
		req.Header.Set(HeaderAuthorization, r.Header.Get(HeaderAuthorization))
	}

	resp, err := c.Do(req)
	if err != nil {
		return &Response{
			Body:   nil,
			Status: http.StatusInternalServerError,
			Error:  err,
		}
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return &Response{
			Body:   nil,
			Status: http.StatusInternalServerError,
			Error:  err,
		}
	}
	return &Response{
		Body:   respBody,
		Status: resp.StatusCode,
		Error:  nil,
	}
}

func SendResponse(w http.ResponseWriter, response *Response) {
	w.WriteHeader(response.Status)

	if response.Error != nil {
		w.Write([]byte(response.Error.Error()))
		return
	}

	w.Header().Set(HeaderContentType, "application/json; charset=UTF-8")
	w.Write(response.Body)
}
