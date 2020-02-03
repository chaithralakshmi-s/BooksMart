package main

//Transaction struct for transaction history
type Transaction struct {
	TransactionId	string
	PaymentType 	string
	UserDetails	User
	Amount		string
	Status		string
}

//Payment type used to extract the data from the request
type Payment struct {
	UserId		string
	PaymentType	string
	Name		string
	UsernameId	string
	Password	string
	Amount		float64
}

//User specific data
type User struct {
	Name 		string
	Id 		string
	Password 	string
}