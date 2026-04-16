package game

import "sync"

type Player struct {
    ID       string `json:"id"`
    Username string `json:"username"`
    Score    int    `json:"score"`
}

type Room struct {
    ID           string             `json:"id"`
    Players      map[string]*Player `json:"players"`
    CurrentTrack int                `json:"current_track_id"`
    IsActive     bool               `json:"is_active"`
    Mu           sync.Mutex         // Pour éviter que deux joueurs marquent des points en même temps et fassent bugger le serveur
}

var (
    ActiveRooms = make(map[string]*Room)
    RoomsMu     sync.Mutex
)

func CreateRoom(id string) *Room {
    RoomsMu.Lock()
    defer RoomsMu.Unlock()
    
    room := &Room{
        ID:      id,
        Players: make(map[string]*Player),
        IsActive: true,
    }
    ActiveRooms[id] = room
    return room
}