/**
		Payment API - Server Module
		Author: Luis Otero
		
		Holds a server creation object
		Initializes all routes for the API
**/
package main

import (
	"github.com/urfave/negroni"
	"github.com/gorilla/mux"
)

/** 
	Server Creation Function
	
	1. Creates server object
	2. Creates router and handler
	3. Attaches router to server
	4. Return server object
**/
func CreateServer() *negroni.Negroni {
	//create Classic server
	server := negroni.Classic()
	
	//create a router object
	router := mux.NewRouter()
	
	//initialize routes
	setRoutes(router)

	//apply route handler to server
	server.UseHandler(router)
	
	return server
}

/**
	Route Initializing Function
	
	1. Set all of the routes for the application
**/
func setRoutes(router *mux.Router) {
	router.HandleFunc("/transaction", AddTransactionToQueue()).Methods("POST")
	router.HandleFunc("/transactions/{id}", SearchForTransaction()).Methods("GET")
	router.HandleFunc("/transactions", SearchForTransaction()).Methods("GET")
	router.HandleFunc("/process", ProcessAllTransactions()).Methods("POST")
	router.HandleFunc("/update", UpdateTransaction()).Methods("PUT")
	router.HandleFunc("/delete/{id}", DeleteTransaction()).Methods("DELETE")
	router.HandleFunc("/delete", DeleteTransaction()).Methods("DELETE")
	router.HandleFunc("/ping", PingResponseProvider()).Methods("GET")
	router.HandleFunc("/", PingResponseProvider()).Methods("GET")
}