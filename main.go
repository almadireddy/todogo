package main

import (
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

var todoItemStore TodoItemStore
var userStore UserStore
var jwtKey = []byte("MyKey")

func main() {
	router := mux.NewRouter()
	err := todoItemStore.InitStore()
	if err != nil {
		log.Fatal(err.Error())
	}
	err = userStore.InitStore()
	if err != nil {
		log.Fatal(err.Error())
	}
	router.HandleFunc("/items", ListItems).Methods("GET")
	router.HandleFunc("/items", AddItem).Methods("POST")
	router.HandleFunc("/signup", SignUp).Methods("POST")
	router.HandleFunc("/signin", SignIn).Methods("POST")
	router.HandleFunc("/refresh", Refresh).Methods("GET")
	log.Println("Listening for requests")
	log.Fatal(http.ListenAndServe(":12345", router))
}