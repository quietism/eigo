package main

import (
       "fmt"
       "os"
       "net"
       "bufio"
)

type Client struct {
     sender chan string
     receiver chan string
     pusher *bufio.Writer
     puller *bufio.Reader
}

func (client *Client) Push() {
     for data := range client.sender {
     	 client.pusher.WriteString(data)
	 client.pusher.Flush()
     }    
}

func (client *Client) Pull() {
     for {
     	 line, _ := client.puller.ReadString(' ')
	 client.receiver <- line
     }
}

func (client *Client) Listen() {
     go client.Push()
     go client.Pull()
}

func NewClient(connection net.Conn) *Client {
     pusher := bufio.NewWriter(connection)
     puller := bufio.NewReader(connection)

     client := &Client{
     	    sender: make(chan string),
     	    receiver: make(chan string),
	    pusher: pusher,
	    puller: puller, 
     }

     client.Listen()

     return client
}

type Gui struct {
	client *Client
	pusher chan string
	puller chan string
	scanner *bufio.Scanner
	printer *bufio.Writer
}

func (gui *Gui) Circuit() {
	client := gui.client
	go func() { 
	   for {
		line := gui.scanner.Text()
		gui.puller <- line
	   }
	}()
	
	go func() {
	   for data := range gui.pusher {
		fmt.Println(data)
		gui.printer.WriteString(data)
		gui.printer.Flush()
	   }	
	}()
	
	go func() { for { client.sender <- <- gui.puller } }()
	go func() { for { gui.pusher <- <- client.receiver } }()
}

func NewGui(client *Client) *Gui {
	scanner := bufio.NewScanner(os.Stdin)
	printer := bufio.NewWriter(os.Stdout)

	gui := &Gui{
		client: client,
		pusher: make(chan string),
		puller: make(chan string),
		scanner: scanner,
		printer: printer,
	}	
	return gui
}



func main() {
     connection, _ := net.Dial("tcp", ":6667")
     client := NewClient(connection)
     gui := NewGui(client)
     fmt.Println("New Gui started.")
     gui.Circuit()
}
