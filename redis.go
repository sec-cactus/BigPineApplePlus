package main

import (
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/gomodule/redigo/redis"
)

var (
	mirrorNetworkDevice string
	mgtNetworkDevice    string
	redisConn           redis.Conn
)

func getRedisConn(redisConf map[string]string) {
	redisconn, err := redis.Dial("tcp", redisConf["redis_address"])
	if err != nil {
		fmt.Println("Connect to redis error", err)
		return
	} else if redisconn != nil {
		redisConn = redisconn
		fmt.Println("Connect to redis done")
		return
	}
	// defer conn.Close()

	// _, err = conn.Do("SET", "mykey", "superWang")
	// if err != nil {
	// 	fmt.Println("redis set failed:", err)
	// }

	// username, err := redis.String(conn.Do("GET", "mykey"))
	// if err != nil {
	// 	fmt.Println("redis get failed:", err)
	// } else {
	// 	fmt.Printf("Get mykey: %v \n", username)
	// }
}

func initMacs(srcmac string, dstmac string) (net.HardwareAddr, net.HardwareAddr) {
	//init nic mac for send packets
	srcMacAddr, errSrcMac := net.ParseMAC(srcmac)
	dstMacAddr, errDstMac := net.ParseMAC(dstmac)
	if errSrcMac != nil {
		fmt.Println("Set mac error", errSrcMac.Error())
		return nil, nil
	}
	if errDstMac != nil {
		fmt.Println("Set mac error", errDstMac.Error())
		return nil, nil
	}
	return srcMacAddr, dstMacAddr
}

func getSysConfig() {
	if redisConn != nil {
		sysconfig, err := redis.Strings(redisConn.Do("HMGET", "sysconfig", "mirrornetworkdevice", "mgtnetworkdevice", "srcmac", "dstmac"))
		if err != nil {
			fmt.Println("redis get failed:", err)
		} else {
			mirrorNetworkDevice = sysconfig[0]
			mgtNetworkDevice = sysconfig[1]
			srcMac, dstMac = initMacs(sysconfig[2], sysconfig[3])

			fmt.Printf("Get sysconfig:", mirrorNetworkDevice, mgtNetworkDevice, srcMac, dstMac)

			return
		}
	}
	return
}

func ipIntervalToTargetList(intervals string, ptargetList *[1 << 29]uint8) *[1 << 29]uint8 {
	//去除单行属性两端的空格
	intervals = strings.TrimSpace(intervals)

	//判断等号=在该行的位置
	intervalIndex := strings.Index(intervals, "-")
	if intervalIndex < 0 {
		fmt.Println("fail to get index")
		return ptargetList
	}
	//取得等号左边的start值，判断是否为空
	startString := strings.TrimSpace(intervals[:intervalIndex])
	if len(startString) == 0 {
		fmt.Println("fail to get start string")
		return ptargetList
	}

	//取得等号右边的end值，判断是否为空
	endString := strings.TrimSpace(intervals[intervalIndex+1:])
	if len(endString) == 0 {
		fmt.Println("fail to get end string")
		return ptargetList
	}

	start := ipStringToInt64(startString)
	end := ipStringToInt64(endString)

	for i := start; i < (end + 1); i++ {
		(*ptargetList)[(i / 8)] = (1 << uint8(i%8))
	}

	return ptargetList
}

func getTargetList() {
	if redisConn != nil {

		rwlock.Lock()

		blackTargetList = [1 << 29]uint8{}
		whiteTargetList = [1 << 29]uint8{}
		pblackTargetList = &blackTargetList
		pwhiteTargetList = &whiteTargetList

		blacktargetlist, err := redis.Strings(redisConn.Do("SMEMBERS", "blacktargetlist"))
		if err != nil {
			fmt.Println("redis get failed:", err)
		} else {
			fmt.Printf("Get blacktargetlist: %v \n", blacktargetlist)
			for targetindex := range blacktargetlist {
				target := blacktargetlist[targetindex]
				//interval or single ip
				index := strings.Index(target, "-")
				if index < 0 {
					//calc target/8 and target%8
					target8 := ipStringToInt64(target) / 8
					targetmod8 := ipStringToInt64(target) % 8
					//black list
					(*pblackTargetList)[target8] = (1 << uint8(targetmod8))
					continue
				} else if index > -1 {
					//interval
					pblackTargetList = ipIntervalToTargetList(target, pblackTargetList)
					continue
				}
			}
		}

		whitetargetlist, err := redis.Strings(redisConn.Do("SMEMBERS", "whitetargetlist"))
		if err != nil {
			fmt.Println("redis get failed:", err)
		} else {
			fmt.Printf("Get whitetargetlist: %v \n", whitetargetlist)
			for targetindex := range whitetargetlist {
				target := whitetargetlist[targetindex]
				//interval or single ip
				index := strings.Index(target, "-")
				if index < 0 {
					//calc target/8 and target%8
					target8 := ipStringToInt64(target) / 8
					targetmod8 := ipStringToInt64(target) % 8
					//black list
					(*pwhiteTargetList)[target8] = (1 << uint8(targetmod8))
					continue
				} else if index > -1 {
					//interval
					pwhiteTargetList = ipIntervalToTargetList(target, pwhiteTargetList)
					continue
				}
			}
		}

		rwlock.Unlock()
	}
}

func initTargetList() {
	getTargetList()
	go refreshTargetList()
}

func refreshTargetList() {
	time.Sleep(time.Minute * 1)
	getTargetList()
	refreshTargetList()
}
