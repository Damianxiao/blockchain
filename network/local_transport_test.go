package network

// func TestBroadcast(t *testing.T) {
// 	local := NewLocalTransport("A")
// 	remoteB := NewLocalTransport("B")
// 	remoteC := NewLocalTransport("C")

// 	local.Connect(remoteB)
// 	local.Connect(remoteC)
// 	msg := []byte("fopop")
// 	err := local.Broadcast(msg)
// 	for rpc := range remoteB.Consume() {
// 		assert.Nil(t, err)
// 		assert.Equal(t, msg, rpc.Payload)
// 	}
// 	for rpc := range remoteC.Consume() {
// 		assert.Nil(t, err)
// 		assert.Equal(t, msg, rpc.Payload)
// 	}

// }
