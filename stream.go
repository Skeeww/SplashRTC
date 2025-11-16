package main

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/pion/webrtc/v4"
)

type IncomingStream struct {
	Id             string                 `json:"id"`
	Publisher      *User                  `json:"publisher"`
	PeerConnection *webrtc.PeerConnection `json:"-"`

	receiversChannel map[*webrtc.RTPReceiver]chan []byte
}

func NewIncomingStream(user *User) (*IncomingStream, error) {
	if user.Room == nil {
		return nil, errors.New("you should join a room before publishing a stream")
	}

	stream := &IncomingStream{
		Id:             uuid.NewString(),
		Publisher:      user,
		PeerConnection: nil,

		receiversChannel: make(map[*webrtc.RTPReceiver]chan []byte),
	}
	pc, err := user.Room.Api.NewPeerConnection(webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun2.l.google.com:19302"},
			},
		},
		BundlePolicy: webrtc.BundlePolicyMaxBundle,
	})
	if err != nil {
		logger.Warn("peer connection failed", err.Error())
		return nil, err
	}
	stream.PeerConnection = pc

	stream.PeerConnection.OnSignalingStateChange(func(rs webrtc.SignalingState) {
		logger.Debug(fmt.Sprintf("signaling state of stream %s changed to %s", stream.Id, rs.String()))
	})
	stream.PeerConnection.OnICEConnectionStateChange(func(cs webrtc.ICEConnectionState) {
		logger.Debug(fmt.Sprintf("ice state of stream %s changed to %s", stream.Id, cs.String()))
	})
	stream.PeerConnection.OnConnectionStateChange(func(pcs webrtc.PeerConnectionState) {
		logger.Debug(fmt.Sprintf("peer state of stream %s changed to %s", stream.Id, pcs.String()))
	})
	stream.PeerConnection.OnTrack(func(t *webrtc.TrackRemote, r *webrtc.RTPReceiver) {
		logger.Debug(fmt.Sprintf("new track on stream %s => %s", stream.Id, t.ID()))
		go stream.handleRTP(r)
	})

	if err := user.Room.AddInStream(stream); err != nil {
		stream.Teardown()
		return nil, err
	}

	for _, candidate := range user.GetIceCandidates() {
		stream.AddIceCandidate(candidate)
	}

	return stream, nil
}

func (s *IncomingStream) AddIceCandidate(candidate webrtc.ICECandidateInit) {
	if err := s.PeerConnection.AddICECandidate(candidate); err != nil {
		logger.Warn(fmt.Sprintf("failed add ice candidate to stream %s", s.Id))
	}
	logger.Debug(fmt.Sprintf("add candidate %v to stream %s", candidate, s.Id))
}

func (s *IncomingStream) Teardown() {
	if err := s.PeerConnection.GracefulClose(); err != nil {
		logger.Warn("failed closing peer connection", err.Error())
	}
}

func (s *IncomingStream) handleRTP(r *webrtc.RTPReceiver) {
	buffer := make([]byte, 1500)
	s.receiversChannel[r] = make(chan []byte)

	defer func() {
		close(s.receiversChannel[r])
		delete(s.receiversChannel, r)
	}()

	for {
		_, _, err := r.Read(buffer)
		if err != nil {
			return
		}
		s.receiversChannel[r] <- buffer
	}
}
