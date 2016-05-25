package settings

type downloadSettings struct {
	MaxConnections int
	MaxChunkSize   int64
}

var Download = downloadSettings{10, 1024}
