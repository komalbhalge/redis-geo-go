package main

import (
	"log"
	"net/http"
	"sync"

	"github.com/go-redis/redis/v8"
	"github.com/julienschmidt/httprouter"
	"golang.org/x/exp/errors/fmt"
)

type RedisClient struct{ *redis.Client }

const key = "users"

/*sync.Once has an atomic counter and it uses atomic.StoreUint32 to set a value to 1, when the function has been called,
and then atomic.LoadUint32 to see if it needs to be called again.
For this basic implementation GetRedisClient will be called from two endpoints but we only want to get one instance. */
var once sync.Once
var redisClient *RedisClient

func GetRedisClient() *RedisClient {
	once.Do(func() {
		client := redis.NewClient(&redis.Options{
			Addr:     "localhost:6379",
			Password: "", // no password set
			DB:       0,  // use default DB
		})
		redisClient = &RedisClient{client}
	})
	_, err := redisClient.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Could not connect to redis %v", err)
	}
	return redisClient
}

//AddUserLocation adds locations
func (c *RedisClient) AddUserLocation(lng, lat float64, id int) {

	c.GeoAdd(ctx,
		key,
		&redis.GeoLocation{Longitude: lng, Latitude: lat, Name: string(id)},
	)
}

func (c *RedisClient) RemoveUserLocation(id string) {
	c.ZRem(ctx, key, id)
}

func (c *RedisClient) SearchUsers(req SearchReqBody) []redis.GeoLocation {
	/*
		WITHDIST: Also return the distance of the returned items from    the specified center. The distance is returned in the same unit as the unit specified as the radius argument of the command.

		WITHCOORD: Also return the longitude,latitude coordinates of the  matching items.

		WITHHASH: Also return the raw geohash-encoded sorted set score of the item, in the form of a 52 bit unsigned integer. This is only useful for low level hacks or debugging and is otherwise of little interest for the general user.
	*/
	var limit = req.Limit
	var radius = req.Radius

	if limit == 0 {
		limit = 40
	}

	if radius == 0 {
		radius = 4.0
	}
	fmt.Println("Limit:", req.Limit)
	fmt.Println("Raduis:", req.Radius)

	res, _ := c.GeoRadius(ctx, key, req.Lng, req.Lat, &redis.GeoRadiusQuery{
		Radius:      radius,
		Unit:        "km",
		WithGeoHash: true,
		WithCoord:   true,
		WithDist:    true,
		Count:       limit,
		Sort:        "ASC",
	}).Result()
	return res
}
func initRedisGeo() {
	router := setupRoutes()
	log.Println("Server started to accept request in 8080 port.")
	http.ListenAndServe(":8080", router)
}
func setupRoutes() *httprouter.Router {
	router := httprouter.New()
	router.POST("/addlocation", AddLocation)
	router.POST("/searchLocation", SearchLocation)
	return router
}
