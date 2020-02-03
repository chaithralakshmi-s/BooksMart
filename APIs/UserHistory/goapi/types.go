package main


type UserTransactionInput struct {
	UserName      string  `json:"user"`
	TransactionId string `json:"transactionid"`
	Products[] string `json:"products"`
	Amount string `json:"amount"`
}

type UserTransactions struct {
	UserName      string
	TransactionId string
	Products[] string
	Amount string
	TransactionDate string
}

type ProductsSet struct {
	Products[] string
}

type UserTransactionIds struct {
	UserName      string
	TransactionId string
	TransactionDate string
}










