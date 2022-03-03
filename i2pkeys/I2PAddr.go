package i2pkeys

import (
	"bytes"
	"crypto"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base32"
	"encoding/base64"
	"errors"
	"io"
	"os"
	"strings"

	"github.com/eyedeekay/goSam"
)

var (
	i2pB64enc *base64.Encoding = base64.NewEncoding("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-~")
	i2pB32enc *base32.Encoding = base32.NewEncoding("abcdefghijklmnopqrstuvwxyz234567")
)

// The public and private keys associated with an I2P destination. I2P hides the
// details of exactly what this is, so treat them as blobs, but generally: One
// pair of DSA keys, one pair of ElGamal keys, and sometimes (almost never) also
// a certificate. String() returns you the full content of I2PKeys and Addr()
// returns the public keys.
type I2PKeys struct {
	Address I2PAddr // only the public key
	Both    string  // both public and private keys
}

// Creates I2PKeys from an I2PAddr and a public/private keypair string (as
// generated by String().)
func NewKeys(addr I2PAddr, both string) I2PKeys {
	return I2PKeys{addr, both}
}

// fileExists checks if a file exists and is not a directory before we
// try using it to prevent further errors.
func fileExists(filename string) (bool, error) {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return !info.IsDir(), nil
}

// load keys from non standard format
func LoadKeysIncompat(r io.Reader) (k I2PKeys, err error) {
	var buff bytes.Buffer
	_, err = io.Copy(&buff, r)
	if err == nil {
		parts := strings.Split(buff.String(), "\n")
		k = I2PKeys{I2PAddr(parts[0]), parts[1]}
	}
	return
}

// load keys from non-standard format by specifying a text file.
// If the file does not exist, generate keys, otherwise, fail
// closed.
func LoadKeys(r string) (I2PKeys, error) {
	exists, err := fileExists(r)
	if err != nil {
		return I2PKeys{}, err
	}
	if exists {
		fi, err := os.Open(r)
		if err != nil {
			return I2PKeys{}, err
		}
		defer fi.Close()
		return LoadKeysIncompat(fi)
	}
	return I2PKeys{}, err
}

// store keys in non standard format
func StoreKeysIncompat(k I2PKeys, w io.Writer) (err error) {
	_, err = io.WriteString(w, k.Address.Base64()+"\n"+k.Both)
	return
}

func StoreKeys(k I2PKeys, r string) error {
	fi, err := os.Open(r)
	if err != nil {
		return err
	}
	defer fi.Close()
	return StoreKeysIncompat(k, fi)
}

func (k I2PKeys) Network() string {
	return k.Address.Network()
}

// Returns the public keys of the I2PKeys.
func (k I2PKeys) Addr() I2PAddr {
	return k.Address
}

func (k I2PKeys) Public() crypto.PublicKey {
	return k.Address
}

func (k I2PKeys) Private() []byte {
	src := strings.Split(k.String(), k.Addr().String())[0]
	var dest []byte
	_, err := i2pB64enc.Decode(dest, []byte(src))
	panic(err)
	return dest
}

type SecretKey interface {
	Sign(rand io.Reader, digest []byte, opts crypto.SignerOpts) (signature []byte, err error)
}

func (k I2PKeys) SecretKey() SecretKey {
	var pk ed25519.PrivateKey = k.Private()
	return pk
}

func (k I2PKeys) PrivateKey() crypto.PrivateKey {
	var pk ed25519.PrivateKey = k.Private()
	_, err := pk.Sign(rand.Reader, []byte("nonsense"), crypto.Hash(0))
	if err != nil {
		//TODO: Elgamal, P256, P384, P512, GOST? keys?
	}
	return pk
}

func (k I2PKeys) Ed25519PrivateKey() *ed25519.PrivateKey {
	return k.SecretKey().(*ed25519.PrivateKey)
}

/*func (k I2PKeys) ElgamalPrivateKey() *ed25519.PrivateKey {
	return k.SecretKey().(*ed25519.PrivateKey)
}*/

//func (k I2PKeys) Decrypt(rand io.Reader, msg []byte, opts crypto.DecrypterOpts) (plaintext []byte, err error) {
//return k.SecretKey().(*ed25519.PrivateKey).Decrypt(rand, msg, opts)
//}

func (k I2PKeys) Sign(rand io.Reader, digest []byte, opts crypto.SignerOpts) (signature []byte, err error) {
	return k.SecretKey().(*ed25519.PrivateKey).Sign(rand, digest, opts)
}

// Returns the keys (both public and private), in I2Ps base64 format. Use this
// when you create sessions.
func (k I2PKeys) String() string {
	return k.Both
}

func (k I2PKeys) HostnameEntry(hostname string, opts crypto.SignerOpts) (string, error) {
	sig, err := k.Sign(rand.Reader, []byte(hostname), opts)
	if err != nil {
		return "", err
	}
	return string(sig), nil
}

// I2PAddr represents an I2P destination, almost equivalent to an IP address.
// This is the humongously huge base64 representation of such an address, which
// really is just a pair of public keys and also maybe a certificate. (I2P hides
// the details of exactly what it is. Read the I2P specifications for more info.)
type I2PAddr string

// an i2p destination hash, the .b32.i2p address if you will
type I2PDestHash [32]byte

// create a desthash from a string b32.i2p address
func DestHashFromString(str string) (dhash I2PDestHash, err error) {
	if strings.HasSuffix(str, ".b32.i2p") && len(str) == 60 {
		// valid
		_, err = i2pB32enc.Decode(dhash[:], []byte(str[:52]+"===="))
	} else {
		// invalid
		err = errors.New("invalid desthash format")
	}
	return
}

// create a desthash from a []byte array
func DestHashFromBytes(str []byte) (dhash *I2PDestHash, err error) {
	if len(str) == 32 {
		// valid
		//_, err = i2pB32enc.Decode(dhash[:], []byte(str[:52]+"===="))
		copy(dhash[:], str)
	} else {
		// invalid
		err = errors.New("invalid desthash format")
	}
	return
}

// get string representation of i2p dest hash(base32 version)
func (h I2PDestHash) String() string {
	b32addr := make([]byte, 56)
	i2pB32enc.Encode(b32addr, h[:])
	return string(b32addr[:52]) + ".b32.i2p"
}

// get base64 representation of i2p dest sha256 hash(the 44-character one)
func (h I2PDestHash) Hash() string {
	hash := sha256.New()
	hash.Write(h[:])
	digest := hash.Sum(nil)
	buf := make([]byte, 44)
	i2pB64enc.Encode(buf, digest)
	return string(buf)
}

// Returns "I2P"
func (h *I2PDestHash) Network() string {
	return "I2P"
}

// Returns the base64 representation of the I2PAddr
func (a I2PAddr) Base64() string {
	return string(a)
}

// Returns the I2P destination (base32-encoded)
func (a I2PAddr) String() string {
	return string(a.Base32())
}

// Returns "I2P"
func (a I2PAddr) Network() string {
	return "I2P"
}

// Creates a new I2P address from a base64-encoded string. Checks if the address
// addr is in correct format. (If you know for sure it is, use I2PAddr(addr).)
func NewI2PAddrFromString(addr string) (I2PAddr, error) {
	if strings.HasSuffix(addr, ".i2p") {
		if strings.HasSuffix(addr, ".b32.i2p") {
			return I2PAddr(""), errors.New("cannot convert .b32.i2p to full destination")
		}
		// strip off .i2p if it's there
		addr = addr[:len(addr)-4]
	}
	addr = strings.Trim(addr, "\t\n\r\f ")
	// very basic check
	if len(addr) > 4096 || len(addr) < 516 {
		return I2PAddr(""), errors.New("Not an I2P address")
	}
	buf := make([]byte, i2pB64enc.DecodedLen(len(addr)))
	if _, err := i2pB64enc.Decode(buf, []byte(addr)); err != nil {
		return I2PAddr(""), errors.New("Address is not base64-encoded")
	}
	return I2PAddr(addr), nil
}

func FiveHundredAs() I2PAddr {
	s := ""
	for x := 0; x < 517; x++ {
		s += "A"
	}
	r, _ := NewI2PAddrFromString(s)
	return r
}

// Creates a new I2P address from a byte array. The inverse of ToBytes().
func NewI2PAddrFromBytes(addr []byte) (I2PAddr, error) {
	if len(addr) > 4096 || len(addr) < 384 {
		return I2PAddr(""), errors.New("Not an I2P address")
	}
	buf := make([]byte, i2pB64enc.EncodedLen(len(addr)))
	i2pB64enc.Encode(buf, addr)
	return I2PAddr(string(buf)), nil
}

// Turns an I2P address to a byte array. The inverse of NewI2PAddrFromBytes().
func (addr I2PAddr) ToBytes() ([]byte, error) {
	return i2pB64enc.DecodeString(string(addr))
}

func (addr I2PAddr) Bytes() []byte {
	b, _ := addr.ToBytes()
	return b
}

// Returns the *.b32.i2p address of the I2P address. It is supposed to be a
// somewhat human-manageable 64 character long pseudo-domain name equivalent of
// the 516+ characters long default base64-address (the I2PAddr format). It is
// not possible to turn the base32-address back into a usable I2PAddr without
// performing a Lookup(). Lookup only works if you are using the I2PAddr from
// which the b32 address was generated.
func (addr I2PAddr) Base32() (str string) {
	return addr.DestHash().String()
}

func (addr I2PAddr) DestHash() (h I2PDestHash) {
	hash := sha256.New()
	b, _ := addr.ToBytes()
	hash.Write(b)
	digest := hash.Sum(nil)
	copy(h[:], digest)
	return
}

// Makes any string into a *.b32.i2p human-readable I2P address. This makes no
// sense, unless "anything" is an I2P destination of some sort.
func Base32(anything string) string {
	return I2PAddr(anything).Base32()
}

func NewDestination(samaddr string, sigType ...string) (I2PKeys, error) {
	if samaddr == "" {
		samaddr = "127.0.0.1:7656"
	}
	client, err := goSam.NewClient(samaddr)
	if err != nil {
		return I2PKeys{}, err
	}
	var sigtmp string
	if len(sigType) > 0 {
		sigtmp = sigType[0]
	}
	pub, priv, err := client.NewDestination(sigtmp)
	if err != nil {
		return I2PKeys{}, err
	}
	addr, err := NewI2PAddrFromBytes([]byte(pub))
	if err != nil {
		return I2PKeys{}, err
	}
	keys := NewKeys(addr, priv+pub)
	if err != nil {
		return I2PKeys{}, err
	}
	return keys, nil
}
