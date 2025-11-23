package protocol

import (
	"fmt"
	"net"
	"time"

	"go.uber.org/zap"

	v1 "github.com/ferrarinobrakes/unofficial-valorant-api/gen"
	"github.com/ferrarinobrakes/unofficial-valorant-api/internal/lcu"
)

type Client struct {
	serverAddress string
	clientID      string
	version       string
	conn          net.Conn
	resolver      *lcu.Resolver
	logger        *zap.SugaredLogger
	done          chan struct{}
}

func NewClient(serverAddress, clientID, version string, resolver *lcu.Resolver, logger *zap.SugaredLogger) *Client {
	return &Client{
		serverAddress: serverAddress,
		clientID:      clientID,
		version:       version,
		resolver:      resolver,
		logger:        logger,
		done:          make(chan struct{}),
	}
}

func (c *Client) Connect() error {
	conn, err := net.Dial("tcp", c.serverAddress)
	if err != nil {
		return fmt.Errorf("failed to connect to master: %w", err)
	}

	c.conn = conn
	c.logger.Infow("connected to master server", "address", c.serverAddress)

	msg := &v1.Message{
		Id: "0",
		Payload: &v1.Message_ClientRegister{
			ClientRegister: &v1.ClientRegister{
				ClientId: c.clientID,
				Version:  c.version,
			},
		},
	}

	if err := WriteMessage(conn, msg); err != nil {
		return fmt.Errorf("failed to send registration: %w", err)
	}

	c.logger.Infow("registration sent")

	lcuAvailable := c.isLCUAvailable()
	initialHeartbeat := &v1.Message{
		Id: "heartbeat-initial",
		Payload: &v1.Message_ClientHeartbeat{
			ClientHeartbeat: &v1.ClientHeartbeat{
				Timestamp:    time.Now().UnixMilli(),
				LcuAvailable: lcuAvailable,
			},
		},
	}

	if err := WriteMessage(conn, initialHeartbeat); err != nil {
		c.logger.Warnw("failed to send initial heartbeat", "error", err)
	} else {
		c.logger.Infow("initial heartbeat sent", "lcuAvailable", lcuAvailable)
	}

	return nil
}

func (c *Client) Run() error {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	messages := make(chan *v1.Message)
	errors := make(chan error)

	go func() {
		for {
			msg, err := ReadMessage(c.conn)
			if err != nil {
				errors <- err
				return
			}
			messages <- msg
		}
	}()

	for {
		select {
		case <-ticker.C:
			lcuAvailable := c.isLCUAvailable()
			msg := &v1.Message{
				Id: "heartbeat",
				Payload: &v1.Message_ClientHeartbeat{
					ClientHeartbeat: &v1.ClientHeartbeat{
						Timestamp:    time.Now().UnixMilli(),
						LcuAvailable: lcuAvailable,
					},
				},
			}

			if err := WriteMessage(c.conn, msg); err != nil {
				c.logger.Errorw("failed to send heartbeat", "error", err)
			}

		case msg := <-messages:
			c.handleMessage(msg)

		case err := <-errors:
			return fmt.Errorf("connection error: %w", err)

		case <-c.done:
			return nil
		}
	}
}

func (c *Client) handleMessage(msg *v1.Message) {
	switch payload := msg.Payload.(type) {
	case *v1.Message_ResolveAccountRequest:
		c.logger.Infow("received resolve request", "name", payload.ResolveAccountRequest.GameName, "tag", payload.ResolveAccountRequest.GameTag)
		go c.handleResolveRequest(msg.Id, payload.ResolveAccountRequest)
	}
}
func (c *Client) handleResolveRequest(requestID string, req *v1.ResolveAccountRequest) {
	account, err := c.resolver.ResolveAccount(req.GameName, req.GameTag)

	var response *v1.Message
	if err != nil {
		c.logger.Errorw("failed to resolve account", "error", err)
		response = &v1.Message{
			Id: requestID,
			Payload: &v1.Message_ErrorResponse{
				ErrorResponse: &v1.ErrorResponse{
					Code:    "RESOLVE_FAILED",
					Message: err.Error(),
				},
			},
		}
	} else {
		c.logger.Infow("account resolved", "puuid", account.PUUID)
		response = &v1.Message{
			Id: requestID,
			Payload: &v1.Message_ResolveAccountResponse{
				ResolveAccountResponse: &v1.ResolveAccountResponse{
					Puuid:        account.PUUID,
					Region:       account.Region,
					AccountLevel: int32(account.AccountLevel),
					Card:         account.Card,
					Title:        account.Title,
				},
			},
		}
	}

	if err := WriteMessage(c.conn, response); err != nil {
		c.logger.Errorw("failed to send response", "error", err)
	}
}

func (c *Client) isLCUAvailable() bool {
	_, err := lcu.ReadLockfile()
	return err == nil
}

func (c *Client) Stop() error {
	close(c.done)
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
