package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/go-redis/redis"
)

var ctx = context.Background()

func main3() {
	client := rClient()
	//Save simple key and value (String/int)
	setVal(client)

	//Save Map and Slice
	err := setMapAndSlice(client)
	if err != nil {
		log.Println("setMapAndSlice Error: ", err)
	}

	//Save file
	fmt.Println("File: ", saveFile(client))

}
func rClient() *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	pong, err := rdb.Ping(ctx).Result()
	fmt.Println(pong, err)
	return rdb
}

func setVal(client *redis.Client) {
	key := "name"
	value := "Komal Bhalge"
	err := client.Set(ctx, key, value, 0).Err()
	if err != nil {
		panic(err)
	}

	val, err := client.Get(ctx, key).Result()
	if err != nil {
		panic(err)
	}
	fmt.Println("Name: ", val)

}
func saveFile(client *redis.Client) string {

	content, err := ioutil.ReadFile("files/img2.png")
	key := "file"
	if err != nil {
		log.Fatal(err)
	}
	err = client.Set(ctx, key, string(content), 0).Err()
	if err != nil {
		panic(err)
	}

	val, err := client.Get(ctx, key).Result()
	if err != nil {
		panic(err)
	}
	permissions := 0644 // or whatever you need
	byteArray := []byte(val)
	err = ioutil.WriteFile("outImage1.png", byteArray, os.FileMode(permissions))
	if err != nil {
		fmt.Println("Err:", err) // handle error
	}

	return "File fetched!"
}

func setMapAndSlice(client *redis.Client) error {

	key := "Map"
	//Save map
	ip := map[string]int{"komal": 1, "Nelson": 2}

	// serialize User object to JSON
	str, err := json.Marshal(ip)
	if err != nil {
		log.Println(err)
		return err
	}
	//err := client.ser
	err = client.Set(ctx, key, str, 0).Err()
	if err != nil {
		log.Println(err)
	}

	out := make(map[string]int)
	value, err := client.Get(ctx, key).Result()
	if err != nil {
		log.Println(err)
	}
	err = json.Unmarshal([]byte(value), &out)

	fmt.Printf("Map: %+v\n", out)

	//Save Slice
	sl := []string{"g", "h", "i"}
	key = "slice"
	// serialize User object to JSON
	str, err = json.Marshal(sl)
	if err != nil {
		log.Println(err)
		return err
	}
	err = client.Set(ctx, key, str, 0).Err()
	if err != nil {
		log.Println(err)
	}

	var outVal []string
	value, err = client.Get(ctx, key).Result()
	if err != nil {
		log.Println(err)
	}
	err = json.Unmarshal([]byte(value), &outVal)

	fmt.Printf("Slice: %+v\n", outVal)
	return err
}
