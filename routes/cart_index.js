var express = require('express');
var router = express.Router();
var axios = require('axios');

/* GET home page. */
//axios api, axios.get /axios.post
router.get('/', function(req, res, next) {
    var cartItems = [
        
    ];
    var req = {"userId":"user200",
    "items":[
        {
        "name":"starbucks",
        "count":1,
        "rate":3.95
        },
        {
        "name":"peets",
        "count":1,
        "rate":4.95
        }
        ]
    };
    /*axios.get('')
    .then(function(response){
        req = response;
    });*/
    axios.post('http://127.0.0.1:3000/order', req)
    .then(function (res){
        //console.log('response',res.data);
        cartItems=res.data.items;
        console.log("in then", cartItems);
    });
    console.log(cartItems);

  res.render('cart/cart', { items: cartItems });
});

router.get('/clearCart', function (req,res,next) {
  console.log("in delete cart");
  res.render('cart/cart', { items: [] });
});

module.exports = router;
