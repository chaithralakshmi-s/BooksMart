/*
	Gumball API in Go
	Riak KV
*/

package main

type product struct {
	Title           string 	`json:"title_register"`
	Author			string	`json:"author_register"`
	Image_URL		string	`json:"image_URL_register"`
	Price			string	`json:"price_register"`
	Quantity		string	`json:"quantity_register"`
}