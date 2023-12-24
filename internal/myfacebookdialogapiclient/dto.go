package myfacebookdialogapiclient

type DialogMessage struct {
	ID   string `json:"id"`
	From string `json:"from"`
	To   string `json:"to"`
	Text string `json:"text"`
}
