package core

import (
	"context"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/host"
	peerstore "github.com/libp2p/go-libp2p-core/peer"
	"github.com/multiformats/go-multiaddr"
)

type QueueClient struct {
	HostAddresses []multiaddr.Multiaddr
	SelfID        peerstore.ID

	h            host.Host
	qmdns        *QueueMDNS
	queueGateway *QueueGateway
}

func NewQueueClient(ctx context.Context, addresses ...string) (qc *QueueClient, err error) {
	qc = new(QueueClient)
	if qc.h, err = libp2p.New(
		ctx,
		libp2p.ListenAddrStrings(addresses...),
	); err != nil {
		return nil, err
	}

	if qc.qmdns, err = NewMDNS(ctx, qc.h); err != nil {
		return nil, err
	}

	if qc.queueGateway, err = NewGateway(ctx, qc.h); err != nil {
		return nil, err
	}

	qc.SelfID = qc.queueGateway.selfID
	peerInfo := peerstore.AddrInfo{
		ID:    qc.SelfID,
		Addrs: qc.h.Addrs(),
	}
	if qc.HostAddresses, err = peerstore.AddrInfoToP2pAddrs(&peerInfo); err != nil {
		return nil, err
	}

	return qc, err
}

func (qc *QueueClient) Close() (err error) {
	if qc.queueGateway.Close(); err != nil {
		return err
	}

	if qc.qmdns.Close(); err != nil {
		return err
	}

	if qc.h.Close(); err != nil {
		return err
	}

	return nil
}

func (qc *QueueClient) Read(ctx context.Context, bufferSize uint) (qm chan *QueueMessage) {
	qm = make(chan *QueueMessage, bufferSize)
	go func() {
		var (
			msg *QueueMessage
			err error
		)
		for {
			if msg, err = qc.queueGateway.NextMessage(ctx); err != nil {
				// TODO: handle the cancel error for the context done case
				continue
			}
			qm <- msg
		}
	}()
	return qm
}

func (qc *QueueClient) Write(ctx context.Context, qm *QueueMessage) (err error) {
	if err = qc.queueGateway.SendMessage(ctx, qm); err != nil {
		return err
	}

	return nil
}
