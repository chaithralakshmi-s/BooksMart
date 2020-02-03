/*
	User API in Go
	 Riak KV
*/

package main

import (
	"fmt"
	"log"
	"net/http"
	//"bufio"
	"io/ioutil"
	"time"
	//"os"
	"strings"
	"encoding/json"
	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
	"github.com/unrolled/render"
	//"github.com/satori/go.uuid"
    
)

/* Riak REST Client */
var debug = false
var server1 = "http://35.167.17.151:8098" // set in environment
var server2 = "http://54.190.63.248:8098" // set in environment
var server3 = "http://54.218.184.246:8098" // set in environment
var server4 = "http://13.56.226.53:8098" // set in environment
var server5 = "http://54.193.104.250:8098" // set in environment

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
		Endpoint:  	server,
		Client: 	&http.Client{Transport: tr},
	}
}

func (c *Client) Ping() (string, error) {
	resp, err := c.Get(c.Endpoint + "/ping" )
	if err != nil {
		fmt.Println("[RIAK DEBUG] " + err.Error())
		return "Ping Error!", err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if debug { fmt.Println("[RIAK DEBUG] GET: " + c.Endpoint  + "/ping => " + string(body)) }
	return string(body), nil
}


func (c *Client) RegisterUser(key string, reqbody string) (user, error) {
	var ord_nil = user {}
	

	 resp, err := c.Post(c.Endpoint + "/types/Usertype/buckets/person/keys/"+key+"?returnbody=true", 
		"application/json", strings.NewReader(reqbody) )
		
		if err != nil {
			fmt.Println("[RIAK DEBUG] " + err.Error())
			return ord_nil, err
		}	
 defer resp.Body.Close()
 body, err := ioutil.ReadAll(resp.Body)
 if debug { fmt.Println("[RIAK DEBUG] POST: " + c.Endpoint + "/types/Usertype/buckets/person/keys/"+key+"?returnbody=true => "  + string(body)) }
 var place user
  msg1 := json.Unmarshal(body, &place); 
 if msg1 != nil {
	fmt.Println("[RIAK DEBUG] JSON unmarshaling failed: %s", msg1)
	return ord_nil, msg1
}

 return place, nil
}

func (c *Client) GetUser(key string) (user, error) {
	var ord_nil = user {}
	
	resp, err := c.Get(c.Endpoint + "/types/Usertype/buckets/person/keys/"+key)
	
	if err != nil {
		fmt.Println("[RIAK DEBUG] " + err.Error())
		return ord_nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if debug { fmt.Println("[RIAK DEBUG] GET: " + c.Endpoint + "/buckets/Usertype/keys/"+key +" => " + string(body)) }
	var ord = user { }
	if err := json.Unmarshal(body, &ord); err != nil {
		fmt.Println("RIAK DEBUG] JSON unmarshaling failed: %s", err)
		return ord_nil, err
	}
	return ord, nil
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
		log.Println("Riak Ping Server3: ", msg)		
	}

	c5 := NewClient(server5)
	msg, err = c5.Ping( )
	if err != nil {
		log.Fatal(err)
	} else {
		log.Println("Riak Ping Server3: ", msg)		
	}

}


// API Routes
func initRoutes(mx *mux.Router, formatter *render.Render) {
	mx.HandleFunc("/ping", pingHandler(formatter)).Methods("GET")
	mx.HandleFunc("/user", createNewUserhandler(formatter)).Methods("POST")
	mx.HandleFunc("/user/{id}", RetrieveUserhandler(formatter)).Methods("GET")
	
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
		formatter.JSON(w, http.StatusOK, struct{ Test string }{"User Login!"})
	}
}

// API Create New User
func createNewUserhandler(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		var ord user
		//uuid, _ := uuid.NewV4()
		decoder := json.NewDecoder(req.Body)
		err := decoder.Decode(&ord)
		
		if err != nil {
			ErrorWithJSON(w, "Incorrect body", http.StatusBadRequest)
			fmt.Println("[HANDLER DEBUG] ", err.Error())
			return
		}
		
		reqbody, _ := json.Marshal(ord)

		c1 := NewClient(server1)
		
		val_resp, err := c1.RegisterUser(ord.UserId,string(reqbody))
		
		if err != nil {
			log.Fatal(err)
			formatter.JSON(w, http.StatusBadRequest, err)
		} else {
			formatter.JSON(w, http.StatusOK, val_resp)
		}
		}
	}

// API Get User
func RetrieveUserhandler(formatter *render.Render) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		
		params := mux.Vars(req)
		fmt.Println(params)
		
		var uuid string = params["id"]
		
		c1 := make(chan user)
    	c2 := make(chan user)
		c3 := make(chan user)
		c4 := make(chan user)
		c5 := make(chan user)

		if uuid == ""  {
			fmt.Println(uuid)
			formatter.JSON(w, http.StatusBadRequest, "Invalid Request. ID Missing.")
		} else {

			go getOrderServer1(uuid, c1) 
			go getOrderServer2(uuid, c2) 
			go getOrderServer3(uuid, c3) 
			go getOrderServer3(uuid, c4) 
			go getOrderServer3(uuid, c5) 

			var ord user
		  	select {
			    case ord = <-c1:
			        fmt.Println("Received Server1: ", ord)
			    case ord = <-c2:
			        fmt.Println("Received Server2: ", ord)
			    case ord = <-c3:
					fmt.Println("Received Server3: ", ord)
				case ord = <-c4:
					fmt.Println("Received Server4: ", ord)
				case ord = <-c5:
			        fmt.Println("Received Server5: ", ord)
		    }

			if ord == (user{}) {
				formatter.JSON(w, http.StatusBadRequest, "")
			} else {
				
				formatter.JSON(w, http.StatusOK, ord)
			}
		}
	}
}

func getOrderServer1(uuid string, chn chan<- user) {
	var ord_nil =  user {}
	c := NewClient(server1)
	ord, err := c.GetUser(uuid)
	if err != nil {
		chn <- ord_nil
	} else {
		fmt.Println( "Server1: ", ord)
		chn <- ord
	}
}

func getOrderServer2(uuid string, chn chan<- user) {
	var ord_nil = user {}
	c := NewClient(server2)
	ord, err := c.GetUser(uuid)
	if err != nil {
		chn <- ord_nil
	} else {
		fmt.Println( "Server2: ", ord)
		chn <- ord
	}
}

func getOrderServer3(uuid string, chn chan<- user) {
	var ord_nil = user {}
	c := NewClient(server3)
	ord, err := c.GetUser(uuid)
	if err != nil {
		chn <- ord_nil
	} else {
		fmt.Println( "Server3: ", ord)
		chn <- ord
	}
}
func getOrderServer4(uuid string, chn chan<- user) {
	var ord_nil = user {}
	c := NewClient(server4)
	ord, err := c.GetUser(uuid)
	if err != nil {
		chn <- ord_nil
	} else {
		fmt.Println( "Server4: ", ord)
		chn <- ord
	}
}
func getOrderServer5(uuid string, chn chan<- user) {
	var ord_nil = user {}
	c := NewClient(server5)
	ord, err := c.GetUser(uuid)
	if err != nil {
		chn <- ord_nil
	} else {
		fmt.Println( "Server5: ", ord)
		chn <- ord
	}
}

func ErrorWithJSON(w http.ResponseWriter, message string, code int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	fmt.Fprintf(w, "{message: %q}", message)
}



