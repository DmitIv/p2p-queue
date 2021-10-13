package core

import (
	"context"
	"encoding/json"

	"github.com/libp2p/go-libp2p-core/host"
	peerstore "github.com/libp2p/go-libp2p-core/peer"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
)

const (
	queueTopicName = "p2p-pubsub-queue"
)

type QueueMessage struct {
	Message string
}

type QueueGateway struct {
	selfID       peerstore.ID
	pubSub       *pubsub.PubSub
	topic        *pubsub.Topic
	subscription *pubsub.Subscription
}

type fromItselfError struct {
	Message *pubsub.Message
}

func (err *fromItselfError) Error() string {
	return "get message from itself."
}

func NewGateway(ctx context.Context, h host.Host) (qg *QueueGateway, err error) {
	qg = new(QueueGateway)
	if qg.pubSub, err = pubsub.NewGossipSub(ctx, h); err != nil {
		return nil, err
	}

	if qg.topic, err = qg.pubSub.Join(queueTopicName); err != nil {
		return nil, err
	}

	if qg.subscription, err = qg.topic.Subscribe(); err != nil {
		return nil, err
	}
	qg.selfID = h.ID()

	return qg, nil
}

func (qg *QueueGateway) nextMessage(ctx context.Context) (msg *pubsub.Message, err error) {
	if msg, err = qg.subscription.Next(ctx); err != nil {
		return nil, err
	}

	if msg.ReceivedFrom == qg.selfID {
		return nil, &fromItselfError{msg}
	}

	return msg, nil
}

func (qg *QueueGateway) sendMessage(ctx context.Context, msgBytes []byte) (err error) {
	if err = qg.topic.Publish(ctx, msgBytes); err != nil {
		return err
	}

	return nil
}

func (qg *QueueGateway) NextMessage(ctx context.Context) (qmsg *QueueMessage, err error) {
	var msg *pubsub.Message
	if msg, err = qg.nextMessage(ctx); err != nil {
		return nil, err
	}

	qmsg = new(QueueMessage)
	if err = json.Unmarshal(msg.Data, qmsg); err != nil {
		return nil, err
	}

	return qmsg, nil
}

func (qg *QueueGateway) SendMessage(ctx context.Context, qmsg *QueueMessage) (err error) {
	var qmsgBytes []byte
	if qmsgBytes, err = json.Marshal(qmsg); err != nil {
		return err
	}

	return qg.sendMessage(ctx, qmsgBytes)
}

func (qg *QueueGateway) Close() (err error) {
	qg.subscription.Cancel()
	return nil
}
