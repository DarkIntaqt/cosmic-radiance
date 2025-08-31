package queue

/*
INTERNAL:
Drains all entries from the RingBuffer. Then resets the buffer.
This function has currently no use. I'm not sure if should get one, but it is here
*/
func (rb *RingBuffer) drain() {
	for rb.count > 0 {
		entry := rb.dequeue()
		// Send a failed to close the proxy
		entry.FailedResponse(nil)
	}

	// This is totally unnecessary but makes stuff more consistent
	rb.head = 0
	rb.tail = 0
	rb.count = 0
}
