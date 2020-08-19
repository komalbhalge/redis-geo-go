package main

import (
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func main() {

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
