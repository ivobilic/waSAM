// +build nettest

package sam3

import (
	"fmt"
	"log"
	"strings"
	"testing"
	"time"

	"github.com/eyedeekay/sam3/i2pkeys"
)

func Test_PrimaryStreamingDial(t *testing.T) {
	if testing.Short() {
		return
	}
	fmt.Println("Test_PrimaryStreamingDial")
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
	fmt.Println("\tBuilding tunnel")
	ss, err := sam.NewStreamSubSession("primaryStreamTunnel")
	if err != nil {
		fmt.Println(err.Error())
		t.Fail()
		return
	}
	defer ss.Close()
	fmt.Println("\tNotice: This may fail if your I2P node is not well integrated in the I2P network.")
	fmt.Println("\tLooking up i2p-projekt.i2p")
	forumAddr, err := earlysam.Lookup("i2p-projekt.i2p")
	if err != nil {
		fmt.Println(err.Error())
		t.Fail()
		return
	}
	fmt.Println("\tDialing i2p-projekt.i2p(", forumAddr.Base32(), forumAddr.DestHash().Hash(), ")")
	conn, err := ss.DialI2P(forumAddr)
	if err != nil {
		fmt.Println(err.Error())
		t.Fail()
		return
	}
	defer conn.Close()
	fmt.Println("\tSending HTTP GET /")
	if _, err := conn.Write([]byte("GET /\n")); err != nil {
		fmt.Println(err.Error())
		t.Fail()
		return
	}
	buf := make([]byte, 4096)
	n, err := conn.Read(buf)
	if !strings.Contains(strings.ToLower(string(buf[:n])), "http") && !strings.Contains(strings.ToLower(string(buf[:n])), "html") {
		fmt.Printf("\tProbably failed to StreamSession.DialI2P(i2p-projekt.i2p)? It replied %d bytes, but nothing that looked like http/html", n)
	} else {
		fmt.Println("\tRead HTTP/HTML from i2p-projekt.i2p")
	}
}

func Test_PrimaryStreamingServerClient(t *testing.T) {
	if testing.Short() {
		return
	}

	fmt.Println("Test_StreamingServerClient")
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

	sam, err := earlysam.NewPrimarySession("PrimaryServerClientTunnel", keys, []string{"inbound.length=0", "outbound.length=0", "inbound.lengthVariance=0", "outbound.lengthVariance=0", "inbound.quantity=1", "outbound.quantity=1"})
	if err != nil {
		t.Fail()
		return
	}
	defer sam.Close()
	fmt.Println("\tServer: Creating tunnel")
	ss, err := sam.NewUniqueStreamSubSession("primaryExampleServerTun")
	if err != nil {
		return
	}
	defer ss.Close()
	time.Sleep(time.Second * 10)
	c, w := make(chan bool), make(chan bool)
	go func(c, w chan (bool)) {
		if !(<-w) {
			return
		}
		/*sam2, err := NewSAM(yoursam)
		if err != nil {
			c <- false
			return
		}
		defer sam2.Close()
		keys, err := sam2.NewKeys()
		if err != nil {
			c <- false
			return
		}*/
		fmt.Println("\tClient: Creating tunnel")
		ss2, err := sam.NewStreamSubSession("primaryExampleClientTun")
		if err != nil {
			c <- false
			return
		}
		defer ss2.Close()
		fmt.Println("\tClient: Connecting to server")
		conn, err := ss2.DialI2P(ss.Addr())
		if err != nil {
			c <- false
			return
		}
		fmt.Println("\tClient: Connected to tunnel")
		defer conn.Close()
		_, err = conn.Write([]byte("Hello world <3 <3 <3 <3 <3 <3"))
		if err != nil {
			c <- false
			return
		}
		c <- true
	}(c, w)
	l, err := ss.Listen()
	if err != nil {
		fmt.Println("ss.Listen(): " + err.Error())
		t.Fail()
		w <- false
		return
	}
	defer l.Close()
	w <- true
	fmt.Println("\tServer: Accept()ing on tunnel")
	conn, err := l.Accept()
	if err != nil {
		t.Fail()
		fmt.Println("Failed to Accept(): " + err.Error())
		return
	}
	defer conn.Close()
	buf := make([]byte, 512)
	n, err := conn.Read(buf)
	fmt.Printf("\tClient exited successfully: %t\n", <-c)
	fmt.Println("\tServer: received from Client: " + string(buf[:n]))
}

func ExamplePrimaryStreamSession() {
	// Creates a new StreamingSession, dials to idk.i2p and gets a SAMConn
	// which behaves just like a normal net.Conn.

	const samBridge = "127.0.0.1:7656"

	earlysam, err := NewSAM(yoursam)
	if err != nil {
		log.Fatal(err.Error())
		return
	}
	defer earlysam.Close()
	keys, err := earlysam.NewKeys()
	if err != nil {
		log.Fatal(err.Error())
		return
	}

	sam, err := earlysam.NewPrimarySession("PrimaryServerClientTunnel", keys, []string{"inbound.length=0", "outbound.length=0", "inbound.lengthVariance=0", "outbound.lengthVariance=0", "inbound.quantity=1", "outbound.quantity=1"})
	if err != nil {
		log.Fatal(err.Error())
		return
	}
	defer sam.Close()
	// See the example Option_* variables.
	ss, err := sam.NewStreamSubSession("stream_example")
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	ss.Close()
	someone, err := earlysam.Lookup("idk.i2p")
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	conn, err := ss.DialI2P(someone)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	defer conn.Close()
	fmt.Println("Sending HTTP GET /")
	if _, err := conn.Write([]byte("GET /\n")); err != nil {
		fmt.Println(err.Error())
		return
	}
	buf := make([]byte, 4096)
	n, err := conn.Read(buf)
	if !strings.Contains(strings.ToLower(string(buf[:n])), "http") && !strings.Contains(strings.ToLower(string(buf[:n])), "html") {
		fmt.Printf("Probably failed to StreamSession.DialI2P(idk.i2p)? It replied %d bytes, but nothing that looked like http/html", n)
		log.Printf("Probably failed to StreamSession.DialI2P(idk.i2p)? It replied %d bytes, but nothing that looked like http/html", n)
	} else {
		fmt.Println("Read HTTP/HTML from idk.i2p")
		log.Println("Read HTTP/HTML from idk.i2p")
	}
	return

	// Output:
	//Sending HTTP GET /
	//Read HTTP/HTML from idk.i2p
}

func ExamplePrimaryStreamListener() {
	// One server Accept()ing on a StreamListener, and one client that Dials
	// through I2P to the server. Server writes "Hello world!" through a SAMConn
	// (which implements net.Conn) and the client prints the message.

	const samBridge = "127.0.0.1:7656"

	earlysam, err := NewSAM(yoursam)
	if err != nil {
		log.Fatal(err.Error())
		return
	}
	defer earlysam.Close()
	keys, err := earlysam.NewKeys()
	if err != nil {
		log.Fatal(err.Error())
		return
	}

	sam, err := earlysam.NewPrimarySession("PrimaryListenerTunnel", keys, []string{"inbound.length=0", "outbound.length=0", "inbound.lengthVariance=0", "outbound.lengthVariance=0", "inbound.quantity=1", "outbound.quantity=1"})
	if err != nil {
		log.Fatal(err.Error())
		return
	}
	defer sam.Close()

	quit := make(chan bool)

	// Client connecting to the server
	go func(server i2pkeys.I2PAddr) {
		cs, err := sam.NewStreamSubSession("client_example")
		if err != nil {
			fmt.Println(err.Error())
			quit <- false
			return
		}
		defer cs.Close()
		conn, err := cs.DialI2P(server)
		if err != nil {
			fmt.Println(err.Error())
			quit <- false
			return
		}
		buf := make([]byte, 256)
		n, err := conn.Read(buf)
		if err != nil {
			fmt.Println(err.Error())
			quit <- false
			return
		}
		fmt.Println(string(buf[:n]))
		quit <- true
	}(keys.Addr()) // end of client

	ss, err := sam.NewStreamSubSession("server_example")
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	defer ss.Close()
	l, err := ss.Listen()
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	conn, err := l.Accept()
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	conn.Write([]byte("Hello world!"))

	<-quit // waits for client to die, for example only

	// Output:
	//Hello world!
}
