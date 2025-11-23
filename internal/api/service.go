package api

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"connectrpc.com/connect"
	v1 "github.com/ferrarinobrakes/unofficial-valorant-api/gen"
	"github.com/ferrarinobrakes/unofficial-valorant-api/internal/cache"
	"github.com/ferrarinobrakes/unofficial-valorant-api/internal/db"
	"github.com/ferrarinobrakes/unofficial-valorant-api/internal/protocol"
)

type Service struct {
	tcpServer *protocol.Server
	cache     *cache.Cache
	db        *db.Database
	logger    *zap.SugaredLogger
}

func NewService(tcpServer *protocol.Server, cache *cache.Cache, database *db.Database, logger *zap.SugaredLogger) *Service {
	return &Service{
		tcpServer: tcpServer,
		cache:     cache,
		db:        database,
		logger:    logger,
	}
}

func (s *Service) GetAccount(
	ctx context.Context,
	req *connect.Request[v1.GetAccountRequest],
) (*connect.Response[v1.GetAccountResponse], error) {
	name := req.Msg.Name
	tag := req.Msg.Tag

	s.logger.Infow("get account request", "name", name, "tag", tag)

	cacheKey := cache.MakeKey(name, tag)
	if cached, ok := s.cache.Get(cacheKey); ok {
		s.logger.Debugw("cache hit (memory)", "name", name, "tag", tag)
		accountData := cached.(*v1.AccountData)
		return connect.NewResponse(&v1.GetAccountResponse{
			Status: 200,
			Data:   accountData,
		}), nil
	}

	dbAccount, err := s.db.GetAccountByNameTag(ctx, db.GetAccountByNameTagParams{
		Name: name,
		Tag:  tag,
	})

	if err == nil {
		if time.Since(dbAccount.UpdatedAt) < 1*time.Hour {
			s.logger.Debug("cache hit (database)", "name", name, "tag", tag)

			accountData := &v1.AccountData{
				Puuid:        dbAccount.Puuid,
				Region:       dbAccount.Region,
				AccountLevel: int32(dbAccount.AccountLevel),
				Name:         dbAccount.Name,
				Tag:          dbAccount.Tag,
				Card:         dbAccount.Card,
				Title:        dbAccount.Title,
				UpdatedAt:    dbAccount.UpdatedAt.Format(time.RFC3339),
			}

			s.cache.Set(cacheKey, accountData)

			return connect.NewResponse(&v1.GetAccountResponse{
				Status: 200,
				Data:   accountData,
			}), nil
		}
	}

	s.logger.Debugw("cache miss, resolving via client", "name", name, "tag", tag)

	response, err := s.tcpServer.ResolveAccount(name, tag)
	if err != nil {
		s.logger.Errorw("failed to resolve account", "error", err)
		return connect.NewResponse(&v1.GetAccountResponse{
			Status: 500,
			Error:  fmt.Sprintf("failed to resolve account: %v", err),
		}), nil
	}

	if response.Error != "" {
		s.logger.Errorw("client returned error", "error", response.Error)
		return connect.NewResponse(&v1.GetAccountResponse{
			Status: 500,
			Error:  response.Error,
		}), nil
	}

	now := time.Now().Format(time.RFC3339)
	accountData := &v1.AccountData{
		Puuid:        response.PUUID,
		Region:       response.Region,
		AccountLevel: int32(response.AccountLevel),
		Name:         name,
		Tag:          tag,
		Card:         response.Card,
		Title:        response.Title,
		UpdatedAt:    now,
	}

	s.cache.Set(cacheKey, accountData)

	err = s.db.UpsertAccount(ctx, db.UpsertAccountParams{
		Puuid:        response.PUUID,
		Region:       response.Region,
		AccountLevel: int64(response.AccountLevel),
		Name:         name,
		Tag:          tag,
		Card:         response.Card,
		Title:        response.Title,
	})

	if err != nil {
		s.logger.Warnw("failed to store account in database", "error", err)
	}

	s.logger.Infow("account resolved successfully", "puuid", response.PUUID)

	return connect.NewResponse(&v1.GetAccountResponse{
		Status: 200,
		Data:   accountData,
	}), nil
}
