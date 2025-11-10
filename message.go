package main

import "encoding/json"

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

type UsersListReply struct {
	ServerToUserMessage
	Users []*User `json:"users"`
}

func NewReplyUsersList(users []*User) UsersListReply {
	return UsersListReply{
		ServerToUserMessage: ServerToUserMessage{
			Type: "users_list",
		},
		Users: users,
	}
}

func NewReplyErrorRoomCreate(reason string) ErrorMessage {
	return ErrorMessage{
		Error:  "create_room_failure",
		Reason: reason,
	}
}

type RoomLeaveReply struct {
	ServerToUserMessage
	Room  *Room  `json:"room"`
	Cause string `json:"cause"`
}

func NewReplyErrorRoomLeave(reason string) ErrorMessage {
	return ErrorMessage{
		Error:  "leave_room_failure",
		Reason: reason,
	}
}

func NewReplyRoomLeaved(room *Room, cause string) RoomLeaveReply {
	return RoomLeaveReply{
		ServerToUserMessage: ServerToUserMessage{
			Type: "room_leaved",
		},
		Room:  room,
		Cause: cause,
	}
}

type RoomJoinRequest struct {
	UserToServerMessage
	RoomId string `json:"room_id"`
}

type RoomJoinReply struct {
	ServerToUserMessage
	Room *Room `json:"room"`
}

func NewReplyErrorRoomJoin(reason string) ErrorMessage {
	return ErrorMessage{
		Error:  "join_room_failure",
		Reason: reason,
	}
}

func NewRequestRoomJoin(msg []byte) (RoomJoinRequest, error) {
	request := RoomJoinRequest{}

	err := json.Unmarshal(msg, &request)
	if err != nil {
		return request, err
	}

	return request, nil
}

func NewReplyRoomJoined(room *Room) RoomJoinReply {
	return RoomJoinReply{
		ServerToUserMessage: ServerToUserMessage{
			Type: "room_joined",
		},
		Room: room,
	}
}
