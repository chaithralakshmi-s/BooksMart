/**
		Payment API v1
		Author: Luis Otero
		
		Main Features:
			1. Place a payment
			2. Store a payment into a payment history bucket
			3. Get a payment based on the payment ID
			4. Get all payments from the payment history bucket
			5. Process all payment transactions
			6. Update a payment that hasn't been processed
			7. Delete a payment transaction
**/
package main

import (
	"os"
)

/**
	Main Function
	
	1. Get Sever object
	2. Get Environment variables
	3. Start server
**/
func main() {

	//Get server object
        api_server := CreateServer()
     
	//Get environment variables
	port := ":" + os.Getenv("PORT")
	if port == ":" {
		port = ":4500"
	}
        
	//Start listening on specified port
	api_server.Run(port)
	
}
