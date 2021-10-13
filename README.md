# p2p-queue

## About

A test bench with a queue based on a p2p network. Under the hood uses [the libp2p](https://github.com/libp2p/go-libp2p) package. In the current implementation, the nodes participating in the network exchange short messages with "Hi, I'm \<peer id\>" within the sole topic. There is no authorization to participate in the topic, no [NAT punching](https://docs.libp2p.io/concepts/nat/), and no log contains messages history.
  
 ## Usage
  
 ```shell
 git clone git@github.com:DmitIv/p2p-queue.git
 cd p2p-queue
 go run .
 ```
