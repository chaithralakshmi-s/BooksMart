/**
		Payment API - Helper Function Module
		Author: Luis Otero
		
		Database functions (Connect, Get, Set, GetAll, Delete)
		RabbitMQ functions (Connect, Enqueue, Dequeue)
		Format functions (Serialize, Unserialize, FloatToString, StringToFloat)
**/
package main

import (
	"encoding/json"
	"github.com/streadway/amqp"
	"gopkg.in/zegl/goriak.v3"
	"strconv"
)

/**
	Float to String Conversion Function
	
	1. Take a Float64 number
	2. Convert the number to a string
	3. Return the converted number
**/
func FloatToString(number float64) string {
	conversion := strconv.FormatFloat(number, 'f', 2, 64)
	
	return conversion
}

/**
	String to Float Conversion Function
	
	1. Take a string that resembles a Float64 number
	2. Attempt to convert the string to a Float64 number
	3. Return the converted string (if it was possible)
**/
func StringToFloat(number string) (float64, error) {
	conversion, error := strconv.ParseFloat(number, 64)
	
	return conversion,error
}

/**
	String to Int Conversion Function
	
	1. Take a string that resembles an integer
	2. Attempt to convert the string to an integer
	3. Return the converted string (if it was possible)
**/
func StringToInt(number string) (int, error) {
	conversion, error := strconv.Atoi(number)
	
	return conversion,error
}

/**
	Marshal a Transaction object Function
	
	1. Take a Transaction object
	2. Convert the object into a byte array
	3. Return the byte array (if conversion was possible)
**/
func Serialize(object Transaction) ([]byte, error) {
	transaction,error := json.Marshal(object)
	
	return transaction,error
}

/**
	Unmarshal a Transaction object Function
	
	1. Take a byte array of data that holds a Transaction object
	2. Convert the byte array into a Transaction object
	3. Return the transaction object (if it was possible)
**/
func Unserialize(data []byte) (Transaction, error) {
	var transaction Transaction
	
	error := json.Unmarshal(data,&transaction)
	
	return transaction,error
}

/**
	Connect to a RabbitMQ instance Function
	
	1. Connect to a rabbitmq instance
	2. Create a channel for the connection
	3. Attempt to make a queue (if not already created)
	4. Return the connection and channel for the rabbitMQ instance
	
	@Reference Partial code provided by Professor Paul Nguyen from goapi
 **/
func RabbitmqConnect(server string, port string, queue string, user string, pass string) (*amqp.Connection, *amqp.Channel, error){

	//Attempt a connection to a rabbitmq instance
	rabbit_connection, connect_err := amqp.Dial("amqp://"+user+":"+pass+"@"+server+":"+port+"/")
	
	if connect_err != nil {
		return nil, nil, connect_err
	}
	
	//Create a channel for the connection
	channel, ch_err := rabbit_connection.Channel()
	
	if(ch_err != nil) {
		return nil, nil, ch_err
	}
	
	//Attempt to create a queue for the connection
	_,queue_creation_err := channel.QueueDeclare(
		queue_name, //name of queue
		false,      //Durability (True: durable, False: not durable)
		false,      //Delete When Not Used (True: yes, False: no)
		false,      //Is It Exclusive (True: yes, False, no)	
		false,      //Delay (True: yes, False: no-wait)
		nil,        //Optional Arguments
	)
	
	if queue_creation_err != nil {
		return nil, nil, queue_creation_err
	}
	
	return rabbit_connection, channel, nil
}

/**
	Enqueue Payment Message to RabbitMQ Function
	
	1. Get a channel object for rabbitmq instance
	2. Attempt to create a queue (if it doesn't already exist)
	3. Push content to the queue
	4. Return any errors that might have occurred
	
	@Reference Partial code provide by Professor Paul Nguyen from goapi
**/
func Enqueue(channel *amqp.Channel, queue_name string, transaction string) error {
	
	//Attempt to create a queue for the connection
	queue, queue_error := channel.QueueDeclare(
		queue_name, //name of queue
		false,      //Durability (True: durable, False: not durable)
		false,      //Delete When Not Used (True: yes, False: no)
		false,      //Is It Exclusive (True: yes, False, no)	
		false,      //Delay (True: yes, False: no-wait)
		nil,        //Optional Arguments
	)
	
	if queue_error != nil {
		return queue_error
	}

	//Attempt to push the message into the queue via Publish function
	publish_error := channel.Publish(
		"",     					//Exchange
		queue.Name, 					//Name of queue
		false,  					//Is the key mandatory (True: yes, False: no)
		false,  					//Execute immediately (True: yes, False: no)
		amqp.Publishing{				//Parameters
			ContentType: "text/plain",		//Content Type
			Body:        []byte(transaction),	//Message
		})
		
	if publish_error != nil {
		return publish_error
	}
	
	return nil
}

/**
	Dequeue the entire queue Function
	
	1. Attempt to create a queue
	2. Dequeue all of the messages from the queue
	3. Pass on the messages asynchronously to a channel buffer
	4. Extrach the messages from the channel
	5. Return the messages
	
	@Reference Partial Code Provided by Professor Paul Nguyen from goapi
**/
func DequeueAll(channel *amqp.Channel, rabbitmq_queue string) ([]string, error) {
	var ids []string //message buffer
	
	//Attempt to create a queue for the connection
	queue, queue_creation_error := channel.QueueDeclare(
		queue_name, //name of queue
		false,      //Durability (True: durable, False: not durable)
		false,      //Delete When Not Used (True: yes, False: no)
		false,      //Is It Exclusive (True: yes, False, no)	
		false,      //Delay (True: yes, False: no-wait)
		nil,        //Optional Arguments
	)

	if queue_creation_error != nil {
		return nil, queue_creation_error
	}

	//Attempt to dequeue the queue and store all messages
	order_list, queue_consumption_error := channel.Consume(
		queue.Name, 	//Route Key
		"payments",     //Queue Name
		true,   	//Auto Acknowledgement (True: yes, False: no)
		false,  	//Is It Exclusive (True: yes, False: no)
		false,  	//No Local (True: yes, False: no)
		false,  	//Delay (True: yes, False: no)
		nil,    	//Optional Arguments
	)
	
	if queue_consumption_error != nil {
		return nil, queue_consumption_error
	}
	
	//Create a channel to receive the data 
	id_channel := make(chan string)
	
	//Grab the messages from the list asynchronously
	go func() {
		for value := range order_list {
			id_channel <- string(value.Body)
		}
		close(id_channel)
	}()

	//Attempt to close the channel
	cancel_error := channel.Cancel("payments", false)
	if cancel_error != nil {
	    return nil, cancel_error
	}

	//Get all the vaues from the channel
	for value := range id_channel {
		ids = append(ids, value)
	}	
	
	return ids, nil
}

/**
	Connect to Riak Cluster Function
	
	1. Get a set of addresses (ip:port)
	2. Add the addresses to an option variable
	3. Connect to Riak Cluster with addresses given
	4. Return the session object
**/
func RiakConnect(addresses []string) (*goriak.Session, error) {
	options := goriak.ConnectOpts {
		Addresses: addresses,
	}
	
	session, err := goriak.Connect(options)
	
	return session,err
}

/**
	Send a Transaction object to the Riak Cluster Function
	
	1. Get a session, Transaction object, and key identifiers
	2. Use the key identifiers to attempt to insert the object into the cluster
	3. Return any errors that might have occurred
**/
func RiakSet(db *goriak.Session, bucket string, key string, object Transaction) error {
	// bucket = <bucket name> "maps" = <bucket type>
	_,err := goriak.Bucket(bucket, "maps").Set(object).Key(key).Run(db)
	
	return err
}

/**
	Get all of the payment transactions Function
	
	1. Get a session and key identifiers
	2. Use the key identifiers to retrieve all of the keys from the Riak cluster
	3. Store each of the keys in a buffer
	4. Attempt to get the values for each of the keys retrieved
	5. Return the payment transactions (if successful)
**/
func RiakGetAll(db *goriak.Session, bucket string) ([]Transaction, error) {
	var transactions []Transaction
	
	//Attempt to get all of the keys from the bucket
	_,err := goriak.Bucket(bucket, "maps").AllKeys(func(str []string) error {
		//From the keys, extract all of the values
		for _,value := range str {
			var temp Transaction
			
			temp,error := RiakGet(db,bucket,value)
			if error != nil {
				return error
			}
			transactions = append(transactions, temp)
		}
		
		return nil
		
	}).Run(db)
	
	return transactions,err
}

/**
	Get a value from a Riak Cluster Function
	
	1. Get a session object and key identifiers
	2. Use the key identifiers to attempt to get a value from the cluster
	3. Return the value based on the key (if value was found)
**/
func RiakGet(db *goriak.Session, bucket string, key string) (Transaction, error) {
	var value Transaction
	
	//Attempt to get value -> bucket = <bucket_name> "maps" = <bucket_type>
	_,error := goriak.Bucket(bucket, "maps").Get(key, &value).Run(db)
	
	return value,error	
}

/**
	Delete an object from a Riak Cluster Function
	
	1. Get a session object and key identifiers
	2. Use the key identifiers to execute the delete function to attempt to delete an object
	3. Return any errors that might have occurred
**/
func RiakDelete(db *goriak.Session, bucket string, key string) error {

	//Attempt to delete an object -> bucket = <bucket_name> "maps" = <bucket_type>
	_,error := goriak.Bucket(bucket, "maps").Delete(key).Run(db)
	
	return error
}