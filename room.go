package main

import (
	"errors"
	"fmt"
	"slices"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/pion/webrtc/v4"
)

type NewRoomOptions struct {
	VideoCodec string `json:"video_codec,omitempty"`
}

type Room struct {
	Id         string  `json:"id"`
	Users      []*User `json:"users"`
	usersMutex *sync.Mutex

	Api *webrtc.API `json:"-"`

	VideoCodec string `json:"video_codec"`
	AudioCodec string `json:"audio_codec"`

	InStreams      map[string]*IncomingStream `json:"in_streams"`
	inStreamsMutex *sync.Mutex

	timeoutDestroyStarted       *atomic.Bool
	cancelTimeoutDestroyChannel chan bool
}

var (
	rooms      map[string]*Room = make(map[string]*Room)
	roomsMutex *sync.RWMutex    = new(sync.RWMutex)
)

func NewRoom(opts *NewRoomOptions) (*Room, error) {
	mediaEngine := new(webrtc.MediaEngine)

	if opts != nil && opts.VideoCodec != "" {
		switch opts.VideoCodec {
		case "av1":
			if err := addVideoCodecAV1(mediaEngine); err != nil {
				return nil, err
			}
		default:
			return nil, errors.New("unsupported provided video codec")
		}
	} else {
		if err := mediaEngine.RegisterDefaultCodecs(); err != nil {
			return nil, err
		}
	}

	api := webrtc.NewAPI(webrtc.WithMediaEngine(mediaEngine))

	room := &Room{
		Id:         uuid.NewString(),
		Users:      make([]*User, 0),
		usersMutex: new(sync.Mutex),

		Api: api,

		InStreams:      make(map[string]*IncomingStream),
		inStreamsMutex: new(sync.Mutex),

		timeoutDestroyStarted:       new(atomic.Bool),
		cancelTimeoutDestroyChannel: make(chan bool),
	}
	room.timeoutDestroyStarted.Store(false)

	AddRoom(room)

	return room, nil
}

func (room *Room) AddUser(user *User) error {
	room.usersMutex.Lock()
	defer room.usersMutex.Unlock()

	if slices.Contains(room.Users, user) {
		return errors.New("user already is the room")
	}

	if room.timeoutDestroyStarted.Load() {
		room.stopDestroyTimeout()
		logger.Info(fmt.Sprintf("room %s destroy cancel, new user joined the room", room.Id))
	}

	room.Users = append(room.Users, user)
	logger.Debug(fmt.Sprintf("add user %s to room %s", user.Id, room.Id))

	return nil
}

func (room *Room) RemoveUser(user *User) error {
	room.usersMutex.Lock()
	defer room.usersMutex.Unlock()

	if !slices.Contains(room.Users, user) {
		return errors.New("user is not in the room")
	}

	for _, stream := range room.GetInStreamsByPublisher(user) {
		stream.Teardown()
		if err := room.RemoveInStream(stream); err != nil {
			logger.Warn("failed removing stream", stream.Id)
			continue
		}
	}

	idx := slices.Index(room.Users, user)
	room.Users = slices.Delete(room.Users, idx, idx+1)

	if len(room.Users) == 0 {
		logger.Info(fmt.Sprintf("room %s is empty, leaving timeout of 30 seconds before destroy", room.Id))

		room.timeoutDestroyStarted.Store(true)
		go room.startDestroyTimeout()
	}

	return nil
}

func (room *Room) Destroy() {
	close(room.cancelTimeoutDestroyChannel)

	for _, user := range room.Users {
		if err := user.LeaveCurrentRoom("room has been destroyed"); err != nil {
			logger.Warn(fmt.Sprintf("user %s failed leaving room %s, %s", user.Id, room.Id, err.Error()))
		}
	}

	roomsMutex.Lock()
	delete(rooms, room.Id)
	roomsMutex.Unlock()

	logger.Info(fmt.Sprintf("room %s has been destroyed", room.Id))
}

func (room *Room) AddInStream(stream *IncomingStream) error {
	room.inStreamsMutex.Lock()
	defer room.inStreamsMutex.Unlock()

	if _, ok := room.InStreams[stream.Id]; ok {
		return errors.New("this stream has already been published in this room")
	}

	room.InStreams[stream.Id] = stream

	// TODO: broadcast message in room to update streams

	return nil
}

func (room *Room) RemoveInStream(stream *IncomingStream) error {
	room.inStreamsMutex.Lock()
	defer room.inStreamsMutex.Unlock()

	if _, ok := room.InStreams[stream.Id]; !ok {
		return errors.New("this stream has not been published in this room")
	}

	delete(room.InStreams, stream.Id)
	logger.Debug(fmt.Sprintf("remove in stream %s from %s", stream.Id, stream.Publisher.Id))

	// TODO: broadcast message in room to update streams

	return nil
}

func (room *Room) GetInStreamsByPublisher(user *User) []*IncomingStream {
	room.inStreamsMutex.Lock()
	defer room.inStreamsMutex.Unlock()

	streams := make([]*IncomingStream, 0)
	for _, stream := range room.InStreams {
		if stream.Publisher != user {
			continue
		}
		streams = append(streams, stream)
	}

	return streams
}

func (room *Room) startDestroyTimeout() {
	timer := time.NewTimer(30 * time.Second)

	defer func() {
		timer.Stop()
		room.timeoutDestroyStarted.Store(false)
	}()

	select {
	case <-room.cancelTimeoutDestroyChannel:
		return
	case <-timer.C:
		room.Destroy()
		return
	}
}

func (room *Room) stopDestroyTimeout() {
	if !room.timeoutDestroyStarted.Load() {
		return
	}

	room.cancelTimeoutDestroyChannel <- true
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
