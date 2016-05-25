package download

import (
	"github.com/teq0/godown/settings"
	"github.com/teq0/godown/sitemanager"
	"log"
	"net/http"
	"net/http/cookiejar"
	"os"
	"sync"
	//"time"
)

type Download struct {
	OriginalURL   string
	SourceURL     string
	FileDesc      FileDesc
	ChunkRequests []ChunkRequest
	NextRequest   int
	statusChan    chan ChunkStatus
	reqChan       chan ChunkRequest
	resultChan    chan ChunkResult
	wgStatus      sync.WaitGroup
	wgFileWrite   sync.WaitGroup
}

func (dl *Download) redirectHandler(req *http.Request, via []*http.Request) error {

	if req.URL.String() != dl.OriginalURL {
		dl.SourceURL = req.URL.String()
		log.Printf("Download: redirectng to %s", dl.SourceURL)
		dl.FileDesc.ContentLength = via[0].ContentLength
	}

	return nil
}

func (dl *Download) sendNextRequest() bool {
	if dl.NextRequest < len(dl.ChunkRequests) {
		dl.sendRequest(dl.ChunkRequests[dl.NextRequest])
		dl.NextRequest++
		return true
	}

	return false
}

func (dl *Download) sendRequest(chunkReq ChunkRequest) {
	dl.wgStatus.Add(1)
	dl.reqChan <- chunkReq
}

// TODO - work out the file extension from the content-type

func (dl *Download) Start(srcUrl string, destPath string) {

	dl.OriginalURL = srcUrl
	dl.SourceURL = srcUrl

	cookieJar, _ := cookiejar.New(nil)

	client := &http.Client{CheckRedirect: dl.redirectHandler, Jar: cookieJar}

	// get headers so we can process any redirects

	req, _ := http.NewRequest("HEAD", dl.OriginalURL, nil)

	sitemanager.AddAuthHeaders(req)
	req.Header.Add("Accept", "*/*")

	resp, err := client.Do(req)

	if err != nil {
		log.Fatal("Download: initial request failed", err)
		return
	}

	if dl.FileDesc.ContentLength == 0 {
		dl.FileDesc.ContentLength = resp.ContentLength
	}

	log.Printf("HEAD request, status = %d, total size = %d", resp.StatusCode, dl.FileDesc.ContentLength)

	dl.reqChan = make(chan ChunkRequest, 100)
	dl.statusChan = make(chan ChunkStatus, 100)
	dl.resultChan = make(chan ChunkResult, 100)

	go dl.StatusWatcher()

	if err == nil {

		dl.FileDesc = FileDesc{SourceURL: dl.SourceURL, DestPath: destPath, ContentLength: dl.FileDesc.ContentLength}

		log.Printf("dl.FileDesc.Jar created")

		dl.wgFileWrite.Add(1)
		go dl.FileWriter()

		totalLength := dl.FileDesc.ContentLength

		// TODO - figure out how better to deal with Content-Length 0

		if totalLength < 1 {
			totalLength = 40000
		}

		totalChunks := int((dl.FileDesc.ContentLength / settings.Download.MaxChunkSize) + 1)
		chunkSize := settings.Download.MaxChunkSize
		concurrentChunks := settings.Download.MaxConnections

		dl.ChunkRequests = make([]ChunkRequest, totalChunks)

		if concurrentChunks > totalChunks {
			concurrentChunks = totalChunks
		}

		log.Printf("totalChunks: %d\n\rconcurrentchnks: %d", totalChunks, concurrentChunks)

		var chunkHandlers = make([]ChunkHandler, concurrentChunks)

		for idx := 0; idx < totalChunks; idx++ {
			log.Printf("creating request %d...", idx)
			dl.ChunkRequests[idx] = ChunkRequest{Id: idx, Client: client, SourceURL: dl.SourceURL, Offset: int64(idx) * chunkSize, Length: chunkSize}
		}

		for idx := 0; idx < concurrentChunks; idx++ {
			log.Printf("creating goroutine %d...", idx)
			chunkHandlers[idx] = ChunkHandler{Id: idx}
			go chunkHandlers[idx].Get(dl.reqChan, dl.resultChan, dl.statusChan)

			// prime the chans with two requests per handler
			if dl.sendNextRequest() {
				dl.sendNextRequest()
			}
		}

	} else {
		log.Fatal(err)
	}

	//time.Sleep(14000)

	// wait until we've processed all the requests
	dl.wgStatus.Wait()

	log.Printf("Waitgroup done, closing channels...")

	// close the channels
	close(dl.reqChan)
	close(dl.statusChan)
	close(dl.resultChan)

	log.Printf("Waiting on file writer...")
	dl.wgFileWrite.Wait()
	log.Printf("See ya")
}

func (dl *Download) StatusWatcher() {
	for stat := range dl.statusChan {
		log.Printf("Download.StatusWatcher: status received, chunk id = %d, offset = %d, status = %s, bytes written = %d", stat.ChunkId, stat.Offset, stat.StatusCode, stat.BytesWritten)
		if stat.StatusCode == WRITTEN {

			if dl.sendNextRequest() {
				log.Printf("Download.StatusWatcher: next request queued.")
			} else {
				log.Printf("Download.StatusWatcher: no more requests to send.")
			}
		}

		dl.wgStatus.Done()
	}
}

func (dl *Download) FileWriter() {

	log.Printf("Download.FileWriter: downloading file %s to %s", dl.FileDesc.SourceURL, dl.FileDesc.DestPath)

	file, err := os.OpenFile(dl.FileDesc.DestPath, os.O_CREATE|os.O_RDWR, 0777)

	if err != nil {
		log.Fatal("Download.FileWriter: Error opening file %s", dl.FileDesc.DestPath, err)
		return
	}

	defer func() {
		log.Printf("Download.FileWriter: closing file %s", dl.FileDesc.DestPath)
		file.Close()
		dl.wgFileWrite.Done()
	}()

	for res := range dl.resultChan {
		_, err := file.Seek(res.Offset, 0)
		bytesWritten, err := file.Write(res.Body)

		status := ChunkStatus{ChunkId: res.ChunkId, Offset: res.Offset, StatusCode: WRITTEN, BytesWritten: bytesWritten}

		if err != nil {
			log.Fatal("Error writing file %s", dl.FileDesc.DestPath, err)
			status.StatusCode = ERROR
		} else {
			log.Printf("Download.FileWriter: result written, chunk id = %d, offset = %d, status = %s, bytes written = %d", res.ChunkId, res.Offset, res.Status, bytesWritten)
		}

		dl.statusChan <- status
	}

	log.Printf("Download.FileWriter: channel closed")

}
