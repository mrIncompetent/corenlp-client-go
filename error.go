package corenlp

import "fmt"

type ServerError struct {
	statusCode int
	body       string
}

func (e *ServerError) Error() string {
	return fmt.Sprintf(`The server failed to process the request.
Returned HTTP status code: %d
Response body:
%s`, e.statusCode, e.body)
}
