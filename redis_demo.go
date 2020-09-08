package main

import (
	"encoding/json"
	"fmt"

	"github.com/gomodule/redigo/redis"
)

//User hold all details of a person
type User struct {
	Username  string
	MobileID  string
	Email     string
	FirstName string
	LastName  string
}

func initRedis() {
	fmt.Println("Redis demo started!")
	pool := newPool()
	conn := pool.Get()
	defer conn.Close()
	err := ping(conn)
	if err != nil {
		fmt.Println(err)
	}
	setAndGetValues(conn)
}
func setAndGetValues(conn redis.Conn) {
	//set a string
	setValues(conn, "name", "Komal Bhalge")

	value, err := getValues(conn, "name")
	if err != nil {
		fmt.Println("err", err)
	}
	fmt.Println("GET Values from redis: ", value)

	usr := User{Username: "komalbhalge", MobileID: "090807060", Email: "komal.bhalge@yahoo.com", FirstName: "Komal", LastName: "Bhalge"}

	setStruct(conn, usr)
	valueStrct, err := getStruct(conn, "komalbhalge")
	if err != nil {
		fmt.Println("err", err)
	}
	fmt.Println("GET Email from redis: ", valueStrct.Email)
}
func newPool() *redis.Pool {
	return &redis.Pool{

		// Maximum number of idle connections in the pool.
		MaxIdle: 50,
		// max number of connections
		MaxActive: 1000,
		// Dial is an application supplied function for creating and
		// configuring a connection.
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", ":6379")
			if err != nil {
				panic(err.Error())
			}
			return c, err
		},
	}
}

// ping tests connectivity for redis (PONG should be returned)
func ping(c redis.Conn) error {
	// Send PING command to Redis
	pong, err := c.Do("PING")
	if err != nil {
		return err
	}

	// PING command returns a Redis String
	s, err := redis.String(pong, err)
	if err != nil {
		return err
	}

	fmt.Printf("PING Response = %s\n", s)

	return nil
}

// set executes the redis SET command
func setValues(c redis.Conn, key string, value string) error {
	_, err := c.Do("SET", key, value)
	if err != nil {
		return err
	}

	return nil
}

// get executes the redis GET command
func getValues(c redis.Conn, key string) (string, error) {

	// Simple GET example with String helper
	s, err := redis.String(c.Do("GET", key))
	if err == redis.ErrNil {
		fmt.Printf("%s does not exist\n", key)
	} else if err != nil {
		return "", err
	} else {
		fmt.Printf("%s = %s\n", key, s)
	}
	return s, nil
}

func setStruct(c redis.Conn, usr User) error {

	// serialize User object to JSON
	json, err := json.Marshal(usr)
	if err != nil {
		return err
	}
	// SET object
	_, err = c.Do("SET", usr.Username, json)
	if err != nil {
		return err
	}

	return nil
}
func getStruct(c redis.Conn, username string) (User, error) {

	usr := User{}
	s, err := redis.String(c.Do("GET", username))
	if err == redis.ErrNil {
		fmt.Println("User does not exist")
	} else if err != nil {
		return usr, err
	}

	err = json.Unmarshal([]byte(s), &usr)

	fmt.Printf("%+v\n", usr)

	return usr, nil
}
