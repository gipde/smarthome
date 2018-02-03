package controllers

import (
	"github.com/revel/revel"
	"schneidernet/smarthome/app"
	"schneidernet/smarthome/app/models/devcom"
	"sync"
	"time"
)

var log = revel.RootLog.New("module", "state")

// Internal Message Consumer per User
// for notifying all connected clients
type Consumer struct {
	ID    string
	Input chan devcom.DevProto
}

type StateTopic struct {
	Input         chan devcom.DevProto
	Consumer      []Consumer
	TopicMutex    *sync.Mutex
	ConsumerMutex *sync.Mutex
}

// Global Topics-Map per useroid
var topics = make(map[uint]*StateTopic)

func init() {
	revel.AppLog.Debug("Init")
	revel.OnAppStart(consumerPing)
}

// pinging all consumer every 5 min
func consumerPing() {
	go func() {
		for {
			time.Sleep(300 * time.Second)
			log.Info("sending Ping to all consumer...")

			for user := range topics {
				notifyAlLConsumer(user, &devcom.DevProto{
					Action: devcom.Ping,
				})
			}
		}

	}()
}

// register new User for his topic
func register(uoid uint, ip string) (chan devcom.DevProto, Consumer) {
	if _, ok := topics[uoid]; !ok {
		// we create a new StateTopic
		topic := StateTopic{
			Input:         make(chan devcom.DevProto),
			Consumer:      []Consumer{},
			TopicMutex:    &sync.Mutex{},
			ConsumerMutex: &sync.Mutex{},
		}
		topics[uoid] = &topic

		// and start a per user TopicHandler
		topicHandler(&topic)
	}

	usertopic := topics[uoid]
	// we add a consumer
	consumer := Consumer{
		ID:    ip + "-" + app.NewUUID(),
		Input: make(chan devcom.DevProto),
	}
	usertopic.Consumer = append(usertopic.Consumer, consumer)
	log.Debug("register user for topic", "uioid", uoid, "consumer", consumer)

	return usertopic.Input, consumer
}

// unregister user from his topic
func unregister(uoid uint, consumer Consumer) {
	log.Debug("Unregister", "user", uoid, "consumer", consumer)
	usertopic := topics[uoid]
	if usertopic == nil {
		log.Error("want to unregister, but consumers and complete topic is already cleared")
		return
	}

	for i, c := range usertopic.Consumer {
		if c == consumer {
			usertopic.ConsumerMutex.Lock()
			// remove the Consumer
			usertopic.Consumer = append(usertopic.Consumer[:i], usertopic.Consumer[i+1:]...)

			// close the Consumer
			close(c.Input)
			usertopic.ConsumerMutex.Unlock()
		}
	}

	// if this was the last consumer
	if len(usertopic.Consumer) == 0 {
		log.Debug("last Consumer -> we remove topic", "user", uoid, "topic", usertopic.Input)
		usertopic.TopicMutex.Lock()
		// delete the complete usertopic
		delete(topics, uoid)
		// close topichandler goroutine
		close(usertopic.Input)
		usertopic.TopicMutex.Unlock()
	}
}

// the topicHandler
func topicHandler(stateTopic *StateTopic) {
	// start goroutine and loop forever
	go func() {
		for {
			msg, more := <-stateTopic.Input
			if more {
				log.Debug("we got a msg into topic", "user", stateTopic.Input, "msg", msg)

				// send to every consumer
				stateTopic.ConsumerMutex.Lock()
				for _, consumer := range stateTopic.Consumer {
					consumer.Input <- msg
				}
				stateTopic.ConsumerMutex.Unlock()
			} else {
				//closed
				break
			}
		}
		log.Debug("we quit topicHandler", "consumer", stateTopic.Consumer)
	}()
}

// the consumerHandler for every Consumer
// a consumer is a client (e.g. browser or a device-connector)
func consumerHandler(ws revel.ServerWebSocket, consumer Consumer, uid uint) {
	//internal Receiver from StateTopic loop forever
	go func() {
		log.Debug("we start a new Consumer goroutine: " + consumer.ID)
		for {
			msg, more := <-consumer.Input
			if more {
				log.Debug("we got a msg ", "consumer", consumer, "msg", msg)

				// send to Websocket
				err := ws.MessageSendJSON(&msg)

				if err != nil {
					// unregister only in websocket-method
					log.Debug("Websocket closed", "consumer", consumer)
				}
			} else {
				//closed
				break
			}
		}
		log.Debug("we quit the consumer", "consumer", consumer)
	}()
}

func notifyAlLConsumer(uid uint, msg *devcom.DevProto) {
	// check if topic exists -> because it could be cleared from a consumer handler
	if usertopic := topics[uid]; usertopic != nil {
		log.Debug("we notify any consumer", "user", uid, "msg", msg)
		usertopic.TopicMutex.Lock()
		if topics[uid] != nil {
			topics[uid].Input <- *msg
		} else {
			log.Error("want to inform consumers, but topic already closed")
		}
		usertopic.TopicMutex.Unlock()
	}
}
