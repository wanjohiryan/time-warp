package warp

import (
	"context"

	"github.com/kixelated/invoker"
	"github.com/kixelated/quic-go"
	"github.com/kixelated/webtransport-go"
)

// TODO: create a heartbeat stream, to get the exact latency between the server and client

// A single WebTransport session
type PlaySession struct {
	conn  quic.Connection
	inner *webtransport.Session

	pad *Gpad

	streams invoker.Tasks
}

func NewPlaySession(connection quic.Connection, session *webtransport.Session, gpad *Gpad) (s *PlaySession, err error) {
	s = new(PlaySession)
	s.conn = connection
	s.inner = session
	s.pad = gpad
	return s, nil
}

func (s *PlaySession) Play(ctx context.Context) (err error) {
	// return invoker.Run()
}

// func (s *Session) Run(ctx context.Context) (err error) {
// 	s.inits, s.audio, s.video, err = s.media.Start(s.conn.GetMaxBandwidth)
// 	if err != nil {
// 		return fmt.Errorf("failed to start media: %w", err)
// 	}

// 	// Once we've validated the session, now we can start accessing the streams
// 	return invoker.Run(ctx, s.runAccept, s.runAcceptUni, s.runInit, s.runAudio, s.runVideo, s.streams.Repeat, s.heartBeat)
// }

// func (s *Session) runAccept(ctx context.Context) (err error) {
// 	for {
// 		stream, err := s.inner.AcceptStream(ctx)
// 		if err != nil {
// 			return fmt.Errorf("failed to accept bidirectional stream: %w", err)
// 		}

// 		//TODO: handle gamepad input with a bidirectional stream ?
// 		// s.streams.Add(func(ctx context.Context) (err error) {
// 		// 	return s.handleGamepads(ctx, stream)
// 		// })
// 		// Warp doesn't utilize bidirectional streams so just close them immediately.
// 		// We might use them in the future so don't close the connection with an error.
// 		stream.CancelRead(1)
// 	}
// }

// func (s *Session) runAcceptUni(ctx context.Context) (err error) {
// 	for {
// 		stream, err := s.inner.AcceptUniStream(ctx)
// 		if err != nil {
// 			return fmt.Errorf("failed to accept unidirectional stream: %w", err)
// 		}

// 		s.streams.Add(func(ctx context.Context) (err error) {
// 			return s.handleStream(ctx, stream)
// 		})
// 	}
// }

// func (s *Session) handleStream(ctx context.Context, stream webtransport.ReceiveStream) (err error) {
// 	defer func() {
// 		if err != nil {
// 			stream.CancelRead(1)
// 		}
// 	}()

// 	var header [8]byte
// 	for {
// 		_, err = io.ReadFull(stream, header[:])
// 		if errors.Is(io.EOF, err) {
// 			return nil
// 		} else if err != nil {
// 			return fmt.Errorf("failed to read atom header: %w", err)
// 		}

// 		size := binary.BigEndian.Uint32(header[0:4])
// 		name := string(header[4:8])

// 		if size < 8 {
// 			return fmt.Errorf("atom size is too small")
// 		} else if size > 42069 { // arbitrary limit
// 			return fmt.Errorf("atom size is too large")
// 		} else if name != "warp" {
// 			return fmt.Errorf("only warp atoms are supported")
// 		}

// 		payload := make([]byte, size-8)

// 		_, err = io.ReadFull(stream, payload)
// 		if err != nil {
// 			return fmt.Errorf("failed to read atom payload: %w", err)
// 		}

// 		log.Println("received message:", string(payload))

// 		msg := Message{}

// 		err = json.Unmarshal(payload, &msg)
// 		if err != nil {
// 			return fmt.Errorf("failed to decode json payload: %w", err)
// 		}

// 		if msg.Debug != nil {
// 			s.setDebug(msg.Debug)
// 		}
// 	}
// }

// func (s *Session) runInit(ctx context.Context) (err error) {
// 	for _, init := range s.inits {
// 		err = s.writeInit(ctx, init)
// 		if err != nil {
// 			return fmt.Errorf("failed to write init stream: %w", err)
// 		}
// 	}

// 	return nil
// }

// func (s *Session) runAudio(ctx context.Context) (err error) {
// 	for {
// 		segment, err := s.audio.Next(ctx)
// 		if err != nil {
// 			return fmt.Errorf("failed to get next segment: %w", err)
// 		}

// 		if segment == nil {
// 			return nil
// 		}

// 		err = s.writeSegment(ctx, segment)
// 		if err != nil {
// 			return fmt.Errorf("failed to write segment stream: %w", err)
// 		}
// 	}
// }

// func (s *Session) runVideo(ctx context.Context) (err error) {
// 	for {
// 		segment, err := s.video.Next(ctx)
// 		if err != nil {
// 			return fmt.Errorf("failed to get next segment: %w", err)
// 		}

// 		if segment == nil {
// 			return nil
// 		}

// 		err = s.writeSegment(ctx, segment)
// 		if err != nil {
// 			return fmt.Errorf("failed to write segment stream: %w", err)
// 		}
// 	}
// }

// // Create a stream for an INIT segment and write the container.
// func (s *Session) writeInit(ctx context.Context, init *MediaInit) (err error) {
// 	temp, err := s.inner.OpenUniStreamSync(ctx)
// 	if err != nil {
// 		return fmt.Errorf("failed to create stream: %w", err)
// 	}

// 	// Wrap the stream in an object that buffers writes instead of blocking.
// 	stream := NewStream(temp)
// 	s.streams.Add(stream.Run)

// 	defer func() {
// 		if err != nil {
// 			stream.WriteCancel(1)
// 		}
// 	}()

// 	stream.SetPriority(math.MaxInt)

// 	err = stream.WriteMessage(Message{
// 		Init: &MessageInit{Id: init.ID},
// 	})
// 	if err != nil {
// 		return fmt.Errorf("failed to write init header: %w", err)
// 	}

// 	_, err = stream.Write(init.Raw)
// 	if err != nil {
// 		return fmt.Errorf("failed to write init data: %w", err)
// 	}

// 	return nil
// }

// // Create a stream for a segment and write the contents, chunk by chunk.
// func (s *Session) writeSegment(ctx context.Context, segment *MediaSegment) (err error) {
// 	temp, err := s.inner.OpenUniStreamSync(ctx)
// 	if err != nil {
// 		return fmt.Errorf("failed to create stream: %w", err)
// 	}

// 	// Wrap the stream in an object that buffers writes instead of blocking.
// 	stream := NewStream(temp)
// 	s.streams.Add(stream.Run)

// 	defer func() {
// 		if err != nil {
// 			stream.WriteCancel(1)
// 		}
// 	}()

// 	ms := int(segment.timestamp / time.Millisecond)

// 	// newer segments take priority
// 	stream.SetPriority(ms)

// 	err = stream.WriteMessage(Message{
// 		Segment: &MessageSegment{
// 			Init:      segment.Init.ID,
// 			Timestamp: ms,
// 		},
// 	})
// 	if err != nil {
// 		return fmt.Errorf("failed to write segment header: %w", err)
// 	}

// 	for {
// 		// Get the next fragment
// 		buf, err := segment.Read(ctx)
// 		if errors.Is(err, io.EOF) {
// 			break
// 		} else if err != nil {
// 			return fmt.Errorf("failed to read segment data: %w", err)
// 		}

// 		// NOTE: This won't block because of our wrapper
// 		_, err = stream.Write(buf)
// 		if err != nil {
// 			return fmt.Errorf("failed to write segment data: %w", err)
// 		}
// 	}

// 	err = stream.Close()
// 	if err != nil {
// 		return fmt.Errorf("failed to close segment stream: %w", err)
// 	}

// 	return nil
// }

// // get latency between server and client via a heartbeat uni-stream
// func (s *Session) heartBeat(ctx context.Context) (err error) {

// 	temp, err := s.inner.OpenUniStreamSync(ctx)

// 	if err != nil {
// 		return fmt.Errorf("failed to create stream: %w", err)
// 	}

// 	// Wrap the stream in an object that buffers writes instead of blocking.
// 	stream := NewStream(temp)
// 	s.streams.Add(stream.Run)

// 	defer func() {
// 		if err != nil {
// 			stream.WriteCancel(1)
// 		}
// 	}()

// 	start := time.Now()

// 	for {
// 		ms := int(time.Since(start).Milliseconds() / 1000)

// 		// newer heartbeats take priority
// 		stream.SetPriority(ms)

// 		timeNow := int(time.Now().UnixMilli())

// 		err = stream.WriteMessage(Message{
// 			Beat: &MessageHeartBeat{
// 				Timestamp: timeNow,
// 			},
// 		})

// 		if err != nil {
// 			return fmt.Errorf("failed to write heart beat: %w", err)
// 		}

// 		//every 2 seconds
// 		time.Sleep(2 * time.Second)
// 	}
// }

// // handle gamepads for each session
// // func (s *Session) handleGamepads(ctx context.Context, stream webtransport.ReceiveStream) (err error){
// // 	defer func() {
// // 		if err != nil {
// // 			stream.CancelRead(1)
// // 		}
// // 	}()
// // }

// // set max bitrate for bandwidth
// func (s *Session) setDebug(msg *MessageDebug) {
// 	s.conn.SetMaxBandwidth(uint64(msg.MaxBitrate))
// }
