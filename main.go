package main

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
	"io"
)

// Global variable 
var (
	totalSize int64
	written int64
	ErrShortWrite error
	EOF = errors.New("EOF")
)

// Error handler
func Panic(msg string, err error) {
	if err != nil {
		fmt.Println(msg, err)
	}
}

// All download logic comes here.
func download(url string, ch chan int64) {
	// splite URL
	tokens := strings.Split(url, "/")
	// get file name from array
	fileName := tokens[len(tokens)-1]
	fmt.Println(fileName)

  // validate URL 
  // TODO - CAN except multiple format
	if !strings.Contains(url, ".zip") {
		fmt.Println("URL", url, "don't have expected extention")
		os.Exit(0)
	}

  // create file for download
  // default will be in root dir
	downlaod, err := os.Create(fileName)
	if err != nil {
		fmt.Println("Error while creating", fileName, "-", err)
		return
	}
	// cleanup-resource in case panic 
	defer downlaod.Close()

	response, err := http.Get(url)
	defer response.Body.Close()

	totalSize = response.ContentLength

  // Read file data 
	buf := make([]byte, 32*1024)
	for i := 0; ; i++ {
	  // TODO - sleep is not right way. It should be based on INTERNET speed and remaining amount of file size 
		time.Sleep(2 * 1e9)
		nr, er := response.Body.Read(buf)
		if nr > 0 {
			nw, ew := downlaod.Write(buf[0:nr])
			if nw > 0 {
				written += int64(nw)
			}
			if ew != nil {
				err = ew
				break
			}
			if nr != nw {
				err = ErrShortWrite
				break
			}
		}

    // calculate completed file size
		ch <- (written * 100) / totalSize

    // closed channel if file download complete and exit from loop.
		if er == io.EOF {
			close(ch)
			break
		}

    // Exception handle while downloading
		if er != nil {
			fmt.Println("Error while downloading", url, "-", er)
			err = er
			break
		}
	}
}

// Collect download status using channel
func downloadStatus(ch chan int64) {
	for {
	  // Check wether channel have data OR not.
		v, ok := <-ch
		if !ok {
			break
		}
		fmt.Println(v, "%")
	}
}

func main() {
	// make channel
	// TODO - channel should be dynamic as per requested URL, it should create on fly
	ch := make(chan int64)

  // Capture argument
	args := os.Args
	
	// Argument validation
	if len(os.Args) == 1 {
		fmt.Println("Usage: gowget url1 url2 url ...")
		return
	}

  // TODO - this code should be enough smart to handle multiple argument using goroutine. 
  // TODO - To handle multiple URL, multiple channel should be use. 
  // TODO - This code should highly designed to use MAXPROCESS to utilize all process.  
	for i, value := range args {
		if i != 0 {
		  // Used goroutine for asynchronous call
			go download(value, ch)
		}

	}

  // use goroutine for asynchronous result
	downloadStatus(ch)
}
