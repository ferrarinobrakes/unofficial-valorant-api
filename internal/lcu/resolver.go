package lcu

import (
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/ferrarinobrakes/unofficial-valorant-api/internal/valorant"
)

type AccountData struct {
	PUUID        string
	Region       string
	AccountLevel int
	Name         string
	Tag          string
	Card         string
	Title        string
}

type Resolver struct {
	lcuClient *Client
	valClient *valorant.Client
	logger    *zap.SugaredLogger
}

func NewResolver(lcuClient *Client, valClient *valorant.Client, logger *zap.SugaredLogger) *Resolver {
	return &Resolver{
		lcuClient: lcuClient,
		valClient: valClient,
		logger:    logger,
	}
}

func (r *Resolver) ResolveAccount(gameName, gameTag string) (*AccountData, error) {
	r.logger.Infow("resolving account", "name", gameName, "tag", gameTag)

	entitlements, err := r.lcuClient.GetEntitlementsToken()
	if err != nil {
		return nil, fmt.Errorf("failed to get entitlements token: %w", err)
	}

	err = r.lcuClient.SendFriendRequest(gameName, gameTag)
	if err != nil {
		return nil, fmt.Errorf("failed to send friend request: %w", err)
	}

	var friendReq *FriendRequest
	maxAttempts := 10
	for i := 0; i < maxAttempts; i++ {
		time.Sleep(500 * time.Millisecond)

		requests, err := r.lcuClient.GetFriendRequests()
		if err != nil {
			r.logger.Warn("failed to get friend requests", "error", err)
			continue
		}

		for j := range requests.Requests {
			req := &requests.Requests[j]
			if req.GameName == gameName && req.GameTag == gameTag && req.Subscription == "pending_out" {
				friendReq = req
				break
			}
		}

		if friendReq != nil {
			break
		}
	}

	if friendReq == nil {
		return nil, fmt.Errorf("friend request not found after %d attempts", maxAttempts)
	}

	r.logger.Debugw("found friend request", "puuid", friendReq.PUUID, "region", friendReq.Region)

	defer func() {
		err := r.lcuClient.DeleteFriendRequest(friendReq.PUUID)
		if err != nil {
			r.logger.Warnw("failed to delete friend request", "puuid", friendReq.PUUID, "error", err)
		}
	}()

	shard := valorant.RegionToShard(friendReq.Region)
	r.logger.Debugw("mapped region to shard", "region", friendReq.Region, "shard", shard)

	matchHistory, err := r.valClient.GetMatchHistory(shard, friendReq.PUUID, entitlements.AccessToken, entitlements.Token)
	if err != nil {
		return nil, fmt.Errorf("failed to get match history: %w", err)
	}

	if len(matchHistory.History) == 0 {
		return nil, fmt.Errorf("no match history found for player")
	}

	matchID := matchHistory.History[0].MatchID
	r.logger.Debugw("fetching match details", "matchID", matchID)
	matchDetails, err := r.valClient.GetMatchDetails(shard, matchID, entitlements.AccessToken, entitlements.Token)
	if err != nil {
		return nil, fmt.Errorf("failed to get match details: %w", err)
	}

	player := matchDetails.FindPlayerByPUUID(friendReq.PUUID)
	if player == nil {
		return nil, fmt.Errorf("player not found in match details")
	}

	r.logger.Infow("account resolved successfully", "puuid", friendReq.PUUID)

	return &AccountData{
		PUUID:        friendReq.PUUID,
		Region:       friendReq.Region,
		AccountLevel: player.AccountLevel,
		Name:         gameName,
		Tag:          gameTag,
		Card:         player.PlayerCard,
		Title:        player.PlayerTitle,
	}, nil
}
