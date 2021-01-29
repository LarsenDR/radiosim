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
	Version  string
	Protocol string
	Status   string
}

type packetBt struct {
	Status   byte
	radioMAC []byte
	Version  byte
	Board    byte
}

func main() {
	var cfg Config
	var radioname string
	var pBt packetBt

	f, err := os.Open("config.toml")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	if err := toml.NewDecoder(f).Decode(&cfg); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Config:\n Radio %s\n Version %#v\n Protocol %s\n status %s\n", cfg.Radio, cfg.Version, cfg.Protocol, cfg.Status)

	if cfg.Status == "idle" {
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
	} else if radioname == "orian" {
		pBt.Board = 0x05
	} else if radioname == "hermes_lite" {
		pBt.Board = 0x06
	} else if radioname == "tangerinesdr" {
		pBt.Board = 0x07
	}

	// Conver the cfg string to decimal number
	// convert decimal to hex number

	vers := strings.Replace(cfg.Version, ".", "", -1)
	fmt.Printf("**** vers %#v\n", vers)
	version, _ := strconv.ParseInt(vers, 16, 64)
	fmt.Printf("**** version %#v\n", version)
	err = binary.Write(pBt.Version, binary.LittleEndian, version)
	//pBt.Version = []byte(vers)

	fmt.Printf("**** Version %#v\n", pBt.Version)
	if err != nil {
		panic(err)
	}

	inf, err := net.Interfaces()
	if err != nil {
		panic(err)
	}

	fmt.Printf("%#v\n", inf)

	for _, i := range inf {
		//Check interface name
		if i.Name != "lo" {
			pBt.radioMAC = i.HardwareAddr
		}
	}

	fmt.Printf(" radioMAC %#v\n", pBt.radioMAC)

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

	rbuf = append(rbuf, pkt.Status)
	rbuf = append(rbuf, pkt.radioMAC...)
	rbuf = append(rbuf, pkt.Version)
	rbuf = append(rbuf, pkt.Board)

	for i := 1; i < 50; i++ {
		rbuf = append(rbuf, 0x00)
	}

	fmt.Printf("Sent to %s %x length=%d\n", addr, rbuf[:n], len(rbuf))
	pc.WriteTo(rbuf, addr)
}
