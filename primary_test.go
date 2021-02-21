// +build nettest

package sam3

import (
	"fmt"
	//	"log"
	"testing"
	"time"
)

func Test_PrimaryServerClient(t *testing.T) {
	if testing.Short() {
		return
	}

	fmt.Println("Test_DatagramServerClient")
	earlysam, err := NewSAM(yoursam)
	if err != nil {
		t.Fail()
		return
	}
	defer earlysam.Close()
	keys, err := earlysam.NewKeys()
	if err != nil {
		t.Fail()
		return
	}

	sam, err := earlysam.NewPrimarySession("PrimaryTunnel", keys, []string{"inbound.length=0", "outbound.length=0", "inbound.lengthVariance=0", "outbound.lengthVariance=0", "inbound.quantity=1", "outbound.quantity=1"})
	if err != nil {
		t.Fail()
		return
	}
	defer sam.Close()
	//	fmt.Println("\tServer: My address: " + keys.Addr().Base32())
	fmt.Println("\tServer: Creating tunnel")
	ds, err := sam.NewDatagramSubSession("PrimaryTunnel-3", 0)
	if err != nil {
		fmt.Println("Server: Failed to create tunnel: " + err.Error())
		t.Fail()
		return
	}
	defer ds.Close()
	c, w := make(chan bool), make(chan bool)
	go func(c, w chan (bool)) {
		sam2, err := NewSAM(yoursam)
		if err != nil {
			c <- false
			return
		}
		defer sam2.Close()
		keys, err := sam2.NewKeys()
		if err != nil {
			c <- false
			return
		}
		fmt.Println("\tClient: Creating tunnel")
		ds2, err := sam2.NewDatagramSession("PRIMARYClientTunnel", keys, []string{"inbound.length=0", "outbound.length=0", "inbound.lengthVariance=0", "outbound.lengthVariance=0", "inbound.quantity=1", "outbound.quantity=1"}, 0)
		if err != nil {
			c <- false
			return
		}
		defer ds2.Close()
		//		fmt.Println("\tClient: Servers address: " + ds.LocalAddr().Base32())
		//		fmt.Println("\tClient: Clients address: " + ds2.LocalAddr().Base32())
		fmt.Println("\tClient: Tries to send primary to server")
		for {
			select {
			default:
				_, err = ds2.WriteTo([]byte("Hello primary-world! <3 <3 <3 <3 <3 <3"), ds.LocalAddr())
				if err != nil {
					fmt.Println("\tClient: Failed to send primary: " + err.Error())
					c <- false
					return
				}
				time.Sleep(5 * time.Second)
			case <-w:
				fmt.Println("\tClient: Sent primary, quitting.")
				return
			}
		}
		c <- true
	}(c, w)
	buf := make([]byte, 512)
	fmt.Println("\tServer: ReadFrom() waiting...")
	n, _, err := ds.ReadFrom(buf)
	w <- true
	if err != nil {
		fmt.Println("\tServer: Failed to ReadFrom(): " + err.Error())
		t.Fail()
		return
	}
	fmt.Println("\tServer: Received primary: " + string(buf[:n]))
	//	fmt.Println("\tServer: Senders address was: " + saddr.Base32())
}

/*
func ExamplePrimarySession() {
	// Creates a new DatagramSession, which behaves just like a net.PacketConn.

	const samBridge = "127.0.0.1:7656"

	sam, err := NewSAM(samBridge)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	keys, err := sam.NewKeys()
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	myself := keys.Addr()

	// See the example Option_* variables.
	dg, err := sam.NewDatagramSession("DGTUN", keys, Options_Small, 0)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	someone, err := sam.Lookup("zzz.i2p")
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	dg.WriteTo([]byte("Hello stranger!"), someone)
	dg.WriteTo([]byte("Hello myself!"), myself)

	buf := make([]byte, 31*1024)
	n, _, err := dg.ReadFrom(buf)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	log.Println("Got message: '" + string(buf[:n]) + "'")
	fmt.Println("Got message: " + string(buf[:n]))

	return
	// Output:
	//Got message: Hello myself!
}

func ExampleMiniPrimarySession() {
	// Creates a new DatagramSession, which behaves just like a net.PacketConn.

	const samBridge = "127.0.0.1:7656"

	sam, err := NewSAM(samBridge)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	keys, err := sam.NewKeys()
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	myself := keys.Addr()

	// See the example Option_* variables.
	dg, err := sam.NewDatagramSession("MINIDGTUN", keys, Options_Small, 0)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	someone, err := sam.Lookup("zzz.i2p")
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	err = dg.SetWriteBuffer(14 * 1024)
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	dg.WriteTo([]byte("Hello stranger!"), someone)
	dg.WriteTo([]byte("Hello myself!"), myself)

	buf := make([]byte, 31*1024)
	n, _, err := dg.ReadFrom(buf)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	log.Println("Got message: '" + string(buf[:n]) + "'")
	fmt.Println("Got message: " + string(buf[:n]))

	return
	// Output:
	//Got message: Hello myself!
}*/
