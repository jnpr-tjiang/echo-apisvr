package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/jnpr-tjiang/echo-apisvr/pkg/config"
	"github.com/labstack/echo"
	"github.com/labstack/gommon/log"
	"github.com/streadway/amqp"
)

var (
	amqpConn    *amqp.Connection
	amqpChannel *amqp.Channel
	cfg         config.AmqpConfiguration
)

func amqpConnect(url string) (*amqp.Connection, *amqp.Channel, error) {
	var (
		amqpConn    *amqp.Connection
		amqpChannel *amqp.Channel
		err         error
		retries     = 10
	)
	for i := 0; i < retries; i++ {
		log.Printf("Connecting to AMQP at %s", url)
		amqpConn, err = amqp.Dial(url)
		if err != nil {
			log.Errorf("Failed to make AMQP connection: %v", err)
			time.Sleep(5 * time.Second)
			continue
		}
		log.Printf("Successfully connected to AMQP at %s", url)

		amqpChannel, err = amqpConn.Channel()
		if err != nil {
			log.Errorf("Failed to open AMQP channel due to error: %s", err)
			amqpConn.Close()
			time.Sleep(5 * time.Second)
			continue
		}
		log.Printf("Successfully created AMQP channel")
		break
	}
	return amqpConn, amqpChannel, err
}

// InitAsyncRPCHandlers initialize AMQP etc
func InitAsyncRPCHandlers() {
	var err error
	cfg := config.GetConfig()
	if cfg == nil {
		panic("config not initialized")
	}
	url := fmt.Sprintf("amqp://%s:%s@%s:%d", cfg.Amqp.User, cfg.Amqp.User, cfg.Amqp.Host, cfg.Amqp.Port)
	if amqpConn, amqpChannel, err = amqpConnect(url); err != nil {
		panic(fmt.Sprintf("Failed to open AMQP connection: %v", err))
	}
	// TODO: add code to close amqp connection
}

func parse(rawurl string) (exchange string, routingKey string, corrID string, err error) {
	if u, err := url.Parse(rawurl); err == nil {
		if params, err := url.ParseQuery(u.RawQuery); err == nil {
			if v, ok := params["exchange"]; ok {
				exchange = v[0]
			}
			if v, ok := params["routing_key"]; ok {
				routingKey = v[0]
			}
			if v, ok := params["correlation_id"]; ok {
				corrID = v[0]
			}
		}
	}
	return exchange, routingKey, corrID, err
}

// ConfigureRPC handles POST:/rpc/configure
func ConfigureRPC(c echo.Context) error {
	log.Infof("ConfigurePRC is called...")
	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(c.Request().Body)
	if err != nil {
		return err
	}

	type reqPayload struct {
		URL string
	}
	payload := reqPayload{}
	err = json.Unmarshal(buf.Bytes(), &payload)
	if err != nil {
		return err
	}

	go func(payload reqPayload) {
		exchange, routingKey, corrID, err := parse(payload.URL)
		if err != nil {
			log.Errorf("Invalid call back URL: %s", payload.URL)
			return
		}

		log.Infof("Publishing result for request (routing_key=%s, corr_id=%s) ...", routingKey, corrID)
		err = amqpChannel.Publish(
			exchange,   // exchange
			routingKey, // routing key
			false,      // mandatory
			false,      // immediate
			amqp.Publishing{
				ContentType:   "application/octet-stream",
				CorrelationId: corrID,
				Body:          []byte("results"),
				DeliveryMode:  amqp.Transient,
			},
		)
		if err != nil {
			log.Errorf("Failed to publish result for request (%s): %v", corrID, err)
			InitAsyncRPCHandlers()
		}
	}(payload)

	return c.String(http.StatusOK, "test")
}
