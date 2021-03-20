package main

// radiosim is a program to look like a openHPSDR radio and tangerineSDR radio for the purpose of
// testing the discovery tool
//
// by Dave Larsen KV0S
// code at github.com/kv0s/radiosim
//
// licensed under the GPL3

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/naoina/toml"
)

//Config parameters for the radiosim program
type Config struct {
	Radio    string
	Aport    string
	Bport    string
	Version  string
	Protocol string
	Status   string
}

type packetBt struct {
	Status   byte
	radioMAC []byte
	Bport    []byte
	Version  byte
	Board    byte
}

func radiostringtohex(ver string) uint64 {
	v := strings.Replace(ver, ".", "", -1)
	vers, _ := strconv.ParseUint(v, 10, 16)
	return vers
}

func main() {
	var cfg Config
	var radioname string
	var pBt packetBt

	const progname = "Radio Simulator"
	const progvers = "0.1.0"
	const configname = "config.toml"

	fmt.Printf("%s, version:%s, config:%s\n", progname, progvers, configname)

	f, err := os.Open("config.toml")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	if err := toml.NewDecoder(f).Decode(&cfg); err != nil {
		log.Fatal(err)
	}

	stat := strings.ToLower(cfg.Status)
	if stat == "idle" {
		pBt.Status = 0x02
	} else {
		pBt.Status = 0x03
	}

	radioname = strings.ToLower(cfg.Radio)
	if radioname == "metis" {
		pBt.Board = 0x00
	} else if radioname == "hermes" {
		pBt.Board = 0x01
	} else if radioname == "griffin" {
		pBt.Board = 0x02
	} else if radioname == "angelia" {
		pBt.Board = 0x04
	} else if radioname == "orion" {
		pBt.Board = 0x05
	} else if radioname == "tangerinesdr" {
		pBt.Board = 0x0a
	}

	inf, err := net.Interfaces()
	if err != nil {
		panic(err)
	}

	for _, i := range inf {
		//Check interface name
		if i.Name != "lo" {
			pBt.radioMAC = i.HardwareAddr
		}
	}

	ver := radiostringtohex(cfg.Version)
	// fmt.Printf("%#v %x\n", ver, ver)
	pBt.Version = byte(ver)

	bport, _ := strconv.ParseUint(cfg.Bport, 10, 16)
	fmt.Printf("%T %04x\n", bport, bport)

	b := make([]byte, 2)
	binary.BigEndian.PutUint16(b, uint16(bport))
	pBt.Bport = b

	//pBt.Bport = []byte{0xaa, 0x35}
	fmt.Printf("pBt.Bport %#v\n", pBt.Bport)

	// fmt.Printf(" radioMAC %#v\n", pBt.radioMAC)

	pc, err := net.ListenPacket("udp4", ":1024")
	if err != nil {
		panic(err)
	}
	defer pc.Close()

	for {
		buf := make([]byte, 64)
		n, addr, err := pc.ReadFrom(buf)
		if err != nil {
			panic(err)
		}

		go handleConnection(pc, n, addr, buf, pBt)
	}
}

// handleConnection is a goroutine to handle on connection at a time
func handleConnection(pc net.PacketConn, n int, addr net.Addr, buf []byte, pkt packetBt) {

	fmt.Printf("Received from %s %x length=%d\n", addr, buf[:n], len(buf))

	rbuf := make([]byte, 64)

	rbuf, err := hex.DecodeString("effe")
	if err != nil {
		panic(err)
	}

	// fmt.Printf("test %#v length=%d\n", rbuf, len(rbuf))
	rbuf = append(rbuf, pkt.Status)
	// fmt.Printf("test %#v %#v length=%d\n", rbuf, pkt.Status, len(rbuf))
	rbuf = append(rbuf, pkt.radioMAC...)

	// fmt.Printf("test MAC %#v Version %#v \n", pkt.radioMAC, pkt.Version)
	rbuf = append(rbuf, pkt.Version)
	// fmt.Printf("test %#v %#v length=%d\n", rbuf, pkt.Version, len(rbuf))
	rbuf = append(rbuf, pkt.Board)
	// fmt.Printf("test %#v %#v length=%d\n", rbuf, pkt.Board, len(rbuf))
	rbuf = append(rbuf, pkt.Bport...)
	fmt.Printf("test %#v %#v length=%d\n", rbuf, pkt.Bport, len(rbuf))

	for i := 1; i < 53; i++ {
		rbuf = append(rbuf, 0x00)
	}

	fmt.Printf("Sent to %#v %x length=%d\n", addr, rbuf[:n], len(rbuf))
	pc.WriteTo(rbuf, addr)
}
