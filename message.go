package main

type ErrorMessage struct {
	Error  string `json:"error"`
	Reason string `json:"reason"`
}

type UserToServerMessage struct {
	Type string `json:"type"`
}

type UserToServerRequest struct {
	UserToServerMessage
	RequestId string `json:"request_id"`
}

type ServerToUserMessage struct {
	Type string `json:"type"`
}

type UsersListMessage struct {
	ServerToUserMessage
	Users []*User `json:"users"`
}

type RoomJoinedMessage struct {
	ServerToUserMessage
	Room *Room `json:"room"`
}

type RoomLeavedMessage struct {
	ServerToUserMessage
	Room  *Room  `json:"room"`
	Cause string `json:"cause"`
}

func NewMessageUsersList(users []*User) UsersListMessage {
	return UsersListMessage{
		ServerToUserMessage: ServerToUserMessage{
			Type: "users_list",
		},
		Users: users,
	}
}

func NewMessageErrorRoomCreation(reason string) ErrorMessage {
	return ErrorMessage{
		Error:  "create_room_failure",
		Reason: reason,
	}
}

func NewMessageRoomJoined(room *Room) RoomJoinedMessage {
	return RoomJoinedMessage{
		ServerToUserMessage: ServerToUserMessage{
			Type: "room_joined",
		},
		Room: room,
	}
}

func NewMessageErrorRoomLeave(reason string) ErrorMessage {
	return ErrorMessage{
		Error:  "leave_room_failure",
		Reason: reason,
	}
}

func NewMessageRoomLeaved(room *Room, cause string) RoomLeavedMessage {
	return RoomLeavedMessage{
		ServerToUserMessage: ServerToUserMessage{
			Type: "room_leaved",
		},
		Room:  room,
		Cause: cause,
	}
}
