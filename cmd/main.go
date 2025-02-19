package main

import (
	"blockchain/core"
	"blockchain/crypto"
	"blockchain/network"
	"bytes"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"
)

func main() {
	local := network.NewLocalTransport("local")
	remoteA := network.NewLocalTransport("remoteA")
	remoteB := network.NewLocalTransport("remoteB")
	remoteC := network.NewLocalTransport("remoteC")

	local.Connect(remoteA)
	remoteA.Connect(remoteB)
	remoteB.Connect(remoteC)
	remoteC.Connect(local)
	initRemoteServer([]network.Transport{
		remoteA, remoteB, remoteC,
	})
	go func() {
		for {
			if err := sendTransaction(local, remoteA.GetAddr()); err != nil {
				logrus.Error(err)
			}
			time.Sleep(5 * time.Second)
		}
	}()

	pri := crypto.GenerateKeyPair()
	localServer := makeServer("LOCAL", local, &pri)
	localServer.Start()

}

func initRemoteServer(trs []network.Transport) {
	for index, _ := range trs {
		id := fmt.Sprintf("REMOTE %v", index)
		s := makeServer(id, trs[index], nil)
		go s.Start()
	}
}

func makeServer(id string, tr network.Transport, pri *crypto.PrivateKey) *network.Server {
	opts := network.ServerOpts{
		Id:         id,
		Transports: []network.Transport{tr},
		PrivateKey: pri,
	}
	s := network.NewServer(opts)
	return s
}

func sendTransaction(tr network.Transport, to network.NetAddr) error {
	pri := crypto.GenerateKeyPair()
	d := []byte(strconv.Itoa(rand.Intn(10000000)))
	tx := core.NewTransaction(d)
	tx.Sign(&pri)
	buf := &bytes.Buffer{}
	// use proto
	if err := core.NewTxEncoder(buf).Encode(tx); err != nil {
		return err
	}
	msg := network.NewMessage(network.MessageTx, buf.Bytes())

	err := tr.SendMessage(to, msg.Bytes())
	if err != nil {
		return err
	}
	return nil
}
