package main

import (
	"encoding/json"
	"fmt"
)

func (user *User) handleMessage(msg []byte) {
	payload := new(UserToServerMessage)

	err := json.Unmarshal(msg, payload)
	if err != nil {
		logger.Warn(fmt.Sprintf("can't parse msg from %s, %s", user.Id, err.Error()))
		return
	}

	logger.Debug(fmt.Sprintf("receive message of type %s from %s", payload.Type, user.Id))
	switch payload.Type {
	case "users_list":
		user.handleUsersList()
	case "create_room":
		user.handleCreateRoom()
	default:
		logger.Warn(fmt.Sprintf("received message of unknown type from %s", user.Id))
	}
}

func (user *User) handleUsersList() {
	SendUsersList(user.SendMessage)
}

func (user *User) handleCreateRoom() {
	if user.Room != nil {
		user.SendMessageJson(NewMessageErrorRoomCreation("you are already in a room"))
		return
	}

	room := NewRoom()
	logger.Info(fmt.Sprintf("user %s create a new room %s", user.Id, room.Id))

	if err := user.JoinRoom(room); err != nil {
		room.Destroy()
		user.SendMessageJson(NewMessageErrorRoomCreation(err.Error()))
		return
	}
}
