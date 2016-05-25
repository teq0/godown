package download

import (
	"net/http"
	//"bytes"
	"log"
	//"io"
	"io/ioutil"
	//"database/sql"
	"fmt"
	//"sync"
)

type ChunkHandler struct {
	Id             int
	CurrentRequest ChunkRequest
}

func (chunk *ChunkHandler) RangeHeader() string {

	hdr := fmt.Sprintf("bytes=%d-%d", chunk.CurrentRequest.Offset, chunk.CurrentRequest.Offset+chunk.CurrentRequest.Length-1)
	log.Printf("Chunk %d: Range = %s", chunk.Id, hdr)
	return hdr
}

func (chunk *ChunkHandler) RedirectHandler(req *http.Request, via []*http.Request) error {

	req.Header.Add("Accept", "*/*")
	req.Header.Add("Range", chunk.RangeHeader())

	return nil
}

func (chunk *ChunkHandler) Get(reqChan chan ChunkRequest, resultChan chan ChunkResult, statusChan chan ChunkStatus) {

	log.Printf("Chunk %d: starting up", chunk.Id)

	for chunkReq := range reqChan {

		chunk.CurrentRequest = chunkReq
		log.Printf("Chunk %d: received request, id = %d, offset = %d", chunk.Id, chunkReq.Id, chunk.CurrentRequest.Offset)

		// just in case we need to exit without swnding a ChunkResult
		status := ChunkStatus{ChunkId: chunk.CurrentRequest.Id, Offset: chunk.CurrentRequest.Offset, StatusCode: ERROR}

		req, err := http.NewRequest("GET", chunkReq.SourceURL, nil)

		//log.Printf("Chunk %d: request created", chunk.Id)

		if err != nil {
			log.Fatal("Chunk %d: ", chunk.Id, err)
			statusChan <- status
			return
		}

		// req.SetBasicAuth("spinmoment", "n0odl3m4n")
		req.Header.Add("Accept", "*/*")
		req.Header.Add("Range", chunk.RangeHeader())
		resp, err := chunkReq.Client.Do(req)

		if err == nil {

			//log.Printf("Chunk %d, req %d: reading body...", chunk.Id, chunkReq.Id)

			body, err1 := ioutil.ReadAll(resp.Body)

			//log.Printf("Chunk %d, req %d: body read", chunk.Id, chunkReq.Id)

			defer resp.Body.Close()
			if err1 == nil {
				log.Printf("Chunk %d, req %d: status = %d, body length = %d", chunk.Id, chunkReq.Id, resp.StatusCode, len(body))

			} else {
				log.Fatal(err1)
				statusChan <- status
				return
			}

			result := ChunkResult{ChunkId: chunk.CurrentRequest.Id, Offset: chunk.CurrentRequest.Offset, Body: body, Status: "Received"}

			resultChan <- result

		} else {
			log.Fatal(err)
			statusChan <- status
		}
	}
}
