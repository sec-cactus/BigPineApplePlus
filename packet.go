package main

import (
	"fmt"
	//	"math/big"
	//	"net"
	//	"strconv"
	"bytes"
	"crypto/md5"
	"encoding/binary"
	"time"

	//	"net"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

var (
	blackTargetList  [1 << 29]uint8
	whiteTargetList  [1 << 29]uint8
	pblackTargetList *[1 << 29]uint8
	pwhiteTargetList *[1 << 29]uint8
	reverseSeqList   [1 << 16]uint32
	preverseSeqList  *[1 << 16]uint32
)

func initBullets() {
	preverseSeqList = &reverseSeqList
}

func socketToMD5Uint16(srcIP []byte, srcPort uint16, dstIP []byte, dstPort uint16) uint16 {
	bytebuf := bytes.NewBuffer([]byte{})
	binary.Write(bytebuf, binary.BigEndian, srcIP)
	binary.Write(bytebuf, binary.BigEndian, srcPort)
	binary.Write(bytebuf, binary.BigEndian, dstIP)
	binary.Write(bytebuf, binary.BigEndian, dstPort)

	retByte := md5.Sum(bytebuf.Bytes())
	ret := binary.BigEndian.Uint16(retByte[:])
	//ret = ret & ((1 << 18) - 1)
	return ret
}

func matchIPWithTargetList(ptargetList *[1 << 29]uint8, ip int64) bool {
	ip8 := ip >> 3                    // ip / 8
	ipmod8 := uint8(1 << uint8(ip&7)) // ip % 8
	rwlock.RLock()
	match := (*ptargetList)[ip8] & ipmod8
	rwlock.RUnlock()
	if match == 0 {
		return false
	} else if match > 0 {
		return true
	}

	return false
}

func forgeResetForward(ip *layers.IPv4, tcp *layers.TCP) gopacket.Packet {
	//fmt.Println("original packet: ", packet)

	//get eth, tcp, ip from original packet
	//ethLayer := packet.Layer(layers.LayerTypeEthernet)
	//eth, _ := ethLayer.(*layers.Ethernet)

	//remove vlan-id EthernetTypeDot1Q EthernetType = 0x8100
	//if eth.EthernetType != 0x0800 {
	//EthernetTypeIPv4 EthernetType = 0x0800
	//	eth.EthernetType = 0x0800
	//}

	//	ipLayer := packet.Layer(layers.LayerTypeIPv4)
	//	ip, _ := ipLayer.(*layers.IPv4)

	ip.Id = 0x0100
	ip.Flags = 0x0000

	//	tcpLayer := packet.Layer(layers.LayerTypeTCP)
	//	tcp, _ := tcpLayer.(*layers.TCP)

	options := gopacket.SerializeOptions{
		ComputeChecksums: true,
		FixLengths:       true,
	}

	//modify tcp flags and params
	tcp.RST = true
	tcp.NS = false
	tcp.CWR = false
	tcp.ECE = false
	tcp.URG = false
	tcp.PSH = false
	tcp.SYN = false
	tcp.FIN = false

	tcp.ACK = false

	tcp.Window = 32767
	//	tcp.Seq = tcp.Seq + uint32(1)
	tcp.Ack = uint32(1)
	tcp.Options = []layers.TCPOption{
		layers.TCPOption{
			OptionType:   layers.TCPOptionKind(layers.TCPOptionKindWindowScale),
			OptionLength: 0x03,
			OptionData:   []byte{0x0a},
		},
		layers.TCPOption{
			OptionType:   layers.TCPOptionKind(layers.TCPOptionKindNop),
			OptionLength: 0x01,
		},
		layers.TCPOption{
			OptionType:   layers.TCPOptionKind(layers.TCPOptionKindMSS),
			OptionLength: 0x04,
			OptionData:   []byte{0x01, 0x09},
		},
		layers.TCPOption{
			OptionType:   layers.TCPOptionKind(layers.TCPOptionKindTimestamps),
			OptionLength: 0x0a,
			OptionData:   []byte{0x3f, 0x3f, 0x3f, 0x3f, 0x00, 0x00, 0x00, 0x00},
		},
		layers.TCPOption{
			OptionType: layers.TCPOptionKind(layers.TCPOptionKindEndList),
		},
	}

	tcp.SetNetworkLayerForChecksum(ip)

	//assemble forward packet
	resetPacketBuffer := gopacket.NewSerializeBuffer()
	//fmt.Println(srcMac, dstMac)
	err := gopacket.SerializeLayers(resetPacketBuffer, options,
		&layers.Ethernet{
			SrcMAC:       srcMac,
			DstMAC:       dstMac,
			EthernetType: 0x0800,
		},
		ip,
		tcp,
	)
	//err := gopacket.SerializePacket(resetPacketBuffer, options, packet)
	if err != nil {
		panic(err)
	}
	resetPacket := gopacket.NewPacket(resetPacketBuffer.Bytes(), layers.LayerTypeEthernet, gopacket.Default)
	fmt.Println("1st request forward - reset packet: ", resetPacket)

	//calc resetPacketIndex
	//resetPacketIndex := getIndexByTCPIP(ip, tcp)
	//fmt.Println(resetPacketIndex)

	//write reset packet to bulletsList
	/*socketMD5 := uint32(0)
	if forward {
		socketMD5 = tcp.Seq & ((1 << 18) - 1)
	} else {
		socketMD5 = socketToMD5Uint32(ip.SrcIP, uint16(tcp.SrcPort), ip.DstIP, uint16(tcp.DstPort))
	}
	(*pbulletsList)[socketMD5] = resetPacket*/

	return resetPacket

	//return resetPacket, resetPacketIndex
}

func forgeResetReverse(reverseSeq uint32, ip *layers.IPv4, tcp *layers.TCP, lenPayload uint16) gopacket.Packet {
	//fmt.Println("original packet: ", packet)

	//get eth, tcp, ip from original packet
	//ethLayer := packet.Layer(layers.LayerTypeEthernet)
	//eth, _ := ethLayer.(*layers.Ethernet)

	//remove vlan-id EthernetTypeDot1Q EthernetType = 0x8100
	//if eth.EthernetType != 0x0800 {
	//EthernetTypeIPv4 EthernetType = 0x0800
	//	eth.EthernetType = 0x0800
	//}

	//ipLayer := packet.Layer(layers.LayerTypeIPv4)
	//ip, _ := ipLayer.(*layers.IPv4)
	/*
		ip.Id = 0x0100
		ip.Flags = 0x0000

		//tcpLayer := packet.Layer(layers.LayerTypeTCP)
		//tcp, _ := tcpLayer.(*layers.TCP)

		options := gopacket.SerializeOptions{
			ComputeChecksums: true,
			FixLengths:       true,
		}

		//modify tcp flags and params
		tcp.RST = true
		tcp.NS = false
		tcp.CWR = false
		tcp.ECE = false
		tcp.URG = false
		tcp.PSH = false
		tcp.SYN = false
		tcp.FIN = false

		tcp.ACK = false

		tcp.Window = 32767
		tcp.Seq = reverseSeq + uint32(1)
		tcp.Ack = uint32(1) + uint32(lenPayload)
		tcp.Options = []layers.TCPOption{
			layers.TCPOption{
				OptionType:   layers.TCPOptionKind(layers.TCPOptionKindWindowScale),
				OptionLength: 0x03,
				OptionData:   []byte{0x0a},
			},
			layers.TCPOption{
				OptionType:   layers.TCPOptionKind(layers.TCPOptionKindNop),
				OptionLength: 0x01,
			},
			layers.TCPOption{
				OptionType:   layers.TCPOptionKind(layers.TCPOptionKindMSS),
				OptionLength: 0x04,
				OptionData:   []byte{0x01, 0x09},
			},
			layers.TCPOption{
				OptionType:   layers.TCPOptionKind(layers.TCPOptionKindTimestamps),
				OptionLength: 0x0a,
				OptionData:   []byte{0x3f, 0x3f, 0x3f, 0x3f, 0x00, 0x00, 0x00, 0x00},
			},
			layers.TCPOption{
				OptionType: layers.TCPOptionKind(layers.TCPOptionKindEndList),
			},
		}
	*/

	options := gopacket.SerializeOptions{
		ComputeChecksums: true,
		FixLengths:       true,
	}
	tcp.Seq = reverseSeq + uint32(1)
	tcp.Ack = uint32(1) + uint32(lenPayload)
	tcp.SrcPort, tcp.DstPort = tcp.DstPort, tcp.SrcPort
	ip.SrcIP, ip.DstIP = ip.DstIP, ip.SrcIP

	tcp.SetNetworkLayerForChecksum(ip)

	//assemble forward packet
	resetPacketBuffer := gopacket.NewSerializeBuffer()
	//fmt.Println(srcMac, dstMac)
	err := gopacket.SerializeLayers(resetPacketBuffer, options,
		&layers.Ethernet{
			SrcMAC:       srcMac,
			DstMAC:       dstMac,
			EthernetType: 0x0800,
		},
		ip,
		tcp,
	)
	//err := gopacket.SerializePacket(resetPacketBuffer, options, packet)
	if err != nil {
		panic(err)
	}
	resetPacket := gopacket.NewPacket(resetPacketBuffer.Bytes(), layers.LayerTypeEthernet, gopacket.Default)
	fmt.Println("1st request reverse - reset packet: ", resetPacket)

	//calc resetPacketIndex
	//resetPacketIndex := getIndexByTCPIP(ip, tcp)
	//fmt.Println(resetPacketIndex)

	//write reset packet to bulletsList
	/*socketMD5 := uint32(0)
	if forward {
		socketMD5 = tcp.Seq & ((1 << 18) - 1)
	} else {
		socketMD5 = socketToMD5Uint32(ip.SrcIP, uint16(tcp.SrcPort), ip.DstIP, uint16(tcp.DstPort))
	}
	(*pbulletsList)[socketMD5] = resetPacket*/

	return resetPacket

	//return resetPacket, resetPacketIndex
}

/*
func getIndexByTCPIP(ip *layers.IPv4, tcp *layers.TCP) int64 {
	ipsrc := big.NewInt(0)
	ipsrc.SetBytes(ip.SrcIP.To4())
	portsrc := int(tcp.SrcPort)

	ipdst := big.NewInt(0)
	ipdst.SetBytes(ip.DstIP.To4())
	portdst := int(tcp.DstPort)
	//s1 := strconv.Itoa(int(ipsrc.Int64()))
	//s2 := strconv.Itoa(portsrc)
	//s3 := strconv.Itoa(int(ipdst.Int64()))
	//s4 := strconv.Itoa(portdst)
	s := strconv.Itoa(int(ipsrc.Int64()+ipdst.Int64())) + strconv.Itoa(portsrc) + strconv.Itoa(portdst)
	//fmt.Println(s1, s2, s3, s4, s)
	ret, _ := strconv.ParseInt(s, 10, 64)
	return ret
}
*/

/*
func analyseFirstRequestForward(ip *layers.IPv4, tcp *layers.TCP, handleMgt *pcap.Handle) {
	sendResetPacket(forgeResetForward(ip, tcp), handleMgt)
	fmt.Printf("1st request forward - Attacking %s and %s\n", ip.SrcIP, ip.DstIP)

		//white list first - label == 1
		if matchIPWithTargetList(pwhiteTargetList, ipStringToInt64(ip.SrcIP.String())) {
			fmt.Println("1st request forward - Src IP is white: ", ip.SrcIP.String())
			return
		} else if matchIPWithTargetList(pwhiteTargetList, ipStringToInt64(ip.DstIP.String())) {
			fmt.Println("1st request forward - Dst IP is white: ", ip.DstIP.String())
			return
			//black list - label == 0
		} else if matchIPWithTargetList(pblackTargetList, ipStringToInt64(ip.SrcIP.String())) {
			fmt.Println("1st request forward - Src IP is black: ", ip.SrcIP.String())
			//	resetPacketForward, resetPacketIndex := forgeReset(packet)
			//	lock.Lock()
			//	bulletsListMap[resetPacketIndex] = resetPacketForward
			//	lock.Unlock()
			sendResetPacket(forgeResetForward(packet), handleMgt)
			fmt.Printf("1st request forward - Attacking %s and %s\n", ip.SrcIP, ip.DstIP)
			return
		} else if matchIPWithTargetList(pblackTargetList, ipStringToInt64(ip.DstIP.String())) {
			fmt.Println("1st request forward - Dst IP is black: ", ip.DstIP.String())
			//	resetPacketForward, resetPacketIndex := forgeReset(packet)
			//	lock.Lock()
			//	bulletsListMap[resetPacketIndex] = resetPacketForward
			//	lock.Unlock()
			sendResetPacket(forgeResetForward(packet), handleMgt)
			fmt.Printf("1st request forward - Attacking %s and %s\n", ip.SrcIP, ip.DstIP)
			return
		}
		return

}*/

func analyseSecondHandShake(ip *layers.IPv4, tcp *layers.TCP) {
	//white list first - label == 1
	if matchIPWithTargetList(pwhiteTargetList, ipStringToInt64(ip.SrcIP.String())) {
		fmt.Println("2nd handshake - Src IP is white: ", ip.SrcIP.String())
		return
	} else if matchIPWithTargetList(pwhiteTargetList, ipStringToInt64(ip.DstIP.String())) {
		fmt.Println("2nd handshake - Dst IP is white: ", ip.DstIP.String())
		return
		//black list - label == 0
	} else if matchIPWithTargetList(pblackTargetList, ipStringToInt64(ip.SrcIP.String())) {
		fmt.Println("2nd handshake - Src IP is black: ", ip.SrcIP.String())
		//	resetPacketReverse, resetPacketIndex := forgeReset(packet)
		//	lock.Lock()
		//	bulletsListMap[resetPacketIndex] = resetPacketReverse
		//	lock.Unlock()
		socketMD5 := socketToMD5Uint16(ip.DstIP, uint16(tcp.DstPort), ip.SrcIP, uint16(tcp.SrcPort))
		(*preverseSeqList)[socketMD5] = tcp.Seq
		fmt.Println("2nd handshake - restore socket md5 and seq: ", socketMD5, tcp.Seq)
		time.Sleep(time.Millisecond * 100)
		(*preverseSeqList)[socketMD5] = 0

		hitTarget(ip.SrcIP.String())
		return
	} else if matchIPWithTargetList(pblackTargetList, ipStringToInt64(ip.DstIP.String())) {
		fmt.Println("2nd handshake - Dst IP is black: ", ip.DstIP.String())
		//	resetPacketReverse, resetPacketIndex := forgeReset(packet)
		//	lock.Lock()
		//	bulletsListMap[resetPacketIndex] = resetPacketReverse
		//	lock.Unlock()
		socketMD5 := socketToMD5Uint16(ip.DstIP, uint16(tcp.DstPort), ip.SrcIP, uint16(tcp.SrcPort))
		(*preverseSeqList)[socketMD5] = tcp.Seq
		fmt.Println("2nd handshake - restore socket md5 and seq: ", socketMD5, tcp.Seq)
		time.Sleep(time.Millisecond * 100)
		(*preverseSeqList)[socketMD5] = 0

		hitTarget(ip.SrcIP.String())
		return
	}

	return
}

func sendResetPacket(handleMgt *pcap.Handle, c chan [2]gopacket.Packet) {
	for {
		packets := <-c
		if err := handleMgt.WritePacketData(packets[0].Data()); err != nil {
			fmt.Println("Send error", err.Error())
		}
		if err := handleMgt.WritePacketData(packets[1].Data()); err != nil {
			fmt.Println("Send error", err.Error())
		}
	}
}

// func analyseFirstRequestReverse(reverseSeq uint32, ip *layers.IPv4, tcp *layers.TCP, handleMgt *pcap.Handle) {
// 	//calc forward packet index
// 	//resetPacketForwardIndex := getIndexByTCPIP(ip, tcp)
// 	//find and send forward reset packet
// 	//lock.RLock()
// 	//resetPacketForward, exist1 := bulletsListMap[resetPacketForwardIndex]
// 	//lock.RUnlock()
// 	//socketMD5 := tcp.Seq & ((1 << 18) - 1)
// 	//resetPacketForward := (*pbulletsList)[socketMD5]
// 	//(*pbulletsList)[socketMD5] = nil
// 	//if resetPacketForward != nil {
// 	//	fmt.Println("first request index: ", socketMD5)
// 	//	fmt.Printf("Attacking %s and %s\n", ip.SrcIP, ip.DstIP)
// 	//	sendResetPacket(resetPacketForward, handleMgt)
// 	//	lock.Lock()
// 	//	delete(bulletsListMap, resetPacketForwardIndex)
// 	//	lock.Unlock()
// 	//}

// 	//calc reverse packet index
// 	//tcp.SrcPort, tcp.DstPort = tcp.DstPort, tcp.SrcPort
// 	//ip.SrcIP, ip.DstIP = ip.DstIP, ip.SrcIP
// 	//resetPacketReverseIndex := getIndexByTCPIP(ip, tcp)
// 	//find reverse reset packet
// 	//lock.Lock()
// 	//resetPacketReverse, exist2 := bulletsListMap[resetPacketReverseIndex]
// 	//delete(bulletsListMap, resetPacketForwardIndex)
// 	//lock.Unlock()
// 	//	fmt.Println("1st request reverse - get socket md5 and previous seq: ", socketMD5, reverseSeq)
// 	resetPacketReverse := forgeResetReverse(reverseSeq, ip, tcp, uint16(len(tcp.Payload)))
// 	fmt.Printf("1st request reverse - Attacking %s and %s\n", ip.DstIP, ip.SrcIP)
// 	sendResetPacket(resetPacketReverse, handleMgt)
// 	//	if reverseSeq != 0 {
// 	//marked by 2nd handshake, attack
// 	//		fmt.Println("1st request reverse - get socket md5 and previous seq: ", socketMD5, reverseSeq)
// 	//		resetPacketReverse := forgeResetReverse(reverseSeq, ip, tcp, uint16(len(tcp.Payload)))
// 	//prepare ip for checksum
// 	ipLayer0 := resetPacketReverse.Layer(layers.LayerTypeIPv4)
// 	ip0, _ := ipLayer0.(*layers.IPv4)

// 	//modify tcp ack
// 	tcpLayer0 := resetPacketReverse.Layer(layers.LayerTypeTCP)
// 	tcp0, _ := tcpLayer0.(*layers.TCP)
// 	tcp0.Ack = uint32(1) + uint32(len(tcp.Payload))
// 	tcp0.SetNetworkLayerForChecksum(ip0)

// 	options := gopacket.SerializeOptions{
// 		ComputeChecksums: true,
// 		FixLengths:       true,
// 	}

// 	resetPacketBuffer := gopacket.NewSerializeBuffer()
// 	err := gopacket.SerializePacket(resetPacketBuffer, options, resetPacketReverse)
// 	if err != nil {
// 		panic(err)
// 	}
// 	resetPacket := gopacket.NewPacket(resetPacketBuffer.Bytes(), layers.LayerTypeEthernet, gopacket.Default)*/

// 	//		fmt.Printf("1st request reverse - Attacking %s and %s\n", ip.DstIP, ip.SrcIP)
// 	//		sendResetPacket(resetPacketReverse, handleMgt)
// 	//}

// 	return
// }

func analysePacket(packet gopacket.Packet, handleMgt *pcap.Handle, c chan [2]gopacket.Packet) {
	//fmt.Println("get routine ", GoID())
	ipLayer := packet.Layer(layers.LayerTypeIPv4)
	if ipLayer != nil {
		tcpLayer := packet.Layer(layers.LayerTypeTCP)
		tcp, _ := tcpLayer.(*layers.TCP)
		ip, _ := ipLayer.(*layers.IPv4)
		if tcpLayer != nil {
			//	if tcp.SYN && !(tcp.RST || tcp.FIN || tcp.PSH || tcp.ACK) {
			//		analyseFirstHandShake(ip, packet, handleMgt)
			//		return
			//	} else
			if (tcp.SYN && tcp.ACK) && !(tcp.RST || tcp.FIN || tcp.PSH) {
				analyseSecondHandShake(ip, tcp)
				return
			} else if (tcp.PSH && tcp.ACK) && !(tcp.RST || tcp.FIN || tcp.SYN) {
				socketMD5 := socketToMD5Uint16(ip.SrcIP, uint16(tcp.SrcPort), ip.DstIP, uint16(tcp.DstPort))
				reverseSeq := (*preverseSeqList)[socketMD5]
				(*preverseSeqList)[socketMD5] = 0

				if reverseSeq != 0 {
					//marked by 2nd handshake
					if tcp.Seq == (reverseSeq + uint32(1)) {
						//1st request reverse, return
						return
					} else {
						fmt.Println("1st request - get socket md5 and previous seq: ", socketMD5, reverseSeq)

						resetPacketForward := forgeResetForward(ip, tcp)
						resetPacketReverse := forgeResetReverse(reverseSeq, ip, tcp, uint16(len(tcp.Payload)))

						fmt.Printf("1st request reverse - Attacking %s and %s\n", ip.DstIP, ip.SrcIP)

						c <- [2]gopacket.Packet{resetPacketReverse, resetPacketForward}

						//	go sendResetPacket(resetPacketReverse, handleMgt)

						//	go sendResetPacket(resetPacketForward, handleMgt)
						fmt.Printf("1st request forward - Attacking %s and %s\n", ip.SrcIP, ip.DstIP)

						//	analyseFirstRequestForward(ip, tcp, handleMgt)
						//	analyseFirstRequestReverse(reverseSeq, ip, tcp, handleMgt)
						return
					}
				} else {
					//was not marked by 2nd handshake, which means not black socket, return
					return
				}
			}
		}
	}
	return
}
