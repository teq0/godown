package download

type ChunkResult struct {
	ChunkId int
	Offset  int64
	Percent int
	Status  string
	Body    []byte
}
