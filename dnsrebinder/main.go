package main

import (
	"dnsrebinder/capture"
	"dnsrebinder/dns"
	"dnsrebinder/http"
	"flag"
	"fmt"
	"github.com/spf13/viper"
	"os"
	"os/signal"
)

var Config = struct {
	EnableHTTP    bool
	EnableDNS     bool
	EnableCapture bool

	HTTPBind string
	WebRoot  string

	DNSPort string
	DNSName string
	HTTPIP  string

	CaptureBind string
	CaptureRoot string
	CaptureIP   string
}{}
var HTTPState map[string]int = make(map[string]int)

func main() {
	flag.Parse()
	viper.SetConfigName("dnsrebind")
	viper.AddConfigPath("/etc/dnsrebind/")

	err := viper.ReadInConfig()
	if err != nil {
		fmt.Printf("Failed to load configuration file! %v\n", err)
	}

	viper.SetDefault("EnableHttp", false)
	viper.SetDefault("EnableDNS", false)
	viper.SetDefault("EnableCapture", false)
	viper.SetDefault("HTTPBind", "0.0.0.0:80")
	viper.SetDefault("CaptureBind", "0.0.0.0:10000")
	viper.SetDefault("WebRoot", "/var/www")
	viper.SetDefault("CaptureRoot", "/var/www/capture")
	viper.SetDefault("DNSPort", "8053")
	err = viper.Unmarshal(&Config)

	if err != nil {
		fmt.Printf("Failed to unmarshal config! %v\n", err)
	}

	fmt.Printf("Config: %#v\n", Config)

	if Config.EnableHTTP {
		http.WebRoot = Config.WebRoot
		http.StartHTTPServer(Config.HTTPBind, &Config)
		fmt.Println("Started HTTP Server...")
	}
	if Config.EnableDNS {
		dns.HTTPState = HTTPState
		dns.StartDNS(Config.DNSName, Config.DNSPort, Config.HTTPIP, Config.CaptureIP)
		fmt.Println("Started DNS Server...")
	}
	if Config.EnableCapture {
		capture.CaptureRoot = Config.CaptureRoot
		capture.HTTPState = HTTPState
		capture.StartHTTPCaptureServer(Config.CaptureBind)
		fmt.Println("Started HTTP Capture Server...")
	}
	signalChan := make(chan os.Signal, 1)
	cleanupDone := make(chan bool)
	signal.Notify(signalChan, os.Interrupt)
	go func() {
		for _ = range signalChan {
			fmt.Println("\nReceived an interrupt, stopping services...\n")
			cleanupDone <- true
		}
	}()
	<-cleanupDone
}
