package gamestate

import (
	"github.com/google/uuid"
)

type GameStateSingleton struct {
	games                  map[string]*Game
	WaitingRooms           map[string]*Game
	matchMakingWaitingRoom *Game
	NewMatchMakingRequests chan MatchMakingRequest
	NewWaitingRoomRequests chan WaitingRoomRequest
	GameUpdateRequests     chan GameUpdateRequest
}

func Init() GameStateSingleton {
	return GameStateSingleton{
		games:                  make(map[string]*Game),
		NewMatchMakingRequests: make(chan MatchMakingRequest),
		NewWaitingRoomRequests: make(chan WaitingRoomRequest),
		GameUpdateRequests:     make(chan GameUpdateRequest),
	}
}

type MatchMakingRequest struct {
	Res chan GameResponse
}

type WaitingRoomRequest struct {
	res chan GameResponse
}

type FriendJoinRequest struct {
	res chan FriendJoinResponse
}

type GameResponse struct {
	PlayerID string
	G        Game
	Ready    bool
}

type FriendJoinResponse struct {
	G     Game
	Found bool
}

type GameUpdateRequest struct {
	Ch       chan Game
	ID       string
	PlayerID string
	A        Action
}

func (gss *GameStateSingleton) StartProcessing() {
	for {
		select {
		case req := <-gss.NewMatchMakingRequests:
			if gss.matchMakingWaitingRoom == nil {
				id, err := uuid.NewV6()
				if err != nil {
					panic(err)
				}
				g := NewGame(id.String())
				leftPlayerID, err := uuid.NewV6()
				if err != nil {
					panic(err)
				}
				g.LeftPlayerID = leftPlayerID.String()
				req.Res <- GameResponse{
					G:        *g,
					Ready:    false,
					PlayerID: g.LeftPlayerID,
				}
				gss.matchMakingWaitingRoom = g
			} else {
				rightPlayerID, err := uuid.NewV6()
				if err != nil {
					panic(err)
				}
				rightPlayerIDString := rightPlayerID.String()
				gss.matchMakingWaitingRoom.RightPlayerID = rightPlayerIDString
				gss.matchMakingWaitingRoom.start()
				req.Res <- GameResponse{
					G:        *gss.matchMakingWaitingRoom,
					Ready:    true,
					PlayerID: rightPlayerIDString,
				}
				gss.matchMakingWaitingRoom.start()
				gss.games[gss.matchMakingWaitingRoom.ID] = gss.matchMakingWaitingRoom
				gss.matchMakingWaitingRoom = nil
			}
		//case req := <-gss.NewWaitingRoomRequests:

		//case req := <-gss.FriendJoinRequests:

		case req := <-gss.GameUpdateRequests:
			g := gss.games[req.ID]
			g.play(req.A, req.PlayerID)
			req.Ch <- *g
		}
	}
}
