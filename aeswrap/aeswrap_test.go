package aeswrap

import (
	"encoding/base64"
	"fmt"
	"regexp"
	"testing"
)

func TestAes128(t *testing.T) {
	//key的长度必须是16、24或者32字节，分别用于选择AES-128, AES-192, or AES-256
	var aeskey = []byte("12345678abcdefgh")
	pass := []byte("vdncloud123456")
	xpass, err := AesEncrypt(pass, aeskey)
	if err != nil {
		fmt.Println(err)
		return
	}

	pass64 := base64.StdEncoding.EncodeToString(xpass)
	fmt.Printf("加密后:%v\n", pass64)

	bytesPass, err := base64.StdEncoding.DecodeString(pass64)
	if err != nil {
		fmt.Println(err)
		return
	}

	tpass, err := AesDecrypt(bytesPass, aeskey)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("解密后:%s\n", tpass)
}

func TestRegex(t *testing.T) {
	a := "#EXT-X-KEY:METHOD=AES-128,URI=\"https://misc.geekbang.org/serv/v1/decrypt/decryptkms/?Ciphertext=MjUyNWM3ZTQtZGM4Mi00NzI4LTg3YjAtMjk1YWE5N2ViMDg5Sjh2TzJ1NERYU0IyazlEVFdLbGZZZitYMjBhL202U01BQUFBQUFBQUFBQXY2Y2cxNWQxM2pMZ1crWE8vWDR5WFdHUngxNldLbVlPVEp3enlTdC84RnFPYlJhK0g0QWVL&MediaId=c4dddb5fe1a64490939f0ec2e46186ef&MtsHlsUriToken=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJjb2RlIjoiNDFlN2UzNzhlYjhjOGExYmZiODczODczNzQyYTIwZmZkNzIwN2ZhNCIsImV4cCI6MTY0OTIxNTc1MDQxMSwiZXh0cmEiOnsidmlkIjoiYzRkZGRiNWZlMWE2NDQ5MDkzOWYwZWMyZTQ2MTg2ZWYiLCJhaWQiOjIxODQsInVpZCI6MTI1Mjk0MCwicyI6IiJ9LCJzIjoyLCJ0IjoxLCJ2IjoxfQ.sacehfECZw3wyGj4kRFutptGOAT55hfq8aJJIOIkm3Y\""
	re := regexp.MustCompile("#EXT-X-KEY:METHOD=AES-128,URI=\"(.*)\"")

	fmt.Printf("%s\n", re.FindStringSubmatch(a)[1])
}
