package controller

import (
	"sync/atomic"
	"testing"

	"github.com/chzyer/flow"
	"github.com/chzyer/next/packet"
	"github.com/chzyer/test"
)

type dumpReqider struct {
	reqid uint32
}

func (d *dumpReqider) GetReqId() uint32 {
	return atomic.AddUint32(&d.reqid, 1)
}

func TestStage(t *testing.T) {
	defer test.New(t)

	dr := &dumpReqider{}
	s := newStage(flow.New())
	p := packet.New(nil, packet.HEARTBEAT)
	p.InitIV(dr)
	req := NewRequest(p, true)

	{
		s.Add(req)
		req2 := s.Remove(p.IV.ReqId)
		test.Equal(req, req2)

		req2 = s.Remove(p.IV.ReqId)
		test.Nil(req2)

		test.Equal(len(s.ShowStage()), 0)
	}
}
