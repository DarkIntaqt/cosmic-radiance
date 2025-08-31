package queue

// Dynamically manages the size of the queue
func (rb *RingBuffer) Size() int64 {
	return rb.size
}

func (rb *RingBuffer) GetPeakCapacity() int64 {
	peakCapacity := int64(0)
	for _, group := range *rb.Limits {
		peakCapacity += group.PeakCapacity
	}

	return peakCapacity
}
