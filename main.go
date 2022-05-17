package main

import (
	"fmt"
	"crypto/tls"
	"github.com/go-resty/resty/v2"
	"strings"
	"regexp"
	"time"
	"bufio"
	"sync"
	"os"
	"flag"
	"runtime"
	//"strconv"
)
var (
	title = `<title>([\s\S]+?)</title>`
	conf = int(0)// 配置、终端默认设置
	bg = int(0)  // 背景色、终端默认设置
	green = int(32)//绿色
	//read = int(31)

)
type Info struct{
	Code int
	Title string
	Url string
	Bodylength int 
}
func get(urlchan chan string,wg *sync.WaitGroup){
	defer wg.Done()
	for url := range urlchan{
		var info Info
		if(!strings.Contains(url,"http")){
			url = "http://"+url  //默认为http
		}	
		client := resty.New().SetTimeout(time.Duration(2 * time.Second)).SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
		client.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/69.0.3497.100 Safari/537.36")
		resp,err := client.R().EnableTrace().Get(url)
		if err != nil{
			//fmt.Println(err)
			continue
		}
		if (strings.Contains(string(resp.Body()),"HTTP request was sent to HTTPS")){
			url = strings.Replace(url,"http","https",-1)
			resp,err = client.R().EnableTrace().Get(url)
			if err != nil{
				//fmt.Println(err)
				continue
			}
		}
		info.Code = resp.StatusCode()
		str := resp.Body()
		body := string(str)
		if strings.Contains(body,"<title>"){
			re := regexp.MustCompile(title)
			title_name := re.FindAllStringSubmatch(body,1)
			if len(title_name) == 0{
				continue
			}
			info.Title = title_name[0][1]
		}
		info.Url = url
		info.Bodylength = len(body)
		fmt.Printf("%c[%d;%d;%dm%s%c[0m", 0x1B, conf, bg, green, "[+]", 0x1B)
		fmt.Printf(info.Url+" "+info.Title+" %d %d\n",info.Code,info.Bodylength)
	}

}
func read(path string,urlchan chan string){
	defer close(urlchan)
	file,err := os.Open(path)
	if err != nil {
		fmt.Printf("failed %s",err.Error())
		os.Exit(0)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan(){
		urlchan <- strings.TrimSpace(scanner.Text())
	}
	
}
func main(){
	var wg sync.WaitGroup
	var urlchan = make(chan string,1)
	var path string
	var threads int
	flag.StringVar(&path,"p","","the path of the targets")
	flag.IntVar(&threads,"t",runtime.NumCPU(),"the threads of the program")
	flag.Parse()
	if path == "" {
		fmt.Printf("please input the path of the targets,-h for help")
		return
	}
	//path := "url.txt"
	for i:=0;i < threads; i++{
		wg.Add(1)
		go get(urlchan,&wg)
	}
	go read(path,urlchan)
	wg.Wait()
}