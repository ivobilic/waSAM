package sam3

import (
	//	"bufio"
	//	"bytes"
	//	"context"
	//	"errors"
	//	"io"
	//	"log"
	"net"
	//	"strconv"
	//	"strings"
	"time"

	"github.com/eyedeekay/sam3/i2pkeys"
)

// Represents a primary session.
type PrimarySession struct {
	samAddr  string          // address to the sam bridge (ipv4:port)
	id       string          // tunnel name
	conn     net.Conn        // connection to sam
	keys     i2pkeys.I2PKeys // i2p destination keys
	Timeout  time.Duration
	Deadline time.Time
	sigType  string
	//	from     string
	//	to       string
}

func (ss *PrimarySession) From() string {
	return "0"
}

func (ss *PrimarySession) To() string {
	return "0"
}

func (ss *PrimarySession) SignatureType() string {
	return ss.sigType
}

// Returns the local tunnel name of the I2P tunnel used for the stream session
func (ss *PrimarySession) ID() string {
	return ss.id
}

func (ss *PrimarySession) Close() error {
	return ss.conn.Close()
}

// Returns the I2P destination (the address) of the stream session
func (ss *PrimarySession) Addr() i2pkeys.I2PAddr {
	return ss.keys.Addr()
}

// Returns the keys associated with the stream session
func (ss *PrimarySession) Keys() i2pkeys.I2PKeys {
	return ss.keys
}

// Creates a new PrimarySession with the I2CP- and streaminglib options as
// specified. See the I2P documentation for a full list of options.
func (sam *SAM) NewPrimarySession(id string, keys i2pkeys.I2PKeys, options []string) (*PrimarySession, error) {
	conn, err := sam.newGenericSession("PRIMARY", id, keys, options, []string{})
	if err != nil {
		return nil, err
	}
	return &PrimarySession{sam.Config.I2PConfig.Sam(), id, conn, keys, time.Duration(600 * time.Second), time.Now(), Sig_NONE, "0", "0"}, nil
}

// Creates a new PrimarySession with the I2CP- and PRIMARYinglib options as
// specified. See the I2P documentation for a full list of options.
func (sam *SAM) NewPrimarySessionWithSignature(id string, keys i2pkeys.I2PKeys, options []string, sigType string) (*PrimarySession, error) {
	conn, err := sam.newGenericSessionWithSignature("PRIMARY", id, keys, sigType, options, []string{})
	if err != nil {
		return nil, err
	}
	return &PrimarySession{sam.Config.I2PConfig.Sam(), id, conn, keys, time.Duration(600 * time.Second), time.Now(), sigType, "0", "0"}, nil
}
