package main

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"sync"
	"time"

	"github.com/google/uuid"
)

type Room struct {
	Id              string  `json:"id"`
	Users           []*User `json:"users"`
	usersMutex      *sync.Mutex
	cleanRoomTicker *time.Ticker
	cancel          context.CancelFunc
}

var (
	rooms      map[string]*Room = make(map[string]*Room)
	roomsMutex *sync.RWMutex    = new(sync.RWMutex)
)

func NewRoom() *Room {
	room := &Room{
		Id:              uuid.NewString(),
		Users:           make([]*User, 0),
		usersMutex:      new(sync.Mutex),
		cleanRoomTicker: time.NewTicker(30 * time.Second),
		cancel:          nil,
	}

	ctx, cancel := context.WithCancel(context.Background())
	room.cancel = cancel
	go room.checkCleanRoom(ctx)

	AddRoom(room)

	return room
}

func (room *Room) AddUser(user *User) error {
	room.usersMutex.Lock()
	defer room.usersMutex.Unlock()

	if slices.Contains(room.Users, user) {
		return errors.New("user already is the room")
	}

	room.Users = append(room.Users, user)

	return nil
}

func (room *Room) RemoveUser(user *User) error {
	room.usersMutex.Lock()
	defer room.usersMutex.Unlock()

	if !slices.Contains(room.Users, user) {
		return errors.New("user is not in the room")
	}

	idx := slices.Index(room.Users, user)
	room.Users = slices.Delete(room.Users, idx, idx+1)

	return nil
}

func (room *Room) Destroy() {
	room.cancel()

	for _, user := range room.Users {
		user.LeaveCurrentRoom("room has been destroyed")
	}

	roomsMutex.Lock()
	delete(rooms, room.Id)
	roomsMutex.Unlock()

	logger.Info(fmt.Sprintf("room %s has been destroyed", room.Id))
}

func (room *Room) checkCleanRoom(ctx context.Context) {
	defer room.cleanRoomTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-room.cleanRoomTicker.C:
			roomsMutex.RLock()
			users := room.Users
			roomsMutex.RUnlock()

			if len(users) > 0 {
				continue
			}

			room.Destroy()
			return
		}
	}
}

func AddRoom(room *Room) {
	roomsMutex.Lock()
	rooms[room.Id] = room
	roomsMutex.Unlock()
}

func GetRoom(id string) *Room {
	roomsMutex.RLock()
	defer roomsMutex.RUnlock()

	if room, ok := rooms[id]; ok {
		return room
	}
	return nil
}

func RemoveRoom(room *Room) {
	roomsMutex.Lock()
	delete(rooms, room.Id)
	roomsMutex.Unlock()
}
