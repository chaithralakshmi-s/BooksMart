/**
		Payment API - Route Handler Module
		Author: Luis Otero
		
		Handles POST /transaction (using a Payment object to gather request)
		Handles GET /transactions (get all transactions)
		Handles GET /transactions/{id} (get transaction based on an ID)
		Handles POST /process (processes all transactions not processed yet)
		Handles PUT /update (update a payment based on an ID and amount)
		Handles DELETE /delete/{id} (deletes a payment transaction based on an ID)
**/

package main

import (
	"bytes"
	"net/http"
	"github.com/gorilla/mux"
	"github.com/unrolled/render"
	"gopkg.in/zegl/goriak.v3"
	"encoding/json"
	"github.com/satori/go.uuid"
)

//JSON Render machine
var jsonRender = render.New(render.Options{ IndentJSON: true,})

// RabbitMQ queue for sending transactions to be processed
var queue_server = "rabbit-616549229.us-west-1.elb.amazonaws.com" 
var queue_port = "5672"
var queue_name = "payments"
var queue_user = "guest"
var queue_pass = "guest"

//Riak database details (replace ip addresses ip1,ip2,ip3,ip4,ip5 with the ip addresses of each node in the Riak cluster)
//var cluster1 = []string{"54.183.63.24:8087", "52.53.234.97:8087", "13.57.244.4:8087", "13.57.249.22:8087", "54.67.3.147:8087"}
var cluster1 = []string{"riak-cluster-1486631455.us-west-1.elb.amazonaws.com:8087"}
var cluster2 = []string{"riak-elb-final-1701120976.us-west-1.elb.amazonaws.com:8087"}
var bucket_name = "payments"

/**
	POST /transactions
	
	1. Takes a payment
	2. Transforms it to a transaction
	3. Stores it into the database
	4. Pushes it to a queue for processing later
	5. Return a status object to the client
**/
func AddTransactionToQueue() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		
		var newPayment Payment;
		
		//decode payment request
		var decode_err = json.NewDecoder(request.Body).Decode(&newPayment)
		
		//check if payment parameters and object are valid
		if decode_err != nil {
			buf := new(bytes.Buffer)
			buf.ReadFrom(request.Body)
			body := buf.String()
			jsonRender.JSON(writer, 400, "Could not decode the payment request " + body)
			return
		}
		
		//create a new transaction object with payment parameters
		var newTransaction = Transaction {
			TransactionId: uuid.NewV4().String(),
			PaymentType: newPayment.PaymentType,
			UserDetails: User {
				Name: newPayment.Name,
				Id: newPayment.UsernameId,
				Password: newPayment.Password,
			},
			Amount: FloatToString(newPayment.Amount),
			Status: "Payment Pending",
		}
		
		var addresses []string
		
		number,_ := StringToInt(newPayment.UserId)
		
		if number%2 == 0 {
			addresses = cluster2
		}else {
			addresses = cluster1
		}
		
		//connect to database
		database,db_error := RiakConnect(addresses)
		
		if db_error != nil {
			jsonRender.JSON(writer, 500, "Could Not Connect to Database")
			return
		}
		
		//connect to queue
		_,channel,queue_error := RabbitmqConnect(queue_server, queue_port, queue_name, queue_user, queue_pass)
		
		if queue_error != nil {
			jsonRender.JSON(writer, 500, "Could Not Connect to Queue" + queue_server)
			return
		}
		
		//insert transaction to database for record keeping
		RiakSet(database, bucket_name, newTransaction.TransactionId, newTransaction)
		
		//send message to queue
		Enqueue(channel, queue_name, newTransaction.TransactionId)
		
		type TransactionState struct {
			Id string
			Status string
		}
		
		order := TransactionState{ Id : newTransaction.TransactionId, Status : "Payment Placement Successful!"}
		
		//send a status object to the client
		jsonRender.JSON(writer, http.StatusOK, order)
	}
}

/**
	GET /transactions and /transactions/{id}
	
	1. Checks for which of the two get operations to perform
	2. Get the transaction/s from the database
	3. Return the full transaction to the client
**/
func SearchForTransaction() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		parameters := mux.Vars(request)
		
		//connect to database
		database1, db1_connect_error := RiakConnect(cluster1)
		database2, db2_connect_error := RiakConnect(cluster2)
		
		var db []*goriak.Session
		
		if db1_connect_error != nil && db2_connect_error != nil {
			jsonRender.JSON(writer, 505, "Cannot access any of the data on the database.")
			return
		} else if db1_connect_error != nil || db2_connect_error != nil {
			if db1_connect_error == nil {
				db = []*goriak.Session{database1}
			} else {
				db = []*goriak.Session{database2}
			}
		} else {
			db = []*goriak.Session{database1, database2}
		}
		
		//Option 1: get all transactions. Option 2: get transaction based on ID specified
		if parameters["id"] == "" {
			var transactions []Transaction
			
			for _,database := range db {
				if set,error := RiakGetAll(database, bucket_name); error == nil {
					for _,value := range set {
						value.UserDetails.Id = ""
						value.UserDetails.Password = ""
						transactions = append(transactions, value)
					}
					
				}
			}
			
			if transactions != nil {
				jsonRender.JSON(writer, http.StatusOK, transactions)
			}else {
				jsonRender.JSON(writer, 505, "Cannot read from the nodes on either of the clusters")
			}
		}else {
			for _,database := range db {
				if transaction,err := RiakGet(database, bucket_name, parameters["id"]); err == nil {
					transaction.UserDetails.Id = ""
					transaction.UserDetails.Password = ""
					jsonRender.JSON(writer, http.StatusOK, transaction)
					return
				}
			}
			
			jsonRender.JSON(writer, 400, "Transaction not found in any of the clusters")
		}
	}
}

/**
	POST /process
	
	1. Dequeue the transactions
	2. Get all transactions from database
	3. Process transactions from queue
	4. Update values in the history
	5. Add unprocessed transactions to queue
	6. Return processed transactions to client
**/
func ProcessAllTransactions() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		//connect to queue
		_,channel,queue_connect_error := RabbitmqConnect(queue_server, queue_port, queue_name, queue_user, queue_pass)
		
		if queue_connect_error != nil {
			jsonRender.JSON(writer, 500, queue_connect_error.Error())
			return
		}
		
		//dequeue all transactions from queue
		ids,dequeue_error := DequeueAll(channel, queue_name)
		
		if dequeue_error != nil {
			jsonRender.JSON(writer, 500, dequeue_error.Error())
			return
		}
		
		//connect to database
		database1,db1_error := RiakConnect(cluster1)
		database2,db2_error := RiakConnect(cluster2)
		
		var db []*goriak.Session
		
		if db1_error != nil && db2_error != nil {
			jsonRender.JSON(writer, 505, "Cannot access the data for any of the clusters available")
			return
		}else if db1_error != nil || db2_error != nil {
			if db1_error == nil {
				db = []*goriak.Session{database1}
			}else {
				db = []*goriak.Session{database2}
			}
		}else {
			db = []*goriak.Session{database1,database2}
		}
		
		//process all of the transactions from the queue
		
		var transactions []Transaction
		for _,value := range ids {
			for _,database := range db {
				if transaction,get_error := RiakGet(database, bucket_name, value); get_error == nil {
					if transaction.Status != "Payment Process Success!" {
						transaction.Status = "Payment Process Success!"
						transaction.UserDetails.Id = ""
						transaction.UserDetails.Password = ""
						transactions = append(transactions, transaction)
						RiakSet(database, bucket_name, transaction.TransactionId, transaction)
					}
				}
			}
		}
		
		if transactions != nil {
			jsonRender.JSON(writer, http.StatusOK, transactions)
		}
		
		for _,database := range(db){
			//get all of the transactions from the database
			if set,get_all_error := RiakGetAll(database, bucket_name); get_all_error == nil {
				//push transactions to queue that have not been processed
				for _,value := range set {
					if value.Status == "Payment Pending" {
						push_error := Enqueue(channel, queue_name, value.TransactionId)
						
						if push_error != nil {
							jsonRender.JSON(writer, 500, push_error.Error())
							return
						}
					}
				}
			}
			
		}
		
	}

}

/**
	PUT /update
	
	1. Create update and status objects
	2. Attempt to decode update request
	3. Get the transaction from the database
	4. Check if the transaction can be updated
	5. Update the transaction
	6. Store the transaction in the database
	7. Return a update status to the client
**/
func UpdateTransaction() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		//update object struct
		type update_details struct {
			Id string
			UserId string
			Amount float64
		}
		
		//status object struct
		type update_status struct {
			Id string
			Status string
		}
		
		var new_update update_details
		
		//decode update request
		parse_error := json.NewDecoder(request.Body).Decode(&new_update)
		
		if parse_error != nil {
			jsonRender.JSON(writer, 400, parse_error.Error())
			return
		}
		
		//connect to queue
		_,channel,queue_connect_error := RabbitmqConnect(queue_server, queue_port, queue_name, queue_user, queue_pass)
		
		if queue_connect_error != nil {
			jsonRender.JSON(writer, 400, update_status{ Id : new_update.Id, Status : "Payment already processed. Cannot update payment amount"})
			return
		}
		
		//dequeue all transactions from queue
		ids,dequeue_error := DequeueAll(channel, queue_name)
		
		if dequeue_error != nil {
			jsonRender.JSON(writer, 500, dequeue_error.Error())
			return
		}
		
		isInQueue := false
		
		for _,value := range ids {
			if new_update.Id == value {
				isInQueue = true
				break
			}
		}
		
		if !isInQueue {
			jsonRender.JSON(writer, 400, update_status{ Id : new_update.Id, Status : "Payment already processed. Cannot update payment amount"})
			return
		}
		
		var address []string
		
		number,_ := StringToInt(new_update.UserId)
		
		if number%2 == 0 {
			address = cluster2
		}else {
			address = cluster1
		}
		
		//connect to the database
		database,db_error := RiakConnect(address)
		
		if db_error != nil {
			jsonRender.JSON(writer, 500, db_error.Error())
			return
		}
		
		//Get transaction from the database
		newTransaction, get_error := RiakGet(database, bucket_name, new_update.Id)
		
		if get_error != nil {
			jsonRender.JSON(writer, 400, get_error.Error())
			return
		}
		
		//Update transaction and store
		if newTransaction.Status == "Payment Pending" {
			newTransaction.Amount = FloatToString(new_update.Amount)
			store_error := RiakSet(database, bucket_name, newTransaction.TransactionId, newTransaction)
			
			if store_error != nil {
				jsonRender.JSON(writer, 500, store_error.Error())
				return
			}
			
			//send status to client
			jsonRender.JSON(writer, http.StatusOK, update_status{ Id : new_update.Id, Status : "Payment update successful", })
		}else {
			jsonRender.JSON(writer, 400, update_status{ Id : new_update.Id, Status : "Payment already processed. Cannot update payment amount"})
		}		
		
	}
}

/**
	DELETE /delete/{id}
	
	1. Checks if an id was given
	2. Delete transaction from database
	3. Return status of deletion to client
**/
func DeleteTransaction() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		parameters := mux.Vars(request)
		
		//connect to the databse
		database1,connect1_err := RiakConnect(cluster1)
		database2, connect2_err := RiakConnect(cluster2)
		
		var db []*goriak.Session
		
		if connect1_err != nil && connect2_err != nil{
			jsonRender.JSON(writer, 600, "Could not connect to the database")
			return
		}else if connect1_err != nil || connect2_err != nil {
			if connect1_err == nil {
				db = []*goriak.Session{database1}
			}else {
				db = []*goriak.Session{database2}
			}
		}else {
			db = []*goriak.Session{database1,database2}
		}
		
		//check if an ID has been provided
		if parameters["id"] == "" {
				jsonRender.JSON(writer, 404, "Must have a key id to delete. /delete/{id}")
			
		}else {
			for _,database := range db {
				if error := RiakDelete(database, bucket_name, parameters["id"]); error == nil {
					//return status to client
					jsonRender.JSON(writer, http.StatusOK, parameters["id"] + " deleted!")
					return
				}
			}
			jsonRender.JSON(writer, 500, "Cannot delete the payment transaction because it was not found")
			
		}
		

	}
}

/**
	GET /ping and /
	
	1. Returns a json status of the API to the client
**/
func PingResponseProvider() http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		type ping struct{
			API string
			Status string
		}
		
		jsonRender.JSON(writer, http.StatusOK, ping{ API: "Payment API v1", Status: "Server is running on port 4500",})
	}
}