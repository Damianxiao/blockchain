package main

import (
	"blockchain/core"
	"blockchain/crypto"
	"blockchain/network"
	"bytes"
	"net"
	"time"
)

func main() {
	pri := crypto.GenerateKeyPair()
	server := makeServer(&pri, ":30008", []string{":30009", ":30010"})
	// ! nil represent non validtor
	remoteA := makeServer(nil, ":30009", []string{":30008", ":30010"})
	remoteB := makeServer(nil, ":30010", []string{":30008"})
	go server.Start()
	go remoteA.Start()
	go remoteB.Start()
	go dialTest()
	go func() {
		time.Sleep(11 * time.Second)
		lateNode := makeServer(nil, ":6000", []string{":30008"})
		go lateNode.Start()
	}()

	// dialTest()
	select {}
}

func makeServer(pri *crypto.PrivateKey, addr string, seeds []string) *network.Server {
	opts := network.ServerOpts{
		ListenAddress: addr,
		PrivateKey:    pri,
		NodeSeeds:     seeds,
	}
	s := network.NewServer(opts)

	return s
}

func dialTest() error {
	time.Sleep(1 * time.Second)
	conn, err := net.Dial("tcp", "localhost:30008")
	// defer conn.Close()

	if err != nil {
		return err
	}
	pri := crypto.GenerateKeyPair()
	contract := []byte{0x01, 0x0a, 0x46, 0x0c, 0x4f, 0x0c, 0x4f, 0x0c, 0x03, 0x0a, 0x0d, 0x0f}

	tx := core.NewTransaction(contract)
	tx.Sign(&pri)
	buf := &bytes.Buffer{}
	// use proto
	if err := core.NewTxEncoder(buf).Encode(tx); err != nil {
		return err
	}
	msg := network.NewMessage(network.MessageTx, buf.Bytes())

	if err != nil {
		return err
	}
	conn.Write([]byte(msg.Bytes()))
	return nil
}
