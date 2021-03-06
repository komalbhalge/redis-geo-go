package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/go-redis/redis/v8"
	"github.com/julienschmidt/httprouter"
)

//UserLocation holds user data along with location params
type UserLocation struct {
	ID           string       `json:"id"`
	Lat          float64      `json:"lat"`
	Lng          float64      `json:"lng"`
	LocationType LocationType `json:"locationtype"`
}

//SerachBody is request params for a search query
type SearchReqBody struct {
	Lat          float64      `json:"lat"`
	Lng          float64      `json:"lng"`
	LocationType LocationType `json:"locationtype"`
	Limit        int          `json:"limit,omitempty"`
	Radius       float64      `json:"radius,omitempty"`
}

var ctx = context.Background()

//LocationType use to identify type of location
type LocationType int

const (
	//ATM type= 1
	ATM LocationType = 1 + iota
	//BANK type= 2
	BANK
	//MONEYCHANGER type= 3
	MONEYCHANGER
)

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

//AddLocation adds location to redis
func AddLocation(res http.ResponseWriter, req *http.Request, ps httprouter.Params) {
	fmt.Println("Adding location...")

	var user UserLocation
	rClient := GetRedisClient()
	if err := json.NewDecoder(req.Body).Decode(&user); err != nil {
		log.Printf("could not decode request: %v", err)
		http.Error(res, "could not decode request", http.StatusInternalServerError)
		return
	}

	if isValidLocationType(user.LocationType) {
		fmt.Println("Location added!")

		// Add new location
		// You can save locations in another db
		rClient.AddUserLocation(user.Lng, user.Lat, int(user.LocationType))
		res.WriteHeader(http.StatusOK)
	} else {
		res.WriteHeader(422) //422 for Unprocessable Entity or wrong input
	}

	return
}

// SearchLocation receives lat and lng of the picking point and searches users about this point.
func SearchLocation(res http.ResponseWriter, req *http.Request, ps httprouter.Params) {
	fmt.Println("AddLocation...")

	var reqBody SearchReqBody

	if err := json.NewDecoder(req.Body).Decode(&reqBody); err != nil {
		log.Printf("could not decode request: %v", err)
		http.Error(res, "could not decode request", http.StatusInternalServerError)
		return
	}

	rClient := GetRedisClient()

	users := rClient.SearchUsers(reqBody)
	data, err := json.Marshal(users)
	if err != nil {
		fmt.Println("Error: ", err.Error())
		return
	}
	fmt.Println("All Users:", string(data))
	return
}

func isValidLocationType(locationtype LocationType) bool {
	var types = []LocationType{
		ATM,
		BANK,
		MONEYCHANGER,
	}
	for _, t := range types {
		if t == locationtype {
			return true
		}
	}
	return false
}
