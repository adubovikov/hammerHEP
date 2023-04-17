package publish

import (
	"bufio"
	fmt "fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func GeneratePacketsArray() []Packet {

	AnnouncementCall := []Packet{
		Packet{
			Version:   3,
			Protocol:  17,
			SrcIP:     net.ParseIP("10.0.0.6"),
			DstIP:     net.ParseIP("120.1.1.1"),
			SrcPort:   5060,
			DstPort:   5060,
			Tsec:      0,
			Tmsec:     0,
			ProtoType: 1,
			Payload: []byte(`

			`),
		},
	}

	return AnnouncementCall
}

func RandStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func GeneratePacketsArrayFromText(fileName string, replaceCallid bool) []Packet {
	var scenario []Packet

	file, err := os.Open(fileName)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	//proto:UDP 2023-04-17T14:56:19.323629+02:00  10.0.0.2:5060 ---> 192.168.178.20:61000

	scanner := bufio.NewScanner(file)
	// optionally, resize scanner's capacity for lines over 64K, see next example
	firtLine := false
	startMessage := false
	sipMessage := ""
	protoString := ""
	dateString := ""
	var error error
	var date time.Time
	srcIP := ""
	dstIP := ""
	var srcPort uint16
	var dstPort uint16
	generatedCallID := RandStringBytes(30)

	var hepPacket Packet

	for scanner.Scan() {
		lineData := scanner.Text()

		if strings.HasPrefix(lineData, "proto:") {
			fmt.Println("HAS DATA:", lineData)
			lineArray := strings.Split(lineData, " ")
			protoString = strings.TrimPrefix(lineArray[0], "proto:")
			dateString := lineArray[1]

			if sipMessage != "" {
				fmt.Println("MIDDLE:", sipMessage)
				hepPacket.Payload = []byte(sipMessage)
				scenario = append(scenario, hepPacket)
			}

			date, error = time.Parse(time.RFC3339Nano, dateString)

			if error != nil {
				fmt.Println(error)
			}

			srcIPPortString := lineArray[3]
			dstIPPortString := lineArray[5]

			srcData := strings.SplitN(srcIPPortString, ":", 2)
			srcIP = srcData[0]
			ui64, _ := strconv.ParseUint(srcData[1], 10, 64)
			srcPort = uint16(ui64)

			dstData := strings.SplitN(dstIPPortString, ":", 2)
			dstIP = dstData[0]
			ui64, _ = strconv.ParseUint(dstData[1], 10, 64)
			dstPort = uint16(ui64)

			fmt.Println("Proto:", protoString)
			fmt.Println("dateString:", dateString)
			fmt.Println("date:", date.String())
			fmt.Println("src:", srcIP)
			fmt.Println("src:", srcPort)
			fmt.Println("dst:", dstIP)
			fmt.Println("dst:", dstPort)

			hepPacket = Packet{
				Version:   3,
				Protocol:  17,
				SrcIP:     net.ParseIP(srcIP),
				DstIP:     net.ParseIP(dstIP),
				SrcPort:   srcPort,
				DstPort:   dstPort,
				Tsec:      uint32(date.Unix()),
				Tmsec:     uint32(date.UnixMicro() - (date.Unix() * 1000)),
				ProtoType: 1,
			}

			if protoString == "UDP" {
				hepPacket.Protocol = 17
			} else if protoString == "TCP" {
				hepPacket.Protocol = 4
			}

			firtLine = true
			startMessage = false

			sipMessage = ""
		} else if firtLine && lineData == "" {
			startMessage = true
		} else if firtLine && startMessage {
			if replaceCallid && strings.HasPrefix(lineData, "Call-ID: ") {
				lineData = "Call-ID: " + generatedCallID
			}
			sipMessage = sipMessage + lineData + "\r\n"
		}

	}

	fmt.Println("Proto:", protoString)
	fmt.Println("dateString:", dateString)
	fmt.Println("date:", date.String())
	fmt.Println("src:", srcIP)
	fmt.Println("src:", srcPort)
	fmt.Println("dst:", dstIP)
	fmt.Println("dst:", dstPort)

	hepPacket = Packet{
		Version:   3,
		Protocol:  17,
		SrcIP:     net.ParseIP(srcIP),
		DstIP:     net.ParseIP(dstIP),
		SrcPort:   srcPort,
		DstPort:   dstPort,
		Tsec:      uint32(date.Unix()),
		Tmsec:     uint32(date.UnixMicro() - (date.Unix() * 1000)),
		ProtoType: 1,
		Payload:   []byte(sipMessage),
	}

	if protoString == "UDP" {
		hepPacket.Protocol = 17
	} else if protoString == "TCP" {
		hepPacket.Protocol = 4
	}

	fmt.Println("END:", sipMessage)

	scenario = append(scenario, hepPacket)

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	return scenario
}
