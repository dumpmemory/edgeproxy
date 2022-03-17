package stream

import (
	"edgeproxy/metrics"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	"runtime/debug"
)

type copyDirection uint

const (
	readDirection  copyDirection = 0
	writeDirection copyDirection = 1
)

type BidirectionalStream struct {
	doneChan     chan struct{}
	readBytes    *int64
	writtenBytes *int64
	conn1        io.ReadWriter
	conn2        io.ReadWriter
	conn1Name    string
	conn2Name    string
}

func NewBidirectionalStream(conn1, conn2 io.ReadWriter, conn1Name, conn2Name string) *BidirectionalStream {
	return &BidirectionalStream{
		doneChan:     make(chan struct{}, 2),
		readBytes:    new(int64),
		writtenBytes: new(int64),
		conn1:        conn1,
		conn2:        conn2,
		conn1Name:    conn1Name,
		conn2Name:    conn2Name,
	}
}

func (b *BidirectionalStream) Stream() (readBytes int64, writtenBytes int64) {
	go b.copyData(b.conn1, b.conn2, fmt.Sprintf("%s->%s", b.conn1Name, b.conn2Name), readDirection)
	go b.copyData(b.conn2, b.conn1, fmt.Sprintf("%s->%s", b.conn2Name, b.conn1Name), writeDirection)
	b.waitAnyDone()

	log.Debugf("Connection terminated, sent: %d bytes received:%d bytes", *b.readBytes, *b.writtenBytes)
	metrics.IncrementRouterReadBytes(*b.readBytes)
	metrics.IncrementRouterWrittenBytes(*b.writtenBytes)
	return *b.readBytes, *b.writtenBytes
}

func (b *BidirectionalStream) markUniStreamDone() {
	b.doneChan <- struct{}{}
}

func (b *BidirectionalStream) waitAnyDone() {
	<-b.doneChan
}

func (b *BidirectionalStream) copyData(dst, src io.ReadWriter, dir string, direction copyDirection) {
	defer func() {
		if r := recover(); r != nil {
			log.Debugf("Gracefully handled error %v in Streaming for %s, error %s", r, dir, debug.Stack())
		}
	}()
	bytesTransfered, err := io.Copy(dst, src)
	if err != nil {
		log.Debugf("%s copy: %v", dir, err)
	}
	if direction == readDirection {
		*b.readBytes = bytesTransfered
	} else {
		*b.writtenBytes = bytesTransfered
	}
	b.markUniStreamDone()

}
