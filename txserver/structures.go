package main

type Command struct {
	Command  string `json:"Command"`
	Username string `json:"Username"`
	Amount   string `json:"Amount"`
	Stock    string `json:"Stock"`
	Filename string `json:"Filename"`
}

type Response struct {
	Data  []byte `json:"data"`
	Error string `json:"error"`
}
