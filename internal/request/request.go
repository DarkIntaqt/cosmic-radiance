package request

import "time"

type Priority bool

const (
	NormalPriority Priority = false
	HighPriority   Priority = true
	RequestFailed           = -1
)

type Request struct {
	Expire      int64 // Expiration timestamp in milliseconds
	Response    chan *ResponseChannel
	Invalidated bool
}

type ResponseChannel struct {
	KeyId      int
	Update     bool
	RetryAfter *time.Time // Optional
}

// NewRequest creates a new request with an expiration time
func NewRequest(expire time.Duration) *Request {
	return &Request{
		Expire:      time.Now().Add(expire).UnixMilli(),
		Response:    make(chan *ResponseChannel, 1), // A buffer size of 1 to avoid blocking
		Invalidated: false,
	}
}

func (r *Request) Invalidate() {
	r.Invalidated = true
}

// FailedResponse sends a failed response to the requests response channel.
func (r *Request) FailedResponse(time *time.Time) {
	r.Response <- &ResponseChannel{
		KeyId:      RequestFailed,
		Update:     false,
		RetryAfter: time,
	}
}
