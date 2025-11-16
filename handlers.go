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
		user.handleCreateRoom(msg)
	case "leave_room":
		user.handleLeaveRoom()
	case "join_room":
		user.handleJoinRoom(msg)
	case "publish":
		user.handlePublish(msg)
	case "icecandidate":
		user.handleIceCandidate(msg)
	default:
		logger.Warn(fmt.Sprintf("received message of unknown type from %s", user.Id))
	}
}

func (user *User) handleUsersList() {
	SendUsersList(user.SendMessage)
}

func (user *User) handleCreateRoom(msg []byte) {
	if user.Room != nil {
		user.SendMessageJson(NewReplyErrorRoomCreate("you are already in a room"))
		return
	}

	newRoomOptions := new(NewRoomOptions)
	err := json.Unmarshal(msg, newRoomOptions)
	if err != nil {
		user.SendMessageJson(NewReplyErrorRoomCreate("you provided wrongly formatted options"))
	}

	room, err := NewRoom(&NewRoomOptions{
		VideoCodec: newRoomOptions.VideoCodec,
	})
	if err != nil {
		user.SendMessageJson(NewReplyErrorRoomCreate(err.Error()))
		return
	}

	logger.Info(fmt.Sprintf("user %s create a new room %s", user.Id, room.Id))

	if err := user.JoinRoom(room); err != nil {
		room.Destroy()
		user.SendMessageJson(NewReplyErrorRoomCreate(err.Error()))
		return
	}
}

func (user *User) handleLeaveRoom() {
	if user.Room == nil {
		user.SendMessageJson(NewReplyErrorRoomLeave("you are not in a room"))
		return
	}

	if err := user.LeaveCurrentRoom("leave action"); err != nil {
		user.SendMessageJson(NewReplyErrorRoomLeave(err.Error()))
		return
	}
}

func (user *User) handleJoinRoom(msg []byte) {
	request, err := NewRequestRoomJoin(msg)
	if err != nil {
		user.SendMessageJson(NewReplyErrorRoomJoin(err.Error()))
		return
	}

	room := GetRoom(request.RoomId)
	if room == nil {
		user.SendMessageJson(NewReplyErrorRoomJoin("the room does not exist"))
		return
	}

	if err := user.JoinRoom(room); err != nil {
		user.SendMessageJson(NewReplyErrorRoomJoin(err.Error()))
		return
	}
}

func (user *User) handlePublish(msg []byte) {
	payload, err := NewRequestPublish(msg)
	if err != nil {
		user.SendMessageJson(NewReplyErrorPublish(err.Error()))
		return
	}

	stream, err := NewIncomingStream(user)
	if err != nil {
		user.SendMessageJson(NewReplyErrorPublish(err.Error()))
		return
	}

	err = stream.PeerConnection.SetRemoteDescription(payload.SdpOffer)
	if err != nil {
		user.SendMessageJson(NewReplyErrorPublish(err.Error()))
		return
	}

	sdpAnswer, err := stream.PeerConnection.CreateAnswer(nil)
	if err != nil {
		user.SendMessageJson(NewReplyErrorPublish(err.Error()))
		return
	}

	err = stream.PeerConnection.SetLocalDescription(sdpAnswer)
	if err != nil {
		user.SendMessageJson(NewReplyErrorPublish(err.Error()))
		return
	}

	user.SendMessageJson(NewReplyPublish(stream, sdpAnswer))
}

func (user *User) handleIceCandidate(msg []byte) {
	request, err := NewRequestIceCandidate(msg)
	if err != nil {
		user.SendMessageJson(NewReplyErrorIceCandidate(err.Error()))
		return
	}

	user.AddIceCandidate(request.IceCandidate)
}
