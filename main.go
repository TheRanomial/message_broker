package main

import (
	"fmt"
	"sync"
	"time"
)

type Message struct{
	Topic string `json:"topic"`
	Payload interface{} `json:"payload"`
}

type Subscriber struct {
	Channel chan interface{}
	Unsubscribe chan bool
}

type Broker struct{
	subscribers map[string][]*Subscriber
	mutex sync.Mutex
}

func NewBroker() *Broker{
	return &Broker{
		subscribers: make(map[string][]*Subscriber),
	}
}

func (b *Broker) Subscribe(topic string) *Subscriber{
	b.mutex.Lock()
	defer b.mutex.Unlock()

	subscriber:=&Subscriber{
		Channel: make(chan interface{},1),
		Unsubscribe: make(chan bool),
	}

	b.subscribers[topic]=append(b.subscribers[topic], subscriber)

	return subscriber
}

func (b *Broker) Unsubscribe(topic string ,subscriber *Subscriber){
	b.mutex.Lock()
	defer b.mutex.Unlock()

	if subscribers,found:=b.subscribers[topic]; found{
		for i,sub:=range subscribers{
			if sub==subscriber{
				close(sub.Channel)
				b.subscribers[topic]=append(subscribers[:i],subscribers[i+1:]...)
				return
			}
		}
	}
}

func (b *Broker) Publish(topic string,payload interface{}){
	b.mutex.Lock()
	defer b.mutex.Unlock()

	if subscribers,found:=b.subscribers[topic]; found{
		for _,sub:=range subscribers{

			select{
				case sub.Channel<-payload:
				case <-time.After(time.Second):
					fmt.Printf("Subscriber slow. Unsubscribing from topic: %s\n", topic)
					b.Unsubscribe(topic, sub)
			}
		}
	}
}

func main(){

	broker:=NewBroker()

	subscriber:=broker.Subscribe("example")
	subscriber2:=broker.Subscribe("example")

	go func(){
		for{
		select {
		case msg,ok:= <-subscriber.Channel:
			if !ok{
				fmt.Println("Subscriber channel closed.")
     			return
			}
			fmt.Printf("Received %v\n",msg)

		case <-subscriber.Unsubscribe:
			fmt.Println("Unsubscribed.")
     		return

		}
		}
	}()

	go func(){
		for{
		select {
		case msg,ok:= <-subscriber2.Channel:
			if !ok{
				fmt.Println("Subscriber channel closed.")
     			return
			}
			fmt.Printf("Received %v\n",msg)

		case <-subscriber2.Unsubscribe:
			fmt.Println("Unsubscribed.")
     		return
		}
		}
	}()

	broker.Publish("example", "Hello, World!")
	broker.Publish("example", "Welcome to my channel")
	broker.Publish("example", "Hope you enjoyed it")
	
	time.Sleep(5 * time.Second)
	broker.Unsubscribe("example", subscriber)
	broker.Unsubscribe("example", subscriber2)
    broker.Publish("example", "This message won't be received.")
    time.Sleep(5*time.Second)

}

