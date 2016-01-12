/*
Implements the solution to assignment 1 for UBC CS 416 2015 W2.


Usage:
$ go run client.go [local UDP ip:port] [aserver UDP ip:port] [secret]

Example:
$ go run client.go 127.0.0.1:2020 127.0.0.1:7070 1984

*/

package main

import (
	"crypto/md5"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strconv"
)

/////////// Msgs used by both auth and fortune servers:

// An error message from the server.
type ErrMessage struct {
	Error string
}

/////////// Auth server msgs:

// Message containing a nonce from auth-server.
type NonceMessage struct {
	Nonce int64
}

// Message containing an MD5 hash from client to auth-server.
type HashMessage struct {
	Hash string
}

// Message with details for contacting the fortune-server.
type FortuneInfoMessage struct {
	FortuneServer string
	FortuneNonce  int64
}

/////////// Fortune server msgs:

// Message requesting a fortune from the fortune-server.
type FortuneReqMessage struct {
	FortuneNonce int64
}

// Response from the fortune-server containing the fortune.
type FortuneMessage struct {
	Fortune string
}

//helper method for computing MD5 hash
func computeMD5Hash(toBeEncoded int64) string {
	var computedString string

	byteArray := make([]byte, 1024)
	n := binary.PutVarint(byteArray, toBeEncoded)

	computedMD5 := md5.Sum(byteArray[0:n])

	computedString = fmt.Sprintf("%x", computedMD5)

	return computedString

}

// Main workhorse method.
func main() {

	//parse args
	localAddress := os.Args[1]
	aServerArg := os.Args[2]
	secret, _ := strconv.ParseInt(os.Args[3], 10, 64)

	myAddress, _ := net.ResolveUDPAddr("udp", localAddress)

	aServerAddress, err := net.ResolveUDPAddr("udp", aServerArg)
	aServerConn, err := net.DialUDP("udp", myAddress, aServerAddress)
	if err != nil {
		fmt.Println("Error dialing server", err)
		os.Exit(-1)
	}

	//sending arbitrary message
	_, sendErr := aServerConn.Write([]byte("Hello"))
	if err != nil {
		fmt.Printf("Couldn't send arbitrary message %v", sendErr)
	}

	//reading the nonce response from server
	var buf [1024]byte
	num, err := aServerConn.Read(buf[:])
	if err != nil {
		fmt.Println("Error on read: ", err)
		os.Exit(-1)
	}

	//Decoding the nonce response
	var nonceMessage NonceMessage
	nonceDecodeErr := json.Unmarshal(buf[0:num], &nonceMessage)
	if nonceDecodeErr != nil {
		fmt.Println("Error on decode Nonce Response: ", nonceDecodeErr)
		os.Exit(-1)
	}

	//calculate MD5 and marshal
	toBeEncoded := nonceMessage.Nonce + secret
	var EncodedMD5 string

	EncodedMD5 = computeMD5Hash(toBeEncoded)

	hashMessage := HashMessage{EncodedMD5}
	//hashString
	md5Json, md5JErr := json.Marshal(hashMessage)
	if md5JErr != nil {
		fmt.Printf("Couldn't marshall MD5 %v", md5JErr)
	}

	//sendMD5

	_, md5Err := aServerConn.Write(md5Json)
	if md5Err != nil {
		fmt.Printf("Couldn't send MD5 %v", md5Err)
	}

	//reading the fserver info response from aserver
	var fserverbuf [1024]byte
	numToRead, err := aServerConn.Read(fserverbuf[:])
	if err != nil {
		fmt.Println("Error on fserver info read: ", err)
		os.Exit(-1)
	}

	//Decoding the fserver info response
	var fortuneInfo FortuneInfoMessage
	fserverInfoDecodeErr := json.Unmarshal(fserverbuf[0:numToRead], &fortuneInfo)
	if fserverInfoDecodeErr != nil {
		fmt.Println("Error on fserverinfo decode: ", fserverInfoDecodeErr)
		os.Exit(-1)
	}
	aServerConn.Close()

	fServerAddress, err := net.ResolveUDPAddr("udp", fortuneInfo.FortuneServer)
	fServerConn, err := net.DialUDP("udp", myAddress, fServerAddress)
	if err != nil {
		fmt.Println("Error dialing fserver", err)
		os.Exit(-1)
	}

	fortuneReqMessage := FortuneReqMessage{fortuneInfo.FortuneNonce}
	fReqJson, _ := json.Marshal(fortuneReqMessage)

	fServerConn.Write(fReqJson)

	var fortuneBuffer [1024]byte
	numRead, fortuneReadError := fServerConn.Read(fortuneBuffer[:])
	if err != nil {
		fmt.Println("Error reading fortune: ", fortuneReadError)
		os.Exit(-1)
	}

	//Decoding the fortune response
	var fortune FortuneMessage
	fortuneDecodeErr := json.Unmarshal(fortuneBuffer[0:numRead], &fortune)
	if fortuneDecodeErr != nil {
		fmt.Println("Error on fserverinfo decode: ", fortuneDecodeErr)
		os.Exit(-1)
	}

	//Print out the fortune
	println(fortune.Fortune)

	fServerConn.Close()

}
