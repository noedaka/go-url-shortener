package model

type Request struct {
	URL string `json:"url"`
}

type Response struct {
	Result string `json:"result"`
}

type BatchRequest struct {
	CorrelationID string `json:"correlation_id"`
	URL           string `json:"original_url"`
}

type BatchResponse struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

type UniqueViolationError struct {
	ShortID string
	Err     error
}

func (e *UniqueViolationError) Error() string {
	return "unique violation"
}

func (e *UniqueViolationError) Unwrap() error {
	return e.Err
}

func NewUniqueViolationError(shortID string, err error) *UniqueViolationError {
	return &UniqueViolationError{
		ShortID: shortID,
		Err:     err,
	}
}
