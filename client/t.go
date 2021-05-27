package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

func mainn() {
	for _, file := range os.Args {
		f, err := ioutil.ReadFile(file)
		if err != nil {
			continue
		}
		fmt.Println("file: " + file)
		fmt.Printf("size: %d\n", len(f))
		fmt.Println(f[len(f)-10:])
		fmt.Println()
	}
}

func main() {
	if len(os.Args) > 0 {
		rsp, err := http.Get("http://localhost:8080/oi")
		if err != nil {
			log.Fatal(err)
		}

		b, _ := ioutil.ReadAll(rsp.Body)
		log.Println(len(b))
		ioutil.WriteFile("resultado.xlsx", b, 0766)
	}

}
