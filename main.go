package main

import (
	"fmt"
	"io/ioutil"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func webServer(ctx *gin.Context) {
	ctx.HTML(200, "index.html", nil)
}

func whoisServer(ctx *gin.Context) {

	if strings.HasPrefix(ctx.Request.URL.Path, "/whois/RADB/") == true {
		Target := strings.ReplaceAll(ctx.Request.URL.Path, "/whois/RADB/", "")
		ctx.String(200, whoisRADB(Target), nil)
	} else if strings.HasPrefix(ctx.Request.URL.Path, "/whois/") == true {
		Target := "NULL"
		Target = strings.ReplaceAll(ctx.Request.URL.Path, "/whois/", "")
		result, err := Whois(Target)
		if err != nil {
			fmt.Println(result)
		}
		ctx.String(200, result, nil)
	}

}

func whoisRADB(Target string) string {
	result, err := Whois(Target, "whois.radb.net")
	if err != nil {
		fmt.Println(result)
	}
	return result
}

func whoisPOST(ctx *gin.Context) {
	Target := ctx.PostForm("target")
	result, err := Whois(Target)
	if err != nil {
		fmt.Println(result)
	}
	ctx.String(200, result, nil)
}

func whoisRADBPOST(ctx *gin.Context) {
	Target := ctx.PostForm("target")
	result, err := Whois(Target, "whois.radb.net")
	if err != nil {
		fmt.Println(result)
	}
	ctx.String(200, result, nil)
}

func pageNotAvailable(ctx *gin.Context) {
	ctx.HTML(404, "404.html", nil)
}

const (
	// IANAWHOISSERVER is iana whois server
	IANAWHOISSERVER = "whois.iana.org"
	// DEFAULTWHOISPORT is default whois port
	DEFAULTWHOISPORT = "43"
)

// Whois do the whois query and returns whois info
func Whois(domain string, servers ...string) (result string, err error) {
	domain = strings.Trim(strings.TrimSpace(domain), ".")
	if domain == "" {
		return "", fmt.Errorf("whois: domain is empty")
	}

	if !strings.Contains(domain, "as") && !strings.Contains(domain, "AS") {
		if !strings.Contains(domain, ".") && !strings.Contains(domain, ":") {
			checkASN := strings.Replace(strings.ToUpper(domain), "AS", "", -1)
			_, err := strconv.ParseFloat(checkASN, 64)
			if err != nil {
				return query(domain, IANAWHOISSERVER)
			}
		}
	}

	var server string
	if len(servers) == 0 || servers[0] == "" {
		ext := getExtension(domain)
		result, err := query(ext, IANAWHOISSERVER)
		if err != nil {
			return "", fmt.Errorf("whois: query for whois server failed: %v", err)
		}
		server = getServer(result)
		if server == "" {
			return "", fmt.Errorf("whois: no whois server found for domain: %s", domain)
		}
	} else {
		server = strings.ToLower(servers[0])
	}

	result, err = query(domain, server)
	if err != nil {
		return
	}

	refServer := getServer(result)
	if refServer == "" || refServer == server {
		return
	}

	data, err := query(domain, refServer)
	if err == nil {
		result += data
	}

	return
}

// query send query to server
func query(domain, server string) (string, error) {
	if server == "whois.arin.net" {
		domain = "n + " + domain
	}

	conn, err := net.DialTimeout("tcp", net.JoinHostPort(server, DEFAULTWHOISPORT), time.Second*30)
	if err != nil {
		return "", fmt.Errorf("whois: connect to whois server failed: %v", err)
	}

	defer conn.Close()
	_ = conn.SetWriteDeadline(time.Now().Add(time.Second * 30))
	_, err = conn.Write([]byte(domain + "\r\n"))
	if err != nil {
		return "", fmt.Errorf("whois: send to whois server failed: %v", err)
	}

	_ = conn.SetReadDeadline(time.Now().Add(time.Second * 30))
	buffer, err := ioutil.ReadAll(conn)
	if err != nil {
		return "", fmt.Errorf("whois: read from whois server failed: %v", err)
	}

	return string(buffer), nil
}

// getExtension returns extension of domain
func getExtension(domain string) string {
	ext := domain

	if net.ParseIP(domain) == nil {
		domains := strings.Split(domain, ".")
		ext = domains[len(domains)-1]
	}

	if strings.Contains(ext, "/") {
		ext = strings.Split(ext, "/")[0]
	}

	return ext
}

// getServer returns server from whois data
func getServer(data string) string {
	tokens := []string{
		"Registrar WHOIS Server: ",
		"whois: ",
	}

	for _, token := range tokens {
		start := strings.Index(data, token)
		if start != -1 {
			start += len(token)
			end := strings.Index(data[start:], "\n")
			return strings.TrimSpace(data[start : start+end])
		}
	}

	return ""
}

func main() {

	fmt.Print("\n")
	fmt.Print("-------------------\n")
	fmt.Print("\n")
	fmt.Print("SteveYi Whois Service\n")
	fmt.Print("Port listing at 30010\n")
	fmt.Print("Repo: https://github.com/SteveYi-LAB/SteveYi-Whois\n")
	fmt.Print("Author: SteveYi\n")
	fmt.Print("Demo: https://whois.steveyi.net\n")
	fmt.Print("\n")
	fmt.Print("-------------------\n")
	fmt.Print("\n")

	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	router.LoadHTMLGlob("static/*")

	router.GET("/", webServer)
	router.GET("/whois/:target", whoisServer)
	router.POST("/whois/", whoisPOST)

	router.NoRoute(pageNotAvailable)

	router.Run(":30010")
}
