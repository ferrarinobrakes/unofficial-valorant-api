package protocol

import (
	"fmt"
	"net"
	"sync"
	"time"

	"go.uber.org/zap"

	v1 "github.com/ferrarinobrakes/unofficial-valorant-api/gen"
	"github.com/google/uuid"
)

type ClientConnection struct {
	ID             string
	Conn           net.Conn
	Version        string
	LCUAvailable   bool
	LastHeartbeat  time.Time
	PendingRequest chan *Request
}
type Request struct {
	ID       string
	GameName string
	GameTag  string
	Response chan *Response
}
type Response struct {
	PUUID        string
	Region       string
	AccountLevel int
	Card         string
	Title        string
	Error        string
}

type Server struct {
	listener net.Listener
	clients  map[string]*ClientConnection
	mu       sync.RWMutex
	logger   *zap.SugaredLogger
	done     chan struct{}
}

func NewServer(port int, logger *zap.SugaredLogger) (*Server, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, fmt.Errorf("failed to start TCP server: %w", err)
	}

	logger.Infow("TCP server started", "port", port)

	return &Server{
		listener: listener,
		clients:  make(map[string]*ClientConnection),
		logger:   logger,
		done:     make(chan struct{}),
	}, nil
}

func (s *Server) Start() error {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			select {
			case <-s.done:
				return nil
			default:
				s.logger.Errorw("failed to accept connection", "error", err)
				continue
			}
		}

		go s.handleClient(conn)
	}
}

func (s *Server) handleClient(conn net.Conn) {
	defer conn.Close()
	s.logger.Infow("new client connection", "remote", conn.RemoteAddr())

	var clientID string
	var client *ClientConnection

	for {
		msg, err := ReadMessage(conn)
		if err != nil {
			s.logger.Warnw("failed to read message", "error", err, "clientID", clientID)
			break
		}

		switch payload := msg.Payload.(type) {
		case *v1.Message_ClientRegister:
			clientID = payload.ClientRegister.ClientId
			s.logger.Infow("client registered", "clientID", clientID, "version", payload.ClientRegister.Version)

			client = &ClientConnection{
				ID:             clientID,
				Conn:           conn,
				Version:        payload.ClientRegister.Version,
				LastHeartbeat:  time.Now(),
				PendingRequest: make(chan *Request, 1),
			}

			s.mu.Lock()
			s.clients[clientID] = client
			s.mu.Unlock()

		case *v1.Message_ClientHeartbeat:
			if client != nil {
				client.LastHeartbeat = time.Now()
				client.LCUAvailable = payload.ClientHeartbeat.LcuAvailable
				s.logger.Debugw("heartbeat received", "clientID", clientID, "lcuAvailable", client.LCUAvailable)
			}

		case *v1.Message_ResolveAccountResponse:
			if client != nil {
				select {
				case req := <-client.PendingRequest:
					req.Response <- &Response{
						PUUID:        payload.ResolveAccountResponse.Puuid,
						Region:       payload.ResolveAccountResponse.Region,
						AccountLevel: int(payload.ResolveAccountResponse.AccountLevel),
						Card:         payload.ResolveAccountResponse.Card,
						Title:        payload.ResolveAccountResponse.Title,
					}
				default:
					s.logger.Warnw("received response with no pending request", "clientID", clientID)
				}
			}

		case *v1.Message_ErrorResponse:
			if client != nil {
				select {
				case req := <-client.PendingRequest:
					req.Response <- &Response{
						Error: payload.ErrorResponse.Message,
					}
				default:
					s.logger.Warnw("received error with no pending request", "clientID", clientID)
				}
			}
		}
	}

	if clientID != "" {
		s.mu.Lock()
		delete(s.clients, clientID)
		s.mu.Unlock()
		s.logger.Infow("client disconnected", "clientID", clientID)
	}
}

func (s *Server) ResolveAccount(gameName, gameTag string) (*Response, error) {
	s.mu.RLock()
	var client *ClientConnection
	for _, c := range s.clients {
		if c.LCUAvailable && time.Since(c.LastHeartbeat) < 30*time.Second {
			client = c
			break
		}
	}
	s.mu.RUnlock()

	if client == nil {
		return nil, fmt.Errorf("no available clients")
	}

	req := &Request{
		ID:       uuid.New().String(),
		GameName: gameName,
		GameTag:  gameTag,
		Response: make(chan *Response, 1),
	}

	msg := &v1.Message{
		Id: req.ID,
		Payload: &v1.Message_ResolveAccountRequest{
			ResolveAccountRequest: &v1.ResolveAccountRequest{
				GameName: gameName,
				GameTag:  gameTag,
			},
		},
	}

	client.PendingRequest <- req

	if err := WriteMessage(client.Conn, msg); err != nil {
		return nil, fmt.Errorf("failed to send request to client: %w", err)
	}

	s.logger.Infow("request sent to client", "clientID", client.ID, "name", gameName, "tag", gameTag)

	select {
	case resp := <-req.Response:
		return resp, nil
	case <-time.After(60 * time.Second):
		return nil, fmt.Errorf("request timed out")
	}
}

func (s *Server) GetClientCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.clients)
}

func (s *Server) Stop() error {
	close(s.done)
	return s.listener.Close()
}
