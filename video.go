package main

import "github.com/pion/webrtc/v4"

func addVideoCodecAV1(mediaEngine *webrtc.MediaEngine) error {
	err := mediaEngine.RegisterCodec(webrtc.RTPCodecParameters{
		RTPCodecCapability: webrtc.RTPCodecCapability{
			MimeType:     webrtc.MimeTypeAV1,
			ClockRate:    90000,
			Channels:     0,
			SDPFmtpLine:  "",
			RTCPFeedback: []webrtc.RTCPFeedback{{"goog-remb", ""}, {"ccm", "fir"}, {"nack", ""}, {"nack", "pli"}},
		},
		PayloadType: 45,
	}, webrtc.RTPCodecTypeVideo)
	if err != nil {
		return err
	}

	err = mediaEngine.RegisterCodec(webrtc.RTPCodecParameters{
		RTPCodecCapability: webrtc.RTPCodecCapability{
			MimeType:     webrtc.MimeTypeRTX,
			ClockRate:    90000,
			Channels:     0,
			SDPFmtpLine:  "apt=45",
			RTCPFeedback: nil,
		},
		PayloadType: 46,
	}, webrtc.RTPCodecTypeVideo)
	if err != nil {
		return err
	}

	return nil
}
