package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type User struct {
	Id   string          `json:"id"`
	Conn *websocket.Conn `json:"-"`
	Room *Room           `json:"-"`
}

var (
	users      []*User       = make([]*User, 0)
	usersMutex *sync.RWMutex = new(sync.RWMutex)
)

func NewUser(conn *websocket.Conn) *User {
	user := &User{
		Id:   uuid.NewString(),
		Conn: conn,
		Room: nil,
	}

	go func() {
		for {
			_, data, err := user.Conn.ReadMessage()
			if err != nil {
				if user.Room != nil {
					if err := user.LeaveCurrentRoom("user disconnected"); err != nil {
						logger.Warn(fmt.Sprintf("user %s failed leaving is current room during disconnection, %s", user.Id, err.Error()))
					}
				}
				RemoveUser(user)
				return
			}
			user.handleMessage(data)
		}
	}()

	AddUser(user)

	return user
}

func (user *User) SendMessage(msg string) {
	buffer := bytes.NewBufferString(msg)
	if err := user.Conn.WriteMessage(websocket.TextMessage, buffer.Bytes()); err != nil {
		logger.Warn(fmt.Sprintf("failed sending message to user %s, %s", user.Id, err.Error()))
	}
}

func (user *User) SendMessageJson(msg any) {
	payload, err := json.Marshal(msg)
	if err != nil {
		logger.Warn(fmt.Sprintf("failed sending json message, %s", err.Error()))
		return
	}
	if err := user.Conn.WriteMessage(websocket.TextMessage, payload); err != nil {
		logger.Warn(fmt.Sprintf("failed sending json message to user %s, %s", user.Id, err.Error()))
	}
}

func (user *User) JoinRoom(room *Room) error {
	if user.Room != nil {
		if err := user.LeaveCurrentRoom("leave current room, because joining another one"); err != nil {
			return err
		}
	}

	if err := room.AddUser(user); err != nil {
		return fmt.Errorf("failed adding user to room, %w", err)
	}

	user.SendMessageJson(NewReplyRoomJoined(room))
	logger.Info(fmt.Sprintf("user %s join the room %s", user.Id, room.Id))

	user.Room = room

	return nil
}

func (user *User) LeaveCurrentRoom(cause string) error {
	if user.Room == nil {
		return errors.New("no room to leave")
	}

	if err := user.Room.RemoveUser(user); err != nil {
		return fmt.Errorf("failed removing user from room, %w", err)
	}

	user.SendMessageJson(NewReplyRoomLeaved(user.Room, cause))
	logger.Info(fmt.Sprintf("user %s leave the room %s, reason: %s", user.Id, user.Room.Id, cause))

	user.Room = nil

	return nil
}

func (user *User) String() string {
	return fmt.Sprintf("Id: %s", user.Id)
}

func HttpToUser(w http.ResponseWriter, r *http.Request) (*User, error) {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return nil, err
	}

	return NewUser(conn), nil
}

func AddUser(user *User) {
	usersMutex.Lock()
	users = append(users, user)
	usersMutex.Unlock()

	logger.Debug(fmt.Sprintf("add user %s to repository", user.Id))
	SendUsersList(Broadcast)
}

func RemoveUser(user *User) {
	usersMutex.Lock()
	idx := slices.Index(users, user)
	users = slices.Delete(users, idx, idx+1)
	usersMutex.Unlock()

	logger.Debug(fmt.Sprintf("removed user %s from repository", user.Id))
	SendUsersList(Broadcast)
}

func Broadcast(msg string) {
	for _, user := range users {
		user.SendMessage(msg)
	}
}

func SendUsersList(sendFunc func(string)) {
	usersMutex.RLock()
	message := NewReplyUsersList(users)
	usersMutex.RUnlock()

	payload, err := json.Marshal(message)
	if err != nil {
		logger.Error(fmt.Sprintf("can't send users list, %s", err.Error()))
		return
	}

	logger.Debug("send users list")
	sendFunc(string(payload))
}
