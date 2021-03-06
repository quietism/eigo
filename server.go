package main

import (
        "fmt"
        "net"
	"bufio"
)

type Client struct {
     incoming chan string
     outgoing chan string
     reader   *bufio.Reader
     writer   *bufio.Writer
}

func (client *Client) Read() {
     for {
	line, _ := client.reader.ReadString('\n')
	client.incoming <- line
     }
}

func (client *Client) Write() {
    for data := range client.outgoing {
	  client.writer.WriteString(data)
	  client.writer.Flush()
    }      
}

func (client *Client) Listen() {
	go client.Read()
	go client.Write()
}

func NewClient(connection net.Conn) *Client {
	reader := bufio.NewReader(connection)
	writer := bufio.NewWriter(connection)

	client := &Client{
		incoming: make(chan string),
		outgoing: make(chan string),
		reader: reader,
		writer: writer,
	}

	client.Listen()

	return client
}

type ChatRoom struct {
	clients []*Client
	joins chan net.Conn
	incoming chan string // rel server perspective
	outgoing chan string // rel server perspective
}

func (chatRoom *ChatRoom) Broadcast(data string) {
	for _, client := range chatRoom.clients {
		client.outgoing <- data
	}
}

func (chatRoom *ChatRoom) Join(connection net.Conn) {
     client := NewClient(connection)
     chatRoom.clients = append(chatRoom.clients, client)
     go func() { for { chatRoom.incoming <- <- client.incoming } }()
}

func (chatRoom *ChatRoom) Listen() {
	go func() {
		for {
			select {
				case data := <- chatRoom.incoming:
					chatRoom.Broadcast(data)
					fmt.Print(data)
				case conn := <- chatRoom.joins:
					chatRoom.Join(conn)
					chatRoom.Broadcast("Welcome to the chatroom!")
					fmt.Println("New person joined!")
			}
		}		
	}()
}

func NewChatRoom() *ChatRoom {
	chatRoom := &ChatRoom{
		clients: make([]*Client, 0),
		joins: make(chan net.Conn),
		incoming: make(chan string),
		outgoing: make(chan string),
	}

	chatRoom.Listen()
	
	return chatRoom

}

func main() {
     	chatRoom := NewChatRoom()
     	fmt.Println("Starting Chat Server")
     	listener, _ := net.Listen("tcp", ":6667")

	for { 
		conn, _ := listener.Accept()
		chatRoom.joins <- conn
	}
}
