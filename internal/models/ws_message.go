package models

// WSMessage représente un message envoyé via WebSocket
type WSMessage struct {
	Type string `json:"type"` // Ex: "NEW_TRACK", "PLAYER_JOINED", "SCORE_UPDATE"
	Payload interface{} `json:"payload"` // Les données (peut être n'importe quoi)
}
//PlayerInfo pour envoyer la liste des joueurs
type PlayerInfo struct {
	ID string `json:"id"`
	Username string `json:"username"`
	Score int `json:"score"`
}