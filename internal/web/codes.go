package web

const (
	CodeInvalidEmail = 1
)

type ValidationError struct {
	StatusCode int
	Errors     []string
}

type RequestError struct {
	StatusCode int
	Message    string
}

func (r *ValidationError) Error() string {
	if len(r.Errors) == 0 {
		return "validation error"
	}

	return r.Errors[0]
}

type AuthorizationError struct {
	StatusCode int
	Errors     []string
}

func (r *AuthorizationError) Error() string {
	if len(r.Errors) == 0 {
		return "authorization error"
	}

	return r.Errors[0]
}
