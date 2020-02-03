package main

import (
	"fmt"
	"log"
	"net/http"
	"io/ioutil"
	"time"
	//      "strings"
	//"encoding/json"
	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
	"github.com/unrolled/render"
	//"github.com/satori/go.uuid"


	"encoding/json"
	"strings"
)


var debug=true
var server1 = "http://52.53.118.157:8098"
var server2 = "http://52.52.124.207:8098"
var server3 = "http://54.176.35.69:8098"
var server4 = "http://54.219.105.255:8098"
var server5 = "http://54.241.239.217:8098"

type Client struct {
	Endpoint string
	*http.Client
}

var tr = &http.Transport{
	MaxIdleConns:       10,
	IdleConnTimeout:    30 * time.Second,
	DisableCompression: true,
}

func NewClient(server string) *Client {
	return &Client{
		Endpoint:       server,
		Client:         &http.Client{Transport: tr},
	}
}

// NewServer configures and returns a Server.
func NewServer() *negroni.Negroni {
	formatter := render.New(render.Options{
		IndentJSON: true,
	})
	n := negroni.Classic()
	mx := mux.NewRouter()
	initRoutes(mx, formatter)
	n.UseHandler(mx)
	return n
}

func (c *Client) Ping() (string, error) {
	resp, err := c.Get(c.Endpoint + "/ping" )
	if err != nil {
		fmt.Println("[RIAK DEBUG] " + err.Error())
		return "Ping Error!", err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if debug { fmt.Println("[RIAK DEBUG] GET: " + c.Endpoint + "/ping => " + string(body)) }
	return string(body), nil
}


// Init Database Connections
func init() {

	// Riak KV Setup
	c1 := NewClient(server1)
	msg, err := c1.Ping( )
	if err != nil {
		log.Fatal(err)
	} else {
		log.Println("Riak Ping Server1: ", msg)
	}

	c2 := NewClient(server2)
	msg, err = c2.Ping( )
	if err != nil {
		log.Fatal(err)
	} else {
		log.Println("Riak Ping Server2: ", msg)
	}

	c3 := NewClient(server3)
	msg, err = c3.Ping( )
	if err != nil {
		log.Fatal(err)
	} else {
		log.Println("Riak Ping Server3: ", msg)
	}

	c4 := NewClient(server4)
	msg, err = c4.Ping( )
	if err != nil {
		log.Fatal(err)
	} else {
		log.Println("Riak Ping Server4: ", msg)
	}

	c5 := NewClient(server5)
	msg, err = c5.Ping( )
	if err != nil {
		log.Fatal(err)
	} else {
		log.Println("Riak Ping Server5: ", msg)
	}

}






// Helper Functions
func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
		panic(fmt.Sprintf("%s: %s", msg, err))
	}
}


// API Ping Handler
func pingHandler(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		formatter.JSON(w, http.StatusOK, struct{ Test string }{"API version 1.0 alive!"})
	}

}
//


func (c *Client) AddUserTransactions(key string, user_transaction UserTransactionInput) (UserTransactionInput, error) {
	var ut_nil= UserTransactionInput{}



	reqbody :=  "{\"update\": {\"user_register\": \""+
		key +
		"\",\"usertransactionids_set\": {\"add_all\": [\""+
		user_transaction.TransactionId+
		"\"]}}}"

	url:=c.Endpoint+"/types/maps/buckets/usertransactions/datatypes/"+key+"?returnbody=true"

	fmt.Println(reqbody + "key is "+key)
	req, _ := http.NewRequest("POST", url,strings.NewReader(reqbody))
	req.Header.Add("Content-Type", "application/json")
	fmt.Println("Request is: ");
	fmt.Println(req )
	fmt.Println("End of Request ");
	resp, err := c.Do(req)


	if err != nil {
		fmt.Println("[RIAK DEBUG] " + err.Error())
		return ut_nil, err
	}


	defer resp.Body.Close()


	body, err := ioutil.ReadAll(resp.Body)
	if debug {
		fmt.Println("[RIAK DEBUG] PUT: " + url+" => " + string(body))
	}

	var ut= UserTransactionInput{}
	if err := json.Unmarshal(body, &ut); err != nil {
		fmt.Println("[RIAK DEBUG] JSON unmarshaling failed: %s", err)
		return ut_nil, err
	}
	return ut, nil


}


func (c *Client) AddTransactionDetails( user_transaction UserTransactionInput) (UserTransactionInput, error) {



	var ut_nil= UserTransactionInput{}
	transaction_time:=time.Now().Format("02-Jan-2006 15:04:05")
	key:=user_transaction.TransactionId

	reqbody :=  "{\"update\":{\"1transactionid_register\":\""+key+
		"\",\"2time_register\":\""+transaction_time+
		"\",\"4amount_register\":\""+user_transaction.Amount+
		"\",\"3products_set\":{\"add_all\":[\"" +user_transaction.Products[0]+"\""

	for i := 1; i < len(user_transaction.Products); i++ {
		reqbody=reqbody+",\""+user_transaction.Products[i]+"\""

	}
	reqbody=reqbody+"]}}"
	fmt.Println(reqbody)



	url:=c.Endpoint+"/types/maps/buckets/usertransactions/datatypes/"+key+"?returnbody=true"

	fmt.Println(reqbody + "key is "+key)
	req, _ := http.NewRequest("POST", url,strings.NewReader(reqbody))
	req.Header.Add("Content-Type", "application/json")
	fmt.Println("Request is: ");
	fmt.Println(req )
	fmt.Println("End of Request ");
	resp, err := c.Do(req)


	if err != nil {
		fmt.Println("[RIAK DEBUG] " + err.Error())
		return ut_nil, err
	}


	defer resp.Body.Close()


	body, err := ioutil.ReadAll(resp.Body)
	if debug {
		fmt.Println("[RIAK DEBUG] PUT: " + url+" => " + string(body))
	}

	var ut= UserTransactionInput{}
	if err := json.Unmarshal(body, &ut); err != nil {
		fmt.Println("[RIAK DEBUG] JSON unmarshaling failed: %s", err)
		return ut_nil, err
	}
	//	var ut_nil= UserTransaction{}

	/*for i := 0; i < len(user_transaction.TransactionId); i++ {
		if i < (len(user_transaction.TransactionId)-1) {

			user_transaction.TransactionId[i] = "\"" + user_transaction.TransactionId[i] + "\","
		} else{
			user_transaction.TransactionId[i] = "\"" + user_transaction.TransactionId[i] + "\""
		}
	}
	fmt.Println("AFter append:")
	fmt.Println(user_transaction.TransactionId)*/





	return ut, nil
}

func (c *Client) GetTransactionDetails(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {


	fmt.Println("In  getUserTransactionsHandler")
	params := mux.Vars(req)
	var uuid string = params["id"]
	fmt.Println( "User Params ID: ", uuid )

	if uuid == "" {
		formatter.JSON(w, http.StatusBadRequest, "Invalid Request. User ID Missing.")
	} else {

		c := NewClient(server1)

		transactions:=UserTransactions{}
		transactions = c.GetTransactionIds(uuid)
		fmt.Println("After  GetTransactionIds")
		fmt.Println("Your transactions are here: ", transactions)


		if transactions.UserName  == "" {
			formatter.JSON(w, http.StatusBadRequest, "")
		} else {
			fmt.Println("Your transactions are in statusok: ", transactions)
			formatter.JSON(w, http.StatusOK ,transactions)
			fmt.Println(&w)
			fmt.Println(w)
			p:=&w
			fmt.Println(*p)

		}
	}
	fmt.Println("End of  getUserTransactionsHandler")

}

return nil}

func addUserTransactionHandler(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {

		fmt.Println("In addUserTransactionHandler");

		params := mux.Vars(req)

		var uname string = params["id"]

		fmt.Println( "addUserTransactionHandler:UserHistory Params ID: ", uname )

		var utransaction UserTransactionInput
		_ = json.NewDecoder(req.Body).Decode(&utransaction)
		fmt.Println("User Name is: "+utransaction.UserName)

		if uname == ""  {
			formatter.JSON(w, http.StatusBadRequest, "Invalid Request. User ID Missing.")
		} else {
			c1 := NewClient(server1)
			trns, err := c1.AddUserTransactions(uname, utransaction)
			if err != nil {
				log.Fatal(err)
				formatter.JSON(w, http.StatusBadRequest, err)
			} else {
				formatter.JSON(w, http.StatusOK, trns)
			}
			c2 := NewClient(server1)
			tds, err := c2.AddTransactionDetails( utransaction)
			if err != nil {
				log.Fatal(err)
				formatter.JSON(w, http.StatusBadRequest, err)
			} else {
				formatter.JSON(w, http.StatusOK, tds)
			}
			c3 := NewClient(server1)
			tds, err = c3.AddTransactionDetails( utransaction)
			if err != nil {
				log.Fatal(err)
				formatter.JSON(w, http.StatusBadRequest, err)
			} else {
				formatter.JSON(w, http.StatusOK, tds)
			}
			c4 := NewClient(server1)
			tds, err = c4.AddTransactionDetails( utransaction)
			if err != nil {
				log.Fatal(err)
				formatter.JSON(w, http.StatusBadRequest, err)
			} else {
				formatter.JSON(w, http.StatusOK, tds)
			}
			c5 := NewClient(server1)
			tds, err = c5.AddTransactionDetails( utransaction)
			if err != nil {
				log.Fatal(err)
				formatter.JSON(w, http.StatusBadRequest, err)
			} else {
				formatter.JSON(w, http.StatusOK, tds)
			}
		}




		fmt.Println("End of addUserTransactionHandler");

	}



}



func getTransactionDetailsHandler(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {

		fmt.Println("In getTransactionDetailsHandler");

		params := mux.Vars(req)

		var uname string = params["id"]

		fmt.Println( "addUserTransactionHandler:UserHistory Params ID: ", uname )

		var utransaction UserTransactionInput
		_ = json.NewDecoder(req.Body).Decode(&utransaction)
		fmt.Println("User Name is: "+utransaction.UserName)

		if uname == ""  {
			formatter.JSON(w, http.StatusBadRequest, "Invalid Request. User ID Missing.")
		} else {
			c1 := NewClient(server1)
			trns, err := c1.AddUserTransactions(uname, utransaction)
			if err != nil {
				log.Fatal(err)
				formatter.JSON(w, http.StatusBadRequest, err)
			} else {
				formatter.JSON(w, http.StatusOK, trns)
			}
			c2 := NewClient(server1)
			tds, err := c2.AddTransactionDetails( utransaction)
			if err != nil {
				log.Fatal(err)
				formatter.JSON(w, http.StatusBadRequest, err)
			} else {
				formatter.JSON(w, http.StatusOK, tds)
			}
		}




		fmt.Println("End of addUserTransactionHandler");

	}



}



func productGet(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {

		params := mux.Vars(req)
		var uuid string = params["id"]
		fmt.Println( "product Params ID: ", uuid )



		c1 := make(chan UserTransactionIds)
		c2 := make(chan UserTransactionIds)
		c3 := make(chan UserTransactionIds)
		c4 := make(chan UserTransactionIds)
		c5 := make(chan UserTransactionIds)

		go GetTransactionIdsServer1(uuid, c1)
		go GetTransactionIdsServer2(uuid, c2)
		go GetTransactionIdsServer3(uuid, c3)
		go GetTransactionIdsServer4(uuid, c4)
		go GetTransactionIdsServer5(uuid, c5)

		var tid UserTransactions
		select {
		case tid = <-c1:
			fmt.Println("Received Server1: ", tid)
		case tid = <-c2:
			fmt.Println("Received Server2: ", tid)
		case tid = <-c3:
			fmt.Println("Received Server3: ", tid)
		case tid = <-c4:
			fmt.Println("Received Server4: ", tid)
		case tid = <-c5:
			fmt.Println("Received Server5: ", tid)
		}

		if tid.UserName == "" {
			formatter.JSON(w, http.StatusBadRequest, "")
		} else {
			fmt.Println( "product: ", tid )
			formatter.JSON(w, http.StatusOK, tid)
		}

	}
}


func GetTransactionIdsServer1(uuid string,chn chan<- []UserTransactionIds) {

	var tid_nil []UserTransactionIds
	c := NewClient(server1)
	tids, err := c.GetTransactionIds(uuid)
	if err != nil {
		chn <- tid_nil
	} else {
		fmt.Println( "Server1: ", tids)
		chn <- tids
	}
}


func GetTransactionIdsServer2(uuid string,chn chan<- []UserTransactionIds) {

	var tid_nil []UserTransactionIds
	c := NewClient(server2)
	tids, err := c.GetTransactionIds()
	if err != nil {
		chn <- tid_nil
	} else {
		fmt.Println( "Server2: ", tids)
		chn <- tids
	}
}

func GetTransactionIdsServer3(uuid string,chn chan<- []UserTransactionIds) {

	var tid_nil []UserTransactionIds
	c := NewClient(server3)
	tids, err := c.GetTransactionIds(uuid)
	if err != nil {
		chn <- tid_nil
	} else {
		fmt.Println( "Server3: ", tids)
		chn <- tids
	}
}


func GetTransactionIdsServer4(uuid string,chn chan<- []UserTransactionIds) {

	var tid_nil []UserTransactionIds
	c := NewClient(server4)
	tids, err := c.GetTransactionIds(uuid)
	if err != nil {
		chn <- tid_nil
	} else {
		fmt.Println( "Server4: ", tids)
		chn <- tids
	}
}

func GetTransactionIdsServer5(uuid string,chn chan<- []UserTransactionIds) {

	var tid_nil []UserTransactionIds
	c := NewClient(server5)
	tids, err := c.GetTransactionIds(uuid)
	if err != nil {
		chn <- tid_nil
	} else {
		fmt.Println( "Server5: ", tids)
		chn <- tids
	}
}


func (c *Client) GetTransactionIds(key string) (UserTransactionIds, error) {
	var tid_nil = UserTransactionIds {}
	resp, err := c.Get(c.Endpoint + "/types/maps/buckets/usertransactions/datatypes/"+key )
	if err != nil {
		fmt.Println("[RIAK DEBUG] " + err.Error())
		return tid_nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if debug { fmt.Println("[RIAK DEBUG] GET: " + c.Endpoint + "/types/maps/buckets/usertransactions/datatypes/"+key +" => " + string(body)) }
	var tid = UserTransactionIds { }
	if err := json.Unmarshal([]byte(body), &tid); err != nil {
		fmt.Println("[RIAK DEBUG] JSON unmarshaling failed: %s", err)
		return tid_nil, err
	}
	//fmt.Println(prd)
	return tid, nil
}


func GetTransactionDetailsServer1(uuid string,chn chan<- UserTransactions) {

	var tid_nil UserTransactions
	c := NewClient(server1)
	tds, err := c.GetTransactionIds(uuid)
	if err != nil {
		chn <- tid_nil
	} else {
		fmt.Println( "Server1: ", tds)
		chn <- tds
	}
}

func GetTransactionDetailsServer2(uuid string,chn chan<- UserTransactions) {

	var tid_nil UserTransactions
	c := NewClient(server1)
	tds, err := c.GetTransactionIds(uuid)
	if err != nil {
		chn <- tid_nil
	} else {
		fmt.Println( "Server2: ", tds)
		chn <- tds
	}
}



func GetTransactionDetailsServer3(uuid string,chn chan<- UserTransactions) {

	var tid_nil UserTransactions
	c := NewClient(server1)
	tds, err := c.GetTransactionIds(uuid)
	if err != nil {
		chn <- tid_nil
	} else {
		fmt.Println( "Server3: ", tds)
		chn <- tds
	}
}




func GetTransactionDetailsServer4(uuid string,chn chan<- UserTransactions) {

	var tid_nil UserTransactions
	c := NewClient(server1)
	tds, err := c.GetTransactionIds(uuid)
	if err != nil {
		chn <- tid_nil
	} else {
		fmt.Println( "Server4: ", tds)
		chn <- tds
	}
}


func GetTransactionDetailsServer5(uuid string,chn chan<- UserTransactions) {

	var tid_nil UserTransactions
	c := NewClient(server1)
	tds, err := c.GetTransactionIds(uuid)
	if err != nil {
		chn <- tid_nil
	} else {
		fmt.Println( "Server5: ", tds)
		chn <- tds
	}
}



// API Routes
func initRoutes(mx *mux.Router, formatter *render.Render) {
	mx.HandleFunc("/ping", pingHandler(formatter)).Methods("GET")
	mx.HandleFunc("/addtransaction/{id}", addUserTransactionHandler(formatter)).Methods("POST")
	mx.HandleFunc("/getusertransactions/{id}", getUserTransactionsHandler(formatter)).Methods("GET")
	mx.HandleFunc("/getTransactionDetails/{id}", getTransactionDetailsHandler(formatter)).Methods("GET")

}