var express = require('express');
var axios = require('axios');
var path = require('path');
var app = express();
var XMLHttpRequest = require("xmlhttprequest").XMLHttpRequest;
var bodyParser = require('body-parser');
var csurf = require('csurf');
var session = require('express-session');
var cookieParser = require('cookie-parser');
var Client = require('node-rest-client').Client;

var favicon = require('serve-favicon');
var logger = require('morgan');
var del_client = require('request');

//var index = require('./routes/cart_index');
//var users = require('./routes/cart_users');

app.set('port', (process.env.PORT || 5000));

app.use(bodyParser.urlencoded({ extended: true }));
app.use(bodyParser.json());
app.use(cookieParser());
app.use(session({
	secret: 'mysupersecret', 
	resave: false, 
	saveUninitialiazed: false,
	cookie: { maxAge: 180 * 60 * 1000} //in milliseconds 
}));

app.use(express.static(__dirname + '/public'));

// views is directory for all template files
app.set('views', __dirname + '/views');
app.set('view engine', 'ejs');

var productCatalogServer = "http://productapi-elb-2141657141.us-west-1.elb.amazonaws.com:80/";
var userloginServer = "http://lb-userapi-1276859448.us-west-1.elb.amazonaws.com:80/"
var shoppingCartServer= "http://ecs-cali-1807036812.us-west-1.elb.amazonaws.com:80/"
var UserOrderHistoryServer = "http://projectLoadBalancer-1857260383.us-west-1.elb.amazonaws.com:3000/";
var CheckoutPaymentServer = "http://api-load-balancer-173802822.us-west-1.elb.amazonaws.com:4500/"

var isLoggedIn = false;
var cartQuantity = 0;
var products = Object();

products.items = [];
products.total = 0;

app.post('/getusertransactiondetails', function (req, res) {
  //User_id missing in input
  if (!req.session.userid) {
    var userhistoryerror = "User id is missing";
       res.render('pages/userhistoryerror', {
      userhistoryerror: userhistoryerror,
      login: isLoggedIn,
      cartQuantity: cartQuantity
    });
  }
  //Transaction_id missing in input
  else if (!req.body.transaction_id) {
    // res.status(500); 
    req.body.transaction_id = "null";
    var userhistoryerror = "Transaction id is missing";
     res.render('pages/userhistoryerror', {
      userhistoryerror: userhistoryerror,
      login: isLoggedIn,
      cartQuantity: cartQuantity
    });
  }


//All required inputs have been received
  else {

    var client = new Client();

    var url = UserOrderHistoryServer + "getTransactionDetailsHandler/" + req.session.userid + "/" + req.body.transaction_id;
    client.get(url, function (data, response) {

      if (!data) {
        var userhistoryerror = "Oops! Looks like something went wrong. Please try later!";

        res.render('pages/userhistoryerror', {
          userhistoryerror: userhistoryerror,
	  login: isLoggedIn,
	  cartQuantity: cartQuantity
        });
      }
    
      var response = JSON.stringify(data);	

      //User Id / Transaction details not found in Riak database
      if (response == "[]") {
        var userhistoryerror = "User id - " + req.session.userid + " not found";

        res.render('pages/userhistoryerror', {
          userhistoryerror: userhistoryerror,
	  login: isLoggedIn,
	  cartQuantity: cartQuantity
        });
      }

      //No error.Prepare output for html
     
      var arr_parse = JSON.parse(response);

      //send to html
      var userid = req.session.userid;
      var transactionid = arr_parse.uTransactionId;
      var transactionddate = arr_parse.uTransactionsDate;
      var transactiontotal = arr_parse.uTransactiontotal;
      var transactionitems = arr_parse.uTransactionItems;

      
      //Render
      res.render('pages/usertransactiondetails', {
        userid: userid, transactionid: transactionid,
        transactiondate: transactionddate,
        transactiontotal: transactiontotal,
        arr: transactionitems,
	login: isLoggedIn,
	cartQuantity: cartQuantity
      });




    });
 }
})
app.get('/gettransactions', function (req, res) {

  var userhistoryerror = "";

  //If user_id is empty, send error message
  if (!req.session.userid) {
    var userhistoryerror = "User id is Missing. Please enter a user id";
    res.render('pages/userhistoryerror', {
      userhistoryerror: userhistoryerror,
      login: isLoggedIn,
      cartQuantity: cartQuantity
    }); 
  }

  
  
  //All required inputs have been received
  else {
    var client = new Client();
    var url = UserOrderHistoryServer + "getusertransactions/" + req.session.userid;
    
  
    client.get(url, function (data, response) {
      
      //Unexpected error 
      if (!data) {
        var userhistoryerror = "Oops! Looks like something went wrong. Please try later!";

        res.render('pages/userhistoryerror', {
          userhistoryerror: userhistoryerror
        });
	return;
      }
      
     var response = JSON.stringify(data);

      //No Transactions found or User Id not found
      if (response == "[]") {
        var userhistoryerror = "No Transactions found or User Id not found";

        res.render('pages/userhistoryerror', {
          userhistoryerror: userhistoryerror,
	  login: isLoggedIn,
	  cartQuantity: cartQuantity
        });
	return;
      }

      //Render the page
      res.render('pages/usertransactions', {
        userid: req.session.userid,
        arr: JSON.parse(response),
	login: isLoggedIn,
	cartQuantity: cartQuantity
      });
    });
  }
});

app.get('/checkout', function(request, response){
		
		if (request.session.userid) {
			//get products
			var client = new Client();
			var url = shoppingCartServer + "history/" + request.session.userid;
			var read = client.get(url, function(data, get_resp) {
				var obj_data = data[0];
				var isLoggedIn = true;
				response.render('shop/checkout', {login: isLoggedIn, totalAmount: obj_data.total, cartQuantity: cartQuantity, products: obj_data.items});
			});
			
			read.on('error', function(err) {
				response.redirect(req.get('referer'));
			});
		}
		else {
			response.redirect('/signin');
		}
});

app.post('/checkout', function(request, response){
	
	var product_list = JSON.parse(request.body.products);
	
	if (request.body.pay === "Cancel") {
		response.redirect("/");
		return
	}
	
	var client = new Client();
	var url = CheckoutPaymentServer + "transaction";
	var args = {
		data: { "UserId": request.body.userid,
			"PaymentType": "Card", 
			"Name": request.body.card_name, 
			"UsernameId": request.body.card_number, 
			"Password": request.body.card_cvv, 
			"Amount": parseFloat(request.body.amount)
		},
		headers: { "Content-Type": "application/json"}
	};
	
	var send_post = client.post(url,args,function(data,post_response){
		var processed_data = data;
		var history_url = UserOrderHistoryServer + "addtransaction/" + request.session.userid;
		var items = [];
		
		
		for (i = 0; i < product_list.length; i++) {
				var total = product_list[i].count * product_list[i].rate;
				items.push({
					"product": product_list[i].name,
					"quantity": product_list[i].count.toString(),
					"rate": product_list[i].rate.toString(),
					"total": total.toString() 
					}
				);
		}
		var uh_args = {
			data: {
				"userid": request.session.userid,
				"utransactionid": processed_data.Id,
				"utransactionitems": items,
				"utransactiontotal": request.body.amount.toString()
			},
			headers: { "Content-Type": "application/json"}
		}
		
		try {
		var hist_post = client.post(history_url, uh_args, function(data, history_resp) {
				cartQuantity = 0;
				products.items = [];
				products.total = 0;
				
				var url = shoppingCartServer + "clearCart/" + request.session.userid;
				client.delete(url, function(data, resp) {
						response.redirect("/");
				});
		});
		}catch(e) {
			response.redirect("/");
		}
		
		hist_post.on("error", function(err) {
				response.redirect(request.get('referer'));
		});
		
	});
	
	send_post.on('error', function(err) {
		response.redirect(request.get('referer'));
	});
	
});

app.get('/add-to-cart/:id', function(request, response) {
	
	var productId = request.params["id"];
	if (!request.session.userid)
	{
	  	cartQuantity = 0;
	 	response.render('user/signin', {login: isLoggedIn, cartQuantity: cartQuantity});
		return;
	}
	else 
	{
	    var xmlhttp = new XMLHttpRequest();
	    xmlhttp.open("GET", productCatalogServer + "products/"+productId);  
	    xmlhttp.setRequestHeader("Content-Type", "application/json");
	    xmlhttp.send();
	    xmlhttp.onreadystatechange = function() 
	    {
	    	if (this.readyState === 4 && this.status === 200) 
	        {
			var product = JSON.parse(this.responseText);
			var isFound = false;
			for(i = 0; i < products.items.length; i++)
			{	
				if (products.items[i].name === product.title_register)
				{
						isFound = true;
						products.items[i].count++;
						products.total += parseFloat(product.price_register);
						break;
				}
			}
			
			if (isFound == false) {
				products.total += parseFloat(product.price_register);
				products.items.push({"name": product.title_register, "count":1, "rate": parseFloat(product.price_register)});
			}
			
			cartQuantity++;
			
			return response.redirect(request.get('referer'));
		}
	   }
	}
});

app.get('/post-cart', function(request, response) {
		var client = new Client();
		var args = {
			data: { "userId": "" + request.session.userid, 
				"items": products.items
			},
			headers: { "Content-Type": "application/json"}
		};
		
		var url = shoppingCartServer + "order/" + request.session.userid;
					
		var send_post = client.post(url, args, function(data, resp) {
				response.redirect("/checkout");
		});
		
		send_post.on('error', function(err) {
			response.redirect(request.get('referer'));
		});
		
});	
	
app.get('/cart', function (request,response) {
	try{
	    if (request.session.userid === "")
	    {
	    	cartQuantity = 0;
	    	//return response.render('user/signin', {login: isLoggedIn, cartQuantity: cartQuantity});
		return response.redirect("/");
	    }     
	    else
	    {	    	
		response.render('pages/cart', {total: products.total, user: request.session.userid, items: products.items, login: isLoggedIn, cartQuantity: cartQuantity});
	    }
	}
	catch(e) {
	    	//Display alert box and redirect to signin page
	    	if(e.name == "TypeError")
	    	    response.render('user/signin', {login: isLoggedIn, cartQuantity: cartQuantity});
	    
	}
	
});

app.get('/signup', function(request, response) {
  response.render('user/signup', {login: isLoggedIn, cartQuantity: cartQuantity});
});

app.post('/signup', function(request, response) {
	if(request.body.submit == "Cancel") {
		response.redirect('/');
		return;
	}
	
	var xmlhttp = new XMLHttpRequest();
	xmlhttp.open("POST", userloginServer + "user");
	xmlhttp.setRequestHeader("Content-Type", "application/json");
	var temp_userId = request.body.inputUsername;
	var jsonToSend = {
		"UserId": request.body.inputUsername,
		"Name":  request.body.name,
		"Email": request.body.inputPassword
	};
	xmlhttp.send(JSON.stringify(jsonToSend));
	    xmlhttp.onreadystatechange = function() 
	    {
   		if (this.readyState === 4 && this.status === 200) 
		{
			isLoggedIn = true;
			request.session.userid = request.body.inputUsername;
			var products_array = JSON.parse(this.responseText);
			cartQuantity = 0;
			response.redirect("/")
			//response.render('./pages/product_catalog', {products: products_array, login: isLoggedIn, cartQuantity: cartQuantity});
		}
		else if (this.readyState === 4 && this.status !== 200)
		{
			console.log("Cannot post to user database");
			response.redirect("/");
		}
	}	
});

app.get('/signin', function(request, response) { 
	if (request.session.userid) {
		return response.redirect("/");
	}
	response.render('user/signin', {login: isLoggedIn, cartQuantity: cartQuantity});
});

app.post('/signin', function(request, response) {
		inputID = parseInt(request.body.inputUsername);
		var xmlhttp = new XMLHttpRequest(); 
		xmlhttp.open("GET", userloginServer+ "user/" +inputID);  
		xmlhttp.setRequestHeader("Content-Type", "application/json");
		xmlhttp.send();
		xmlhttp.onreadystatechange = function() 
		{
			if (this.readyState === 4 && this.status === 200) 
			{
				var responseText = JSON.parse(this.responseText);
				if (responseText.UserId == inputID)
				{
					var xmlhttp1 = new XMLHttpRequest(); 
					xmlhttp1.open("GET", productCatalogServer+ "products");  
					xmlhttp1.setRequestHeader("Content-Type", "application/json");
					xmlhttp1.send();
					xmlhttp1.onreadystatechange = function() 
					{
						if (this.readyState === 4 && this.status === 200) 
						{
							var client = new Client();
							var prods = JSON.parse(this.responseText);
							var get_url = shoppingCartServer + "history/" + request.body.inputUsername;
							var del_url = shoppingCartServer + "clearCart/" + request.body.inputUsername;
							client.get(get_url, function(data, resp) {
									if(data[0]){
										for (i = 0; i < data[0].items.length; i++) {
											products.items.push({
												"name": data[0].items[i].name,
												"count": data[0].items[i].count,
												"rate": data[0].items[i].rate
											});
											cartQuantity += data[0].items[i].count;
										}
										products.total = data[0].total;
									}
									var products_array = prods;
									isLoggedIn = true;
									request.session.userid = request.body.inputUsername;
									var delete_cart = client.delete(del_url, function(data, resp) {
										return response.render('./pages/product_catalog', {products: products_array, login: isLoggedIn, cartQuantity: cartQuantity});	
									});
									
									delete_cart.on("error", function(err) {
											console.log(err);
											request.session.destroy();
											isLoggedIn = false;
											response.redirect("/");
									});
									
							});
						}
					}
				}
				else {
					response.redirect(request.get('referer'));
				}
			}
			else if (this.readyState === 4 && this.status !== 200) {
				console.log("Cannot get from the user database")
				response.redirect(request.get('referer'));
			}
		}
		
});

app.get('/', function(request, response){
	if (request.session.userid) {
		isLoggedIn = true;
	}
	else {
		isLoggedIn = false;
	}
	var xmlhttp = new XMLHttpRequest(); 
	xmlhttp.open("GET", productCatalogServer+ "products");  
	xmlhttp.setRequestHeader("Content-Type", "application/json");
	xmlhttp.send();
	xmlhttp.onreadystatechange = function() {
		if (this.readyState === 4 && this.status === 200) {
			var products_array = JSON.parse(this.responseText);
			response.render('./pages/product_catalog', {products: products_array, login: isLoggedIn, cartQuantity: cartQuantity});
		}
	}
});

app.get('/logout', function(request, response){
	
	if (cartQuantity > 0) {
		var client = new Client();
		
		var args = {
			data: {
				"userId": request.session.userid,
				"items": products.items
			},
			headers: {"Content-Type": "application/json"}	
		};
		var url = shoppingCartServer + "order/" + request.session.userid;
		
		var post = client.post(url,args,function(data, resp) {
			isLoggedIn = false;
			request.session.destroy();
			cartQuantity = 0;
			products.items = [];
			products.total = 0;
			response.redirect("/");
		});
		
		post.on("error",function(err) {
				return response.redirect(request.get('referer'));
		});
	}
	else {
		isLoggedIn = false;
		request.session.destroy();
		cartQuantity = 0;
		products.items = [];
		products.total = 0;
		response.redirect("/");
	}
});

app.listen(process.env.PORT || 5000, function() {
  console.log('Node app is running on port ' + app.get('port'));
});