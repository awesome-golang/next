package client

import (
	"github.com/chzyer/flow"
	"github.com/chzyer/logex"
	"github.com/chzyer/next/ip"
	"github.com/chzyer/next/uc"
	"github.com/chzyer/tunnel"
)

type Tun struct {
	tun  *tunnel.Instance
	flow *flow.Flow
}

func newTun(f *flow.Flow, remoteCfg *uc.AuthResponse, cfg *Config) (*Tun, error) {
	ipnet, err := ip.ParseCIDR(remoteCfg.Gateway)
	if err != nil {
		return nil, logex.Trace(err)
	}
	ipnet.IP = ip.ParseIP(remoteCfg.INet)

	tun, err := tunnel.New(&tunnel.Config{
		DevId:   cfg.DevId,
		Gateway: ipnet.IP.IP(),
		Mask:    ipnet.Mask,
		MTU:     remoteCfg.MTU,
		Debug:   cfg.Debug,
	})
	if err != nil {
		return nil, logex.Trace(err)
	}
	t := &Tun{
		tun: tun,
	}
	f.ForkTo(&t.flow, t.Close)

	return t, nil
}

// ConfigUpdate: non-test
func (t *Tun) ConfigUpdate(remoteCfg *uc.AuthResponse) {

}

func (t *Tun) Close() {
	t.tun.Close()
	t.flow.Close()
}

func (t *Tun) Run(in, out chan []byte) {
	go t.readLoop(out)
	go t.writeLoop(in)
}

func (t *Tun) writeLoop(in chan []byte) {
	t.flow.Add(1)
	defer t.flow.DoneAndClose()
loop:
	for {
		select {
		case data := <-in:
			n, err := t.tun.Write(data)
			logex.Debug("tun write:", n)
			if err != nil {
				break loop
			}
		case <-t.flow.IsClose():
			break loop
		}
	}
}

func (t *Tun) readLoop(out chan []byte) {
	buf := make([]byte, 65536)

loop:
	for {
		n, err := t.tun.Read(buf)
		logex.Debug("tun read:", n)
		if err != nil {
			break
		}
		b := make([]byte, n)
		copy(b, buf[:n])
		select {
		case out <- b:
		case <-t.flow.IsClose():
			break loop
		}
	}
}

func (t *Tun) Name() string {
	return t.tun.Name
}
