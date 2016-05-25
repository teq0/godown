package download

import "net/http"

type ChunkRequest struct {
	Id        int
	Client    *http.Client
	SourceURL string
	Offset    int64
	Length    int64
}
