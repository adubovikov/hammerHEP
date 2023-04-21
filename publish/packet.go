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

func randomIPFromRange(cidr string) (net.IP, error) {

GENERATE:

	ip, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, err
	}

	// The number of leading 1s in the mask
	ones, _ := ipnet.Mask.Size()
	quotient := ones / 8
	remainder := ones % 8

	// create random 4-byte byte slice
	r := make([]byte, 4)
	rand.Read(r)

	for i := 0; i <= quotient; i++ {
		if i == quotient {
			shifted := byte(r[i]) >> remainder
			r[i] = ^ipnet.IP[i] & shifted
		} else {
			r[i] = ipnet.IP[i]
		}
	}
	ip = net.IPv4(r[0], r[1], r[2], r[3])

	if ip.Equal(ipnet.IP) /*|| ip.Equal(broadcast) */ {
		// we got unlucky. The host portion of our ipv4 address was
		// either all 0s (the network address) or all 1s (the broadcast address)
		goto GENERATE
	}
	return ip, nil
}

func GeneratePacketsArrayFromText(fileName string, replace ReplaceParams) []Packet {
	var scenario []Packet

	file, err := os.Open(fileName)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	//proto:UDP 2023-04-17T14:56:19.323629+02:00  10.0.0.2:5060 ---> 192.168.178.20:61000

	scanner := bufio.NewScanner(file)
	firtLine := false
	startMessage := false
	sipMessage := ""
	protoString := ""
	var error error
	srcIP := ""
	dstIP := ""
	var srcPort uint16
	var dstPort uint16
	generatedCallID := RandStringBytes(30)
	dateNow := time.Now()
	var date time.Time

	mapIP := make(map[string]string)

	var hepPacket Packet

	for scanner.Scan() {
		lineData := scanner.Text()

		if strings.HasPrefix(lineData, "proto:") {
			lineArray := strings.Split(lineData, " ")
			protoString = strings.TrimPrefix(lineArray[0], "proto:")
			dateArray := strings.Split(lineArray[1], "+")
			hepPacket.dateString = dateArray[0]

			if sipMessage != "" {
				hepPacket.Payload = []byte(sipMessage)
				scenario = append(scenario, hepPacket)
				//Debug
				//dumpHEPMessage(hepPacket)
			}

			date, error = time.Parse("2006-01-02T15:04:05.999999", hepPacket.dateString)

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

			//replace IP
			if replace.ReplaceIP {
				if val, ok := mapIP[srcIP]; ok {
					srcIP = val
				} else {
					randomIP, _ := randomIPFromRange("10.0.0.0/8")
					mapIP[srcIP] = randomIP.String()
				}

				if val, ok := mapIP[dstIP]; ok {
					dstIP = val
				} else {
					randomIP, _ := randomIPFromRange("192.168.0.0/8")
					mapIP[dstIP] = randomIP.String()
				}
			}

			hepPacket = Packet{
				Version:   0x02,
				Protocol:  17,
				SrcIP:     net.ParseIP(srcIP).To4(),
				DstIP:     net.ParseIP(dstIP).To4(),
				SrcPort:   srcPort,
				DstPort:   dstPort,
				Tsec:      uint32(date.Unix()),
				Tmsec:     uint32(date.UnixMilli() - (date.Unix() * 1000)),
				ProtoType: 1,
			}

			if protoString == "UDP" {
				hepPacket.Protocol = 17
			} else if protoString == "TCP" {
				hepPacket.Protocol = 4
			}

			firtLine = true

			//Clean up
			startMessage = false
			sipMessage = ""

		} else if firtLine && lineData == "" {
			startMessage = true
			firtLine = false
		} else if startMessage {
			if replace.ReplaceCid && strings.HasPrefix(lineData, "Call-ID: ") {
				lineData = "Call-ID: " + generatedCallID
			}
			sipMessage = sipMessage + lineData + "\r\n"
		}
	}

	hepPacket = Packet{
		Version:   0x02,
		Protocol:  17,
		SrcIP:     net.ParseIP(srcIP).To4(),
		DstIP:     net.ParseIP(dstIP).To4(),
		SrcPort:   srcPort,
		DstPort:   dstPort,
		Tsec:      uint32(date.Unix()),
		Tmsec:     uint32(date.UnixMilli() / (date.Unix() * 1000)),
		ProtoType: 1,
		Payload:   []byte(sipMessage),
	}

	difference := dateNow.Sub(date)

	//fmt.Println("DIFF - ", difference.Seconds())

	if protoString == "UDP" {
		hepPacket.Protocol = 17
	} else if protoString == "TCP" {
		hepPacket.Protocol = 4
	}

	scenario = append(scenario, hepPacket)

	//Range
	if replace.ReplaceTime {
		for i := range scenario {
			//fmt.Println("BEFORE DIFF - ", scenario[i].Tsec)
			scenario[i].Tsec += uint32(difference.Seconds())
			scenario[i].Tmsec = uint32(i * 100)
			//fmt.Println("AFTER DIFF - ", scenario[i].Tsec)
		}
	}

	//Debug
	for _, s := range scenario {
		dumpHEPMessage(s)
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	return scenario
}

func dumpHEPMessage(hep Packet) {

	fmt.Println("Date orig:", hep.dateString)
	fmt.Println("Tsec:", hep.Tsec)
	fmt.Println("Tmsec:", hep.Tmsec)
	fmt.Println("src IP:", hep.SrcIP.String())
	fmt.Println("src Port:", hep.SrcPort)
	fmt.Println("dst IP:", hep.DstIP.String())
	fmt.Println("dst Port:", hep.DstPort)

	fmt.Println("Message:", string(hep.Payload))
}

func dumpHEPMessages(heps []Packet) {

	for _, h := range heps {
		dumpHEPMessage(h)
	}

}
