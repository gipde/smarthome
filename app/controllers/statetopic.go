package controllers

import (
	"github.com/revel/revel"
	"schneidernet/smarthome/app/models/devcom"
)

// Internal Message Consumer
type StateTopic struct {
	Input    chan devcom.DevProto
	Consumer [](chan devcom.DevProto)
}

// Global Topics-Map per useroid
var topics = make(map[uint]*StateTopic)

// register new User for his topic
func register(uoid uint) (chan devcom.DevProto, chan devcom.DevProto) {
	if _, ok := topics[uoid]; !ok {
		// we create a new StateTopic
		topic := StateTopic{
			Input:    make(chan devcom.DevProto),
			Consumer: [](chan devcom.DevProto){},
		}
		topics[uoid] = &topic

		// and start a per user TopicHandler
		topicHandler(&topic)
	}

	usertopic := topics[uoid]
	// we add a consumer
	consumer := make(chan devcom.DevProto)
	usertopic.Consumer = append(usertopic.Consumer, consumer)

	return usertopic.Input, consumer
}

// unregister user from his topic
func unregister(uoid uint, consumer chan devcom.DevProto) {
	usertopic := topics[uoid]

	for i, c := range usertopic.Consumer {
		if c == consumer {
			// send the consumer the Quit Command and close it
			c <- devcom.DevProto{Action: devcom.Quit}
			close(c)

			// remove the Consumer
			usertopic.Consumer = append(usertopic.Consumer[:i], usertopic.Consumer[i+1:]...)
		}
	}

	// if this was the last consumer
	if len(usertopic.Consumer) == 0 {
		// we can send quit to the usertopic and close the usertopic
		usertopic.Input <- devcom.DevProto{Action: devcom.Quit}
		close(usertopic.Input)
		// delete the complete usertopic
		delete(topics, uoid)
	}
}

// the topicHandler
func topicHandler(stateTopic *StateTopic) {
	// start goroutine and loop forever
	go func() {
		for {
			msg := <-stateTopic.Input
			if msg.Action == devcom.Quit {
				// exit loop
				break
			}
			// send to every consumer
			for _, consumer := range stateTopic.Consumer {
				consumer <- msg
			}
		}
	}()
}

// the consumerHandler for every Consumer
func consumerHandler(ws revel.ServerWebSocket, consumer chan devcom.DevProto) {
	//internal Receiver from StateTopic loop forever
	go func() {
		for {
			msg := <-consumer
			// after here, it is possible that WebSocketController is disabled
			if msg.Action == devcom.Quit {
				break
			}

			// send to Websocket
			err := ws.MessageSendJSON(&msg)

			if err != nil {
				break
			}
		}
	}()
}
