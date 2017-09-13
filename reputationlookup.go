package main

//test
import (
	"fmt"
	"log"
	"net"
	"strings"

	"github.com/gin-gonic/gin"
)

type represult struct { // here we'll fill in the results
	ADDR string `json:"ip"`
	REP  string `json:"reputation"`
}

func reverseIPAddress(ip net.IP) string {

	if ip.To4() != nil {
		// split into slice by dot .
		addressSlice := strings.Split(ip.String(), ".")
		reverseSlice := []string{}

		for i := range addressSlice {
			octet := addressSlice[len(addressSlice)-1-i]
			reverseSlice = append(reverseSlice, octet)
		}

		// sanity check
		//fmt.Println(reverseSlice)

		return strings.Join(reverseSlice, ".")

	} else {
		panic("invalid IPv4 address")
	}
}

func getreputation(ip *gin.Context) {

	ipa, _ := ip.GetQuery("IP")
	ipn, _ := ip.GetQuery("NET")

	if ipa != "" { // we have an IP address request

		repcode, err := replookup(ipa)
		if err != nil {
			log.Fatalln(" Lookup error: ", err)

		}
		c := represult{ipa, repcode}

		fmt.Println("repcode", c.REP)

		ip.IndentedJSON(200, c)

	} else if ipn != "" { //we have a Net Request

		hostarray, err := hosts(ipn) // convert CIDR Network into array of hosts
		if err != nil {
			log.Fatalln("CIDR error: ", err)

		}
		var c []represult
		c = make([]represult, len(hostarray))

		for i, v := range hostarray {

			repcode, err := replookup(hostarray[i])
			if err != nil {
				// log.Fatalln(" Lookup error: ", err)
				repcode = "host not found"
			}

			c[i] = represult{v, repcode}
		}

		ip.IndentedJSON(200, c)

	} else {
		log.Fatalln("Query Error", ip)
	}

}

func replookup(ipa string) (string, error) { // lookup ip reputation for given IP address

	ipg := net.ParseIP(ipa)        //net.IP
	repip := reverseIPAddress(ipg) //reverse ip

	lookuptarget := repip + ".score.senderscore.com"

	reputation, err := net.LookupHost(lookuptarget)
	if err != nil {
		//log.Fatalln("host not found")
		return "", err
		//ip.Status(411)
	}
	addressSlice := strings.Split(reputation[0], ".")
	repcode := addressSlice[3]
	return repcode, nil

}

func hosts(cidr string) ([]string, error) { //converts CIDR String to array of hosts
	ip, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, err
	}

	var ips []string
	for ip := ip.Mask(ipnet.Mask); ipnet.Contains(ip); inc(ip) {
		ips = append(ips, ip.String())
	}
	// remove network address and broadcast address
	return ips[1 : len(ips)-1], nil
}

func inc(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

func usage() {} // start page

func main() {

	router := gin.Default()

	// Usage:

	// http://localhost:8080/getreputation/123.123.123.123 returns sender score reputation status

	router.GET("/getreputation/", getreputation)
	router.GET("/", usage)

	router.Run()

}
