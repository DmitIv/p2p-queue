package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	queueCore "github.com/dmitIv/playground/go/libp2p-queue/core"
)

var (
	ctx           context.Context
	cancelFunc    context.CancelFunc
	err           error
	signalChannel chan os.Signal
	queueClient   *queueCore.QueueClient
)

type cliArguments struct {
	portFlag *int
}

func parseArgs() (args cliArguments) {
	args = cliArguments{}

	args.portFlag = flag.Int("port", 0, "the port number for a queue client.")
	flag.Parse()

	return args
}

func init() {
	signalChannel = make(chan os.Signal, 1)
	signal.Notify(signalChannel, syscall.SIGINT, syscall.SIGTERM)

	ctx, cancelFunc = context.WithCancel(context.Background())
}

func main() {
	log.Println("Queue started.")
	defer func() {
		log.Println("Queue finished.")
	}()

	args := parseArgs()
	address := fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", *args.portFlag)

	if queueClient, err = queueCore.NewQueueClient(ctx, address); err != nil {
		log.Fatalf("creating queue client failed: %v\n", err)
	}
	defer func() {
		if err := queueClient.Close(); err != nil {
			log.Printf("queue client closing failed: %v\n", err)
		}
	}()
	log.Printf("client queue open at: %v\n", queueClient.HostAddresses)

	go listen(ctx, queueClient)
	<-signalChannel
	cancelFunc()
	time.Sleep(time.Second)
}

func listen(ctx context.Context, qc *queueCore.QueueClient) {
	defer log.Println("finish listening")
	log.Println("start listening")

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	inboundMessages := qc.Read(ctx, 128)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			msg := fmt.Sprintf("Hi! I'm a %v", qc.SelfID)
			log.Printf("send message to the queue: %s\n", msg)
			queueMessage := new(queueCore.QueueMessage)
			queueMessage.Message = msg

			if err := qc.Write(ctx, queueMessage); err != nil {
				log.Printf("writing to queue failed: %v\n", err)
			}
		case msg := <-inboundMessages:
			log.Printf("message from queue: %v\n", msg.Message)
		}
	}
}
