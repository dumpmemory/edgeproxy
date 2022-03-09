package transport

import (
	log "github.com/sirupsen/logrus"
	"io"
	"runtime/debug"
	"sync/atomic"
)

type bidirectionalStreamStatus struct {
	doneChan chan struct{}
	anyDone  uint32
}

func newBiStreamStatus() *bidirectionalStreamStatus {
	return &bidirectionalStreamStatus{
		doneChan: make(chan struct{}, 2),
		anyDone:  0,
	}
}

func Stream(tunnelConn, originConn io.ReadWriter) {
	status := newBiStreamStatus()
	go copyData(tunnelConn, originConn, "origin->tunnel", status)
	go copyData(originConn, tunnelConn, "tunnel->origin", status)
	// If one side is done, we are done.
	status.waitAnyDone()
	log.Debugf("Connection terminated")
}

func copyData(dst, src io.ReadWriter, dir string, status *bidirectionalStreamStatus) {
	defer func() {
		if r := recover(); r != nil {
			//TODO we must control any panic in this goroutines to prevent leacked connections
			if status.isAnyDone() {
				// We handle such unexpected errors only when we detect that one side of the streaming is done.
				log.Debugf("Gracefully handled error %v in Streaming for %s, error %s", r, dir, debug.Stack())
			} else {
				// Otherwise, this is unexpected, but we prevent the program from crashing anyway.
				log.Warnf("Gracefully handled unexpected error %v in Streaming for %s, error %s", r, dir, debug.Stack())
			}
		}
	}()
	if _, err := io.Copy(dst, src); err != nil {
		log.Debugf("%s copy: %v", dir, err)
	}
	status.markUniStreamDone()

}

func (s *bidirectionalStreamStatus) markUniStreamDone() {
	atomic.StoreUint32(&s.anyDone, 1)
	s.doneChan <- struct{}{}
}

func (s *bidirectionalStreamStatus) waitAnyDone() {
	<-s.doneChan
}
func (s *bidirectionalStreamStatus) isAnyDone() bool {
	return atomic.LoadUint32(&s.anyDone) > 0
}
