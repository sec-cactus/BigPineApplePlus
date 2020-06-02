package main

import (
	// "bytes"
	// "crypto/md5"
	// "encoding/binary"

	//	"fmt"
	"math/big"
	"net"
	//	"runtime"
	//	"strconv"
	//	"strings"
)

/*
func stringToMD5Int(s string) int64 {
	//key取hash
	shash := md5.New()
	sbyte := []byte(s)
	shash.Write(sbyte)
	buf := bytes.NewBuffer(shash.Sum(nil))
	var shex int64
	binary.Read(buf, binary.BigEndian, &shex)
	return shex
}
*/

func ipStringToInt64(s string) int64 {
	ret := big.NewInt(0)
	ret.SetBytes(net.ParseIP(s).To4())
	return ret.Int64()
}

/*
func macStringToBytes(s string) []byte {
	ss := strings.Split(s, ":")
	bb := make([]byte, len(ss))
	i := int64(0x00)
	for index := range ss {
		i, _ = strconv.ParseInt(ss[index], 16, 0)
		bb[index] = uint8(i)

	}
	return bb
}
*/

/*
func bytesToInt64(b []byte) int64 {
	ret := big.NewInt(0)
	ret.SetBytes(b)
	return ret.Int64()
}
*/

/*
func ipBytesToInt32(bytes []byte) int32 {
	ipBytes := big.NewInt(0)
	ipBytes.SetBytes(bytes)
	return int32(ipBytes.Int64())
}
*/

/*
func ipIntervalToIPInt32s(intervals string) []int32 {
	//去除单行属性两端的空格
	intervals = strings.TrimSpace(intervals)

	//判断等号=在该行的位置
	intervalIndex := strings.Index(intervals, "-")
	if intervalIndex < 0 {
		fmt.Println("fail to get index")
		return nil
	}
	//取得等号左边的start值，判断是否为空
	startString := strings.TrimSpace(intervals[:intervalIndex])
	if len(startString) == 0 {
		fmt.Println("fail to get start string")
		return nil
	}

	//取得等号右边的end值，判断是否为空
	endString := strings.TrimSpace(intervals[intervalIndex+1:])
	if len(endString) == 0 {
		fmt.Println("fail to get end string")
		return nil
	}

	start := ipStringToInt64(startString)
	end := ipStringToInt64(endString)

	ret := make([]int32, (end - start + 1))

	for i := start; i < (end + 1); i++ {
		ret[i-start] = int32(i)
	}

	return ret
}
*/

// func GoID() int {
// 	/*var buf [64]byte
// 	n := runtime.Stack(buf[:], false)

// 	idField := strings.Fields(strings.TrimPrefix(string(buf[:n]), "goroutine "))[0]
// 	GoroutineId, err := strconv.Atoi(idField)
// 	if err != nil {
// 		panic(fmt.Sprintf("cannot get goroutine id: %v", err))
// 	}*/

// 	return runtime.NumGoroutine()
// }
