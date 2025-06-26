package utils

const (
	HeaderContentType   = "Content-Type"
	HeaderAuthorization = "Authorization"
)

type AnyMap map[string]interface{}

// Response - Структура для отдачи ответа клиенту
type Response struct {
	Body   []byte
	Status int
	Error  error
}
