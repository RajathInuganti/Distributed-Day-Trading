package main

type Command struct {
	Command  string  `json:"Command"`
	Username string  `json:"Username"`
	Amount   float32 `json:"Amount"`
	Stock    string  `json:"Stock"`
	Filename string  `json:"Filename"`
}

type Response struct {
	Data  []byte `json:"data"`
	Error string `json:"error"`
}

type ParsingErrors struct {
	usernameEmpty               bool
	stockSymbolEmpty            bool
	AmountNotConvertibleToFloat bool
}

type UserAccount struct {
	username string             `bson:"username"`
	balance  float32            `bson:"balance"`
	created  int                `bson:"created"`
	updated  int                `bson:"updated"`
	buy      map[string]float32 `bson:"buy"`
	sell     map[string]float32 `bson:"sell"`
	stocks   map[string]float32 `bson:"stocks"`
}

type Transaction struct {
	TransactionNum int     `bson:"transactionNum"`
	Username       string  `bson:"username"`
	Stock          string  `bson:"stock"`
	Amount         float32 `bson:"amount"`
	command        string  `bson:"command"`
}