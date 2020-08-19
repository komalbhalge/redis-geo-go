package main

import (
	"fmt"
	"io/ioutil"
	"log"
)

func saveMyFile() string {

	content, err := ioutil.ReadFile("/files/image1.png")

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Img1: ", string(content))
	return string(content)
}
