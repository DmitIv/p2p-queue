package core

import (
	"context"
	"log"
	"time"

	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	mdns "github.com/libp2p/go-libp2p/p2p/discovery/mdns_legacy"
)

const (
	discoveryInterval   = 5 * time.Minute
	discoveryServiceTag = "p2p-pubsub-queue-imhere"
)

type QueueMDNS struct {
	mdns     mdns.Service
	peerChan chan peer.AddrInfo
}

func NewMDNS(ctx context.Context, h host.Host) (qmdns *QueueMDNS, err error) {
	qmdns = new(QueueMDNS)
	if qmdns.mdns, err = mdns.NewMdnsService(
		ctx, h,
		discoveryInterval, discoveryServiceTag,
	); err != nil {
		return nil, err
	}
	qmdns.peerChan = make(chan peer.AddrInfo, 1)
	qmdns.mdns.RegisterNotifee(qmdns)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case newPeer := <-qmdns.peerChan:
				if err := h.Connect(ctx, newPeer); err != nil {
					log.Printf("connection to %v failed\n", newPeer)
				}
			}
		}
	}()

	return qmdns, nil
}

func (qmdns *QueueMDNS) HandlePeerFound(newPeer peer.AddrInfo) {
	qmdns.peerChan <- newPeer
}

func (qmdns *QueueMDNS) Close() (err error) {
	if err = qmdns.mdns.Close(); err != nil {
		return err
	}

	return nil
}
