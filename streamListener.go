package sam3

import (
	"bufio"
	"errors"
	"io"
	"log"
	"net"
	"strconv"
	"strings"

	"github.com/eyedeekay/i2pkeys"
)

type StreamListener struct {
	// parent stream session
	session *StreamSession
	// our session id
	id string
	// our local address for this sam socket
	laddr i2pkeys.I2PAddr
}

func (l *StreamListener) From() string {
	return l.session.from
}

func (l *StreamListener) To() string {
	return l.session.to
}

// get our address
// implements net.Listener
func (l *StreamListener) Addr() net.Addr {
	return l.laddr
}

// implements net.Listener
func (l *StreamListener) Close() error {
	return l.session.Close()
}

// implements net.Listener
func (l *StreamListener) Accept() (net.Conn, error) {
	return l.AcceptI2P()
}

func ExtractPairString(input, value string) string {
	parts := strings.Split(input, " ")
	for _, part := range parts {
		if strings.HasPrefix(part, value) {
			kv := strings.SplitN(input, "=", 2)
			if len(kv) == 2 {
				return kv[1]
			}
		}
	}
	return ""
}

func ExtractPairInt(input, value string) int {
	rv, err := strconv.Atoi(ExtractPairString(input, value))
	if err != nil {
		return 0
	}
	return rv
}

func ExtractDest(input string) string {
	return strings.Split(input, " ")[0]
}

// accept a new inbound connection
func (l *StreamListener) AcceptI2P() (*SAMConn, error) {
	s, err := NewSAM(l.session.samAddr)
	if err == nil {
		// we connected to sam
		// send accept() command
		_, err = io.WriteString(s.conn, "STREAM ACCEPT ID="+l.id+" SILENT=false\n")
		// read reply
		rd := bufio.NewReader(s.conn)
		// read first line
		line, err := rd.ReadString(10)
		log.Println(line)
		if err == nil {
			if strings.HasPrefix(line, "STREAM STATUS RESULT=OK") {
				// we gud read destination line
				destline, err := rd.ReadString(10)
				log.Println(destline)
				if err == nil {
					dest := ExtractDest(destline)
					l.session.from = ExtractPairString(destline, "FROM_PORT")
					l.session.to = ExtractPairString(destline, "TO_PORT")
					// return wrapped connection
					dest = strings.Trim(dest, "\n")
					return &SAMConn{
						laddr: l.laddr,
						raddr: i2pkeys.I2PAddr(dest),
						conn:  s.conn,
					}, nil
				} else {
					s.Close()
					return nil, err
				}
			} else {
				s.Close()
				return nil, errors.New("invalid sam line: " + line)
			}
		} else {
			s.Close()
			return nil, err
		}
	}
	s.Close()
	return nil, err
}
