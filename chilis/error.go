package chilis

// BadRequestError is analagous to an HTTP 400 response.
type BadRequestError struct {
	Field string
	Value string
}

func (bre BadRequestError) Error() string {
	return "invalid " + bre.Field
}

// ForbiddenError is analagous to an HTTP 403 response.
type ForbiddenError struct {
	Reason string
}

func (fe ForbiddenError) Error() string {
	return fe.Reason
}
