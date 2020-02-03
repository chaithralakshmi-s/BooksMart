Multi-User Shopping Cart – E-Commerce Application

In this project, we are going to develop a multi-user Shopping cart E-Commerce application. In this application, we are planning to have several types of functionalities under customer services and administrative services. Also, each team member will be implementing one of the different functionalities by maintaining one independent database for maintaining Users, Products, Sellers, Cart and Payment.  This application will be built to a cloud scale SaaS Service on Amazon Cloud and Heroku using the basic scaling design approaches.


Technicalities
•	NoSQL Database : Redis
•	Programming Language : Go Programming, (to be updated)
•	AWS EC2 Instance
•	Heroku


API:Customer Identification

•	A user profile will be set for each customer. 
•	A customer will be allowed to create a profile and will henceforth be authenticated by a login to his user profile which will also enable him to view the order history. 
•	There will also be a Super User who allows creation and deletion of other users on a universal shopping cart.
•	This will be a Customer identification API which handles the REST API calls to create profile, user authentication and viewing order history for authenticated users.

API: Browse Products from product catalog

•	A user is allowed to browse the various products from the product catalog. 
•	This API will be developed allowing the users to obtain all the products from the product catalog
•	Also allows the users to select a particular category from the ones available in the  product catalog. 

API:Shopping Cart

•	Every logged in user will be able to add , delete , checkout and make payments for the different items chosen from the product catalog. 
•	The Super user will also be able to add other users who can add products on to the universal shopping cart from the product catalog.

API: Add/ Delete Product Categories

•	This API will include REST API Calls to allow additions and deletions of different categories into the Product Catalog.


API: Add/ Delete Products

•	This API will include REST API Calls which enable adding, delteing, modifying products in the product catalog.






