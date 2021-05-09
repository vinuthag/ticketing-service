package notfication

import (
	"io/ioutil"
	"net"
	"os"
	"strconv"
	"sync"
	errconst "ticketing-service/constants"
	errors "ticketing-service/error"
	"ticketing-service/logger"
	"ticketing-service/util"
	"time"

	"github.com/streadway/amqp"
	"gopkg.in/matryer/try.v1"
)

//const Constants to store types of events
const (
	ReserveTicketEvent = ".reserve"
	UpdateTicketEvent  = ".updated"
	CancelTicketEvent  = ".deleted"
)

var (
	instance             NotificationManager
	once                 sync.Once
	confSvcEventExchange = util.GetEnv(util.GetProperty(util.RMQ_EVENT_EXCHANGE), util.GetProperty(util.RMQ_EVENT_EXCHANGE))
	host                 = util.GetEnv(util.RMQ_HOST, util.GetProperty(util.RMQ_HOST))
	user                 = util.GetEnv(util.RMQ_USER, util.GetProperty(util.RMQ_USER))
	password             = ""
	pwdPath              = os.Getenv(util.RMQ_SECRET_PATH)
	port                 = util.GetEnv(util.RMQ_PORT, util.GetProperty(util.RMQ_PORT))
	rmqVhost             = util.GetEnv(util.RMQ_VHOST, util.GetProperty(util.RMQ_VHOST))
)

// NotifyRmqEvent ... Structure to store event
type NotifyRmqEvent struct {
	routingKey   string //Rabbitmq routing key
	body         string //Body of event
	exchangeName string //Exchange Name
}
type NotificationManager interface {
	NotifyRMQSubscribers(preferenceName, preferenceEventType, eventDesc string) (int64, error)
	CloseRMQConnection()
}

type notifier struct {
	rmqConn *amqp.Connection
}

func GetNotificationManagerInstance() NotificationManager {
	isRmqEnabled, _ := strconv.ParseBool(util.GetEnv(util.IS_RMQ_ENABLED, util.GetProperty(util.IS_RMQ_ENABLED)))
	if isRmqEnabled {
		once.Do(func() {
			data, err := ioutil.ReadFile(pwdPath)
			if err != nil {
				logger.Log().Error("Error caught while reading the rabbitMQ secret file :", err)
			}
			password = string(data)
			rmqConn, rmqConnError := rmqConnection()
			if rmqConnError == nil {
				instance = &notifier{rmqConn}
			} else {
				instance = &notifier{}
			}
		})
	} else {
		logger.Log().Info("Rabbitmq is disabled, As configured isRmqEnabled : %t", isRmqEnabled)
		instance = &notifier{}
	}
	return instance
}

func failOnError(err error, msg string) {
	if err != nil {
		logger.Log().Errorf("%s: %s", msg, err)
	}
}

//RmqConnection ... Function to get Rabbitmq connection
func rmqConnection() (*amqp.Connection, error) {
	delay, _ := strconv.ParseInt(util.GetEnv(util.RMQ_RETRY_DELAY_MILLISEC, util.GetProperty(util.RMQ_RETRY_DELAY_MILLISEC)), 10, 64)
	maxRetry, _ := strconv.Atoi(util.GetEnv(util.RMQ_RETRY_MAX_ATTEMPTS, util.GetProperty(util.RMQ_RETRY_MAX_ATTEMPTS)))
	connectionTimeOut, _ := strconv.Atoi(util.GetEnv(util.RMQ_CONNECTION_TIME_OUT_MS, util.GetProperty(util.RMQ_CONNECTION_TIME_OUT_MS)))
	dialString := "amqp://" + user + ":" + password + "@" + host + ":" + port + "/" + rmqVhost

	//logger.Log().Debugf("RabbitMQ connection URL : %s", dialString)
	var (
		rmqConn *amqp.Connection
		err     error
	)
	startTime := time.Now()
	sleepTime := time.Duration(delay) * time.Millisecond
	errStatus := try.Do(func(attempt int) (bool, error) {
		rmqConn, err = amqp.DialConfig(dialString, amqp.Config{
			Dial: func(network, addr string) (net.Conn, error) {
				return net.DialTimeout(network, addr, time.Duration(connectionTimeOut)*time.Millisecond)
			},
		})
		if err != nil {
			logger.Log().Debugf("RabbitMQ connection retry attempt :: %d retry delay for :: %s", attempt, sleepTime)
			time.Sleep(sleepTime) // 50 ms
		}
		return attempt < maxRetry, err // try 3 times
	})
	logger.Log().Infof("[TIME TAKEN] to try to establish connection to RabbitMQ :: %s", time.Since(startTime))

	if errStatus != nil {
		failOnError(err, "Failed to connect to RabbitMQ..!")
		err = errors.WrapError(errconst.FAILED_TO_CONNECT_TO_RMQ, err, "Failure in notification publisher")
	} else {
		logger.Log().Info("RMQ Connection Success..!")
	}
	return rmqConn, err
}

func (notifier *notifier) CloseRMQConnection() {
	logger.Log().Infof("Closing RMQ connection")
	if notifier.rmqConn == nil {
		logger.Log().Infof("No connection to RMQ available. Nothing to close")
		return
	}
	err := notifier.rmqConn.Close()
	if err != nil {
		logger.Log().Errorf("Failed to close RMQ connection : %v", err)
	} else {
		logger.Log().Infof("Successfully closed RMQ connection")
	}

}

// NotifyRMQSubscribers ... Function to Notify Rabbitmq Subscribers with event name and description
func (notifier *notifier) NotifyRMQSubscribers(preferenceName, preferenceEventType, eventDesc string) (int64, error) {
	stackMessage := " [NOTIFICATION] Error in : (NotifyRMQSubscribers)"
	startTime := time.Now()
	event := NotifyRmqEvent{
		routingKey:   preferenceName + preferenceEventType,
		body:         eventDesc,
		exchangeName: confSvcEventExchange,
	}
	logger.Log().Infof("Routing Key : %s", preferenceName+preferenceEventType)
	var err error
	if notifier.rmqConn == nil || notifier.rmqConn.IsClosed() {
		notifier.rmqConn, err = rmqConnection()
		if err != nil {
			return -1, errors.NewCustomErr(errconst.FAILED_TO_CONNECT_TO_RMQ, stackMessage)
		}
	}
	status, rmqErr := PublishEvent(event, notifier.rmqConn)
	logger.Log().Debugf("[TIME TAKEN] to send notification to RabbitMQ :: %s", time.Since(startTime))
	return status, rmqErr
}

// PublishEvent ... Function to Publish event to rabbitmq
func PublishEvent(event NotifyRmqEvent, rmqConn *amqp.Connection) (int64, error) {

	channel, err := rmqConn.Channel()
	if err != nil {
		failOnError(err, "Failed to open a channel")
		return -1, err
	}
	defer channel.Close()

	err = channel.ExchangeDeclare(
		event.exchangeName, // name
		"topic",            // type
		true,               // durable
		false,              // auto-deleted
		false,              // internal
		false,              // no-wait
		nil,                // arguments
	)
	if err != nil {
		failOnError(err, "Failed to declare an exchange")
		return -1, err
	}

	eventBody := event.body
	err = channel.Publish(
		event.exchangeName, // exchange
		event.routingKey,   // routing key
		false,              // mandatory
		false,              // immediate
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "text/plain",
			Body:         []byte(eventBody),
		})
	if err != nil {
		failOnError(err, "Failed to publish a message")
		return -1, err
	}

	logger.Log().Debugf(" [x] Sent %s", eventBody)
	return 0, nil
}
