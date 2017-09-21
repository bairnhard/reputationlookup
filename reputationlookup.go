package main

import (
	"log"
	"net"
	"net/http"
	"strings"
	"time"

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

	}
	return ""
}

func getreputation(ip *gin.Context) {

	ipa, _ := ip.GetQuery("ips")
	ipn, _ := ip.GetQuery("net")

	if ipa != "" { // we have an IP address request
		hostarray := strings.Split(ipa, ",")
		var c []represult
		c = make([]represult, len(hostarray))

		for i, v := range hostarray {

			repcode, err := replookup(hostarray[i])
			if err != nil {
				repcode = "host not found"
			}

			c[i] = represult{v, repcode}
		}

		// ipa2 := net.ParseIP(ipa)
		// if ipa2 == nil {
		//			ip.JSON(500, "Invalid Address")
		//			log.Println(time.Now(), "Invalid Address: ", ipa)
		//			return
		//		}

		//		repcode, err := replookup(ipa)
		//		if err != nil {
		//			log.Println(" Lookup error: ", err)

		//		}
		//		c := represult{ipa, repcode}

		//fmt.Println("repcode", c.REP)

		ip.IndentedJSON(200, c)

	} else if ipn != "" { //we have a Net Request

		hostarray, err := hosts(ipn) // convert CIDR Network into array of hosts
		if err != nil {
			ip.JSON(500, "Invalid CIDR Request")
			log.Println(time.Now(), "Invalid CIDR Request: ", ipn)

		}
		var c []represult
		c = make([]represult, len(hostarray))

		for i, v := range hostarray {

			repcode, err := replookup(hostarray[i])
			if err != nil {
				repcode = "host not found"
			}

			c[i] = represult{v, repcode}
		}

		ip.IndentedJSON(200, c)

	} else {
		// log.Fatalln("Query Error", ip)
		ip.JSON(500, "Invalid Request")
		log.Println(time.Now(), "Invalid Request: ", ip)
	}

}

func replookup(ipa string) (string, error) { // lookup ip reputation for given IP address

	ipg := net.ParseIP(ipa)        //net.IP
	repip := reverseIPAddress(ipg) //reverse ip

	lookuptarget := repip + ".score.senderscore.com"

	reputation, err := net.LookupHost(lookuptarget)
	if err != nil {
		return "", err //host not found...
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

func usage(c *gin.Context) {

	c.HTML(http.StatusOK, "index.html", nil)

}

func main() {

	router := gin.Default()
	router.LoadHTMLGlob("*.html")

	router.GET("/", usage)
	router.GET("/reputation/", getreputation)
	log.Println(time.Now(), "reputation lookup started")
	router.Run()

}
