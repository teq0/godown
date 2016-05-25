package download

type ChunkStatusCode int

// TODO - actually don't need most of these

const (
	CREATED  = iota
	SENT     = iota
	RECEIVED = iota
	WRITTEN  = iota
	ERROR    = iota
)

func (csc ChunkStatusCode) String() string {
	var statii = [...]string{
		"CREATED",
		"SENT",
		"RECEIVED",
		"WRITTEN",
		"ERROR",
	}

	return statii[csc]
}

type ChunkStatus struct {
	ChunkId      int
	Offset       int64
	BytesWritten int
	StatusCode   ChunkStatusCode
}
