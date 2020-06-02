package main

import (
	"flag"
	"fmt"
	"net"
	"os"

	// "github.com/google/gopacket/pfring"

	"sync"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
)

var (
	// debug  bool
	srcMac net.HardwareAddr
	dstMac net.HardwareAddr
	rwlock sync.RWMutex

//	lock             sync.RWMutex
)

func main() {

	// Get conf file from flag
	parseFlags()
	// Exit on invalid parameters
	flagsComplete, errString := flagsComplete()
	if !flagsComplete {
		fmt.Println(errString)
		flag.PrintDefaults()
		os.Exit(1)
	}

	//get config from redis
	redisConf := initConfig(confFile)
	getRedisConn(redisConf)

	getSysConfig()

	fmt.Println("Get configs from redis done")

	//init modules
	initTargetList()
	initBullets()
	initStats()

	fmt.Println("Init targetlist, bullets, stats done")

	// Open connection
	handleMirror, errHandleMirror := pcap.OpenLive(
		mirrorNetworkDevice, // network device
		int32(65535),
		true,
		time.Microsecond,
	)
	if errHandleMirror != nil {
		fmt.Println("Mirror Handler error", errHandleMirror.Error())
	}

	handleMgt, errHandleMgt := pcap.OpenLive(
		mgtNetworkDevice, // network device
		int32(65535),
		false,
		time.Microsecond,
	)
	if errHandleMgt != nil {
		fmt.Println("MGT Handler error", errHandleMgt.Error())
	}

	//init send packet channel
	c := make(chan [2]gopacket.Packet)
	go sendResetPacket(handleMgt, c)

	//Close when done
	//defer handleMirror.Close()
	//defer handleMgt.Close()

	//Capture Live Traffic
	packetSource := gopacket.NewPacketSource(handleMirror, handleMirror.LinkType())
	for packet := range packetSource.Packets() {
		go analysePacket(packet, handleMgt, c)
	}

	// if ring, err := pfring.NewRing(mirrorNetworkDevice, 65536, pfring.FlagPromisc); err != nil {
	// 	panic(err)
	// 	// } else if err := ring.SetBPFFilter("tcp and port 80"); err != nil {
	// 	// 	// optional
	// 	// 	panic(err)
	// } else if err := ring.Enable(); err != nil {
	// 	// Must do this!, or you get no packets!
	// 	panic(err)
	// } else {
	// 	packetSource := gopacket.NewPacketSource(ring, layers.LinkTypeEthernet)
	// 	for packet := range packetSource.Packets() {
	// 		// handlePacket(packet) // Do something with a packet here.
	// 		go analysePacket(packet, handleMgt, c)
	// 	}
	// }

}
