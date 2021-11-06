package main

import (
	"bufio"
	"crypto/tls"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"sync"
)

type headers []string

type options struct {
	Extensions          string
	InputFile           string
	Cookies             string
	Proxy               string
	Threads             int
	NotCheckCert        bool
	Headers             headers
	StatusCodeBlacklist string
	OutputFile          string
}

type request struct {
	RequestURL string
	Request    *http.Request
	Client     *http.Client
}

type response struct {
	Response   *http.Response
	RequestURL string
}

var o options

var colorReset string = "\033[0m"
var colorRed string = "\033[31m"
var colorGreen string = "\033[32m"

func colorString(color string, output string) string {
	return color + output + colorReset
}

func isError(e error) bool {
	if e != nil {
		fmt.Println(e.Error())
		return true
	}
	return false
}

func prepareClient() *http.Client {

	var proxyClient func(*http.Request) (*url.URL, error)
	if o.Proxy == "" {
		proxyClient = http.ProxyFromEnvironment
	} else {
		tmp, _ := url.Parse(o.Proxy)
		proxyClient = http.ProxyURL(tmp)
	}

	transport := &http.Transport{
		Proxy:               proxyClient,
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 100,
		TLSClientConfig:     &tls.Config{InsecureSkipVerify: o.NotCheckCert},
	}

	client := &http.Client{
		Transport: transport,
	}

	return client
}

func prepareRequest(requestURL string, host string) *http.Request {

	req, err := http.NewRequest("GET", requestURL, nil)
	if isError(err) {
		os.Exit(1)
	}

	if o.Cookies != "" {
		req.Header.Set("Cookie", o.Cookies)
	}

	if host != "" {
		req.Host = host
	}

	for _, header := range o.Headers {
		h := strings.Split(header, ":")
		req.Header.Set(h[0], h[1])
	}

	return req
}

func (i *headers) String() string {
	var rep string
	for _, e := range *i {
		rep += e
	}
	return rep
}

func (i *headers) Set(value string) error {
	*i = append(*i, value)
	return nil
}

func setHostHeaderIfExists() (host string) {
	for _, header := range o.Headers {
		h := strings.Split(header, ":")
		if len(h) != 2 {
			fmt.Printf("Error in headers declaration: %s\n", header)
		}
		if h[0] == "Host" {
			host = h[1]
		}
	}
	return
}

func getResponseFromURL(r request) *http.Response {
	response, err := r.Client.Do(r.Request)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	return response
}

func addExtensionToUrl(url string, ext string) string {
	splitted := strings.Split(url, "?")
	if len(splitted) > 1 {
		return splitted[0] + ext + "?" + splitted[1]
	}
	return url + ext
}

func printResponse(response *http.Response, url string, file *os.File) {
	result := fmt.Sprintf("%s -> %s", url, response.Status)
	if !strings.Contains(o.StatusCodeBlacklist, response.Status[0:2]) {
		if strings.Contains(response.Status, "200") {
			fmt.Println(colorString(colorGreen, result))
		} else {
			fmt.Println(result)
		}
		if file != nil {
			file.WriteString(url + "\n")
		}
	}
}

func checkStatusCodeBlacklist() bool {
	re := regexp.MustCompile("^[1-5][0-9][0-9]$")
	for _, element := range strings.Split(o.StatusCodeBlacklist, ",") {
		if !re.MatchString(element) {
			return false
		}
	}
	return true
}

func showHelper() {
	helper := []string{
		"brb by zblurx",
		"",
		"Bruteforcing URLs with specific extensions (like .bak, .bak1, ~, .old, etc.)",
		"",
		" -i, --input-file <path>\t\tSpecify filepath",
		" -x, --extensions-list <list>\t\tSpecify extensions in comma separated list. Default .bak,.bak1,~..old",
		" -H, --header <header>\t\t\tSpecify header. Can be used multiple times",
		" -c, --cookies <cookies>\t\tSpecify cookies",
		" -x, --proxy <proxy>\t\t\tSpecify proxy",
		" -k, --insecure\t\t\t\tAllow insecure server connections when using SSL",
		" -t, --threads <int>\t\t\tNumber of thread. Default 10",
		" -b, --status-code-blacklist <list>\tComme separated list of status code not to output",
	}

	fmt.Println(strings.Join(helper, "\n"))
}

func init() {
	flag.Usage = func() {
		showHelper()
	}
}

func main() {
	flag.StringVar(&o.Extensions, "extensions-list", ".bak,.bak1,~,.old", "")
	flag.StringVar(&o.Extensions, "x", ".bak,.bak1,~,.old", "")

	flag.StringVar(&o.InputFile, "input-file", "", "")
	flag.StringVar(&o.InputFile, "i", "", "")

	flag.StringVar(&o.Cookies, "cookies", "", "")
	flag.StringVar(&o.Cookies, "c", "", "")

	flag.StringVar(&o.Proxy, "proxy", "", "")
	flag.StringVar(&o.Proxy, "p", "", "")

	flag.IntVar(&o.Threads, "threads", 10, "")
	flag.IntVar(&o.Threads, "t", 10, "")

	flag.Var(&o.Headers, "header", "")
	flag.Var(&o.Headers, "H", "")

	flag.BoolVar(&o.NotCheckCert, "insecure", false, "")
	flag.BoolVar(&o.NotCheckCert, "k", false, "")

	flag.StringVar(&o.StatusCodeBlacklist, "status-code-blacklist", "404,302", "")
	flag.StringVar(&o.StatusCodeBlacklist, "b", "404,302", "")

	flag.StringVar(&o.OutputFile, "output-file", "", "")
	flag.StringVar(&o.OutputFile, "o", "", "")

	flag.Parse()

	var wg sync.WaitGroup
	var pg sync.WaitGroup

	requests := make(chan request)
	responses := make(chan response)

	if !checkStatusCodeBlacklist() {
		fmt.Println("Status Code Blacklist is not correct")
		os.Exit(1)
	}

	for i := 0; i < o.Threads; i++ {
		wg.Add(1)

		go func() {
			for req := range requests {
				responses <- response{
					RequestURL: req.RequestURL,
					Response:   getResponseFromURL(req),
				}
			}
			wg.Done()
		}()
	}

	pg.Add(1)
	go func() {
		var output_file *os.File
		if o.OutputFile != "" {
			var err error
			output_file, err = os.Create(o.OutputFile)
			if isError(err) {
				os.Exit(1)
			}

			defer output_file.Close()
		}
		for resp := range responses {
			printResponse(resp.Response, resp.RequestURL, output_file)
		}
		pg.Done()
	}()

	host := setHostHeaderIfExists()

	httpClient := prepareClient()

	inputFile, err := os.Open(o.InputFile)
	if isError(err) {
		os.Exit(1)
	}

	scanner := bufio.NewScanner(inputFile)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		for _, ext := range strings.Split(o.Extensions, ",") {
			url := addExtensionToUrl(scanner.Text(), ext)
			requests <- request{
				RequestURL: url,
				Client:     httpClient,
				Request:    prepareRequest(url, host),
			}
		}

	}

	inputFile.Close()
	close(requests)
	wg.Wait()
	close(responses)
	pg.Wait()
}
