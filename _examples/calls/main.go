// Copyright 2023 Pius Alfred <me.pius1102@gmail.com>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy of this software
// and associated documentation files (the “Software”), to deal in the Software without restriction,
// including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense,
// and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so,
// subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all copies or substantial
// portions of the Software.
//
// THE SOFTWARE IS PROVIDED “AS IS”, WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT
// LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
// IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY,
// WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE
// SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

// Package main implements a server application that handles incoming WhatsApp voice calls
// using the WhatsApp Business Platform API. It uses WebRTC to establish a peer-to-peer
// connection, records the caller's audio, and can play back a pre-recorded message.
package main

//)
//
//// Config holds all the necessary configuration for the application.
//// These values are typically obtained from your Meta for Developers dashboard
//// and server environment.
//type Config struct {
//	// WhatsApp Business Platform API credentials.
//	BaseURL           string
//	APIVersion        string
//	AccessToken       string
//	PhoneNumberID     string
//	BusinessAccountID string
//	AppSecret         string
//	AppID             string
//
//	// SecureRequests enables HTTPS signature validation for API requests.
//	SecureRequests bool
//
//	// Webhook configuration.
//	WebhookToken                   string // A secret token to verify webhook origins.
//	WebhookValidate                bool   // If true, validates webhook signature.
//	WebhookAppSecret               string // App secret for webhook signature validation.
//	WebhookEndpoint                string // The path for receiving webhook notifications (e.g., "/webhook").
//	WebhookSubVerificationEndpoint string // The path for handling subscription verification requests.
//
//	// Application-specific settings.
//	RecordingsDir     string        // Directory to save call recordings.
//	Addr              string        // The address and port for the HTTP server (e.g., ":8080").
//	MaxRecordDuration time.Duration // The maximum duration for a single call recording.
//	PromptWAV         string        // Path to a WAV file to play at the beginning of a call.
//
//	// WebRTC ICE server configuration.
//	ICEServers []webrtc.ICEServer
//}
//
//// Client encapsulates the application's HTTP router.
//type Client struct {
//	Router *router.SimpleRouter
//}
//
//// NewClient creates and configures the main application client, which includes
//// setting up the webhook listener and routing.
//func NewClient(cfg *Config, handler *CallHandler) (*Client, error) {
//	// Ensure the directory for storing recordings exists.
//	if err := os.MkdirAll(cfg.RecordingsDir, 0o755); err != nil {
//		return nil, fmt.Errorf("failed to create recordings directory %s: %w", cfg.RecordingsDir, err)
//	}
//
//	// Create a webhook handler and register our custom CallHandler to process call status updates.
//	h := webhooks.NewHandler()
//	h.SetBusinessCallStatusUpdateHandler(handler)
//
//	// Create a webhook listener. It uses a ConfigReaderFunc to dynamically provide
//	// configuration (like tokens) for validating incoming webhook requests.
//	l := webhooks.NewListener(h, webhooks.ConfigReaderFunc(func(request *http.Request) (*webhooks.Config, error) {
//		webhookCfg := &webhooks.Config{
//			Token:     cfg.WebhookToken,
//			Validate:  cfg.WebhookValidate,
//			AppSecret: cfg.WebhookAppSecret,
//		}
//		return webhookCfg, nil
//	}))
//
//	// Set up a simple HTTP router that directs webhook and verification requests
//	// to the listener.
//	sr, err := router.NewSimpleRouter(
//		l,
//		router.WithSimpleRouterEndpoints(router.Endpoints{
//			Webhook:                  cfg.WebhookEndpoint,
//			SubscriptionVerification: cfg.WebhookSubVerificationEndpoint,
//		}),
//	)
//	if err != nil {
//		return nil, fmt.Errorf("failed to create new router: %w", err)
//	}
//
//	c := &Client{
//		Router: sr,
//	}
//
//	return c, nil
//}
//
//func main() {
//	// In a real application, this Config struct would be populated from a file,
//	// environment variables, or command-line flags.
//	cfg := &Config{
//		// Example: Addr = os.Getenv("PORT")
//	}
//
//	// Create the call handler, which contains the core logic for managing WebRTC calls.
//	ch, err := NewCallHandler(&config.Config{
//		BaseURL:           cfg.BaseURL,
//		APIVersion:        cfg.APIVersion,
//		AccessToken:       cfg.AccessToken,
//		PhoneNumberID:     cfg.PhoneNumberID,
//		BusinessAccountID: cfg.BusinessAccountID,
//		AppSecret:         cfg.AppSecret,
//		AppID:             cfg.AppID,
//		SecureRequests:    cfg.SecureRequests,
//	}, WithRecordingsDir(cfg.RecordingsDir),
//		WithPromptWAV(cfg.PromptWAV),
//		WithMaxRecordDuration(cfg.MaxRecordDuration),
//		WithICEServers(cfg.ICEServers...),
//	)
//	if err != nil {
//		fmt.Printf("Error creating CallHandler: %v\n", err)
//		os.Exit(1)
//	}
//
//	// Create the HTTP client and router.
//	c, err := NewClient(cfg, ch)
//	if err != nil {
//		fmt.Printf("Error creating client: %v\n", err)
//		os.Exit(1)
//	}
//
//	// Start the HTTP server to listen for incoming webhooks.
//	fmt.Printf("Starting server on %s\n", cfg.Addr)
//	if err := http.ListenAndServe(cfg.Addr, c.Router); err != nil {
//		fmt.Printf("Failed to start HTTP server: %v\n", err)
//		os.Exit(1)
//	}
//}
//
//// इंश्योर करता है कि CallHandler, webhooks.EventHandler इंटरफ़ेस को सही ढंग से लागू करता है।
//// यह एक कंपाइल-टाइम चेक है।
//var _ webhooks.EventHandler[webhooks.BusinessNotificationContext, webhooks.CallStatusUpdate] = (*CallHandler)(nil)
//
//// CallHandler manages the state and logic for all active WhatsApp calls.
//// It acts as a state machine, handling events for each call from initiation to termination.
//type CallHandler struct {
//	MessageClientConfig *config.Config     // Configuration for the WhatsApp API client.
//	MessageClient       *calls.BaseClient  // Client to send API requests (e.g., accept, pre-accept).
//	RecordingsDir       string             // Directory to store OGG recordings.
//	PromptWAV           string             // Path to a WAV file to play at the start of a call.
//	MaxRecordDuration   time.Duration      // Maximum recording duration.
//	ICEServers          []webrtc.ICEServer // STUN/TURN servers for WebRTC NAT traversal.
//
//	// A mutex is crucial here to protect concurrent access to the `peers` map,
//	// as webhook events for different calls can arrive simultaneously in separate goroutines.
//	mu    sync.Mutex
//	peers map[string]*peerState // Maps a call ID to its active WebRTC peer state.
//}
//
//// peerState holds the active WebRTC resources for a single call.
//type peerState struct {
//	pc          *webrtc.PeerConnection         // The main WebRTC peer connection object.
//	senderTrack *webrtc.TrackLocalStaticSample // The audio track we send to the user (e.g., prompt, beep).
//}
//
//// CallHandlerOption is a function type used for configuring a CallHandler instance.
//// This is the "functional options" pattern, providing a clean and extensible way to
//// initialize the handler.
//type CallHandlerOption func(*CallHandler)
//
//// WithRecordingsDir sets the directory for saving call recordings.
//func WithRecordingsDir(dir string) CallHandlerOption {
//	return func(h *CallHandler) { h.RecordingsDir = dir }
//}
//
//// WithPromptWAV sets the path to the WAV prompt file.
//func WithPromptWAV(path string) CallHandlerOption {
//	return func(h *CallHandler) { h.PromptWAV = path }
//}
//
//// WithMaxRecordDuration sets the maximum duration for call recordings.
//func WithMaxRecordDuration(d time.Duration) CallHandlerOption {
//	return func(h *CallHandler) { h.MaxRecordDuration = d }
//}
//
//// WithICEServers configures the ICE servers for WebRTC.
//func WithICEServers(servers ...webrtc.ICEServer) CallHandlerOption {
//	return func(h *CallHandler) { h.ICEServers = append([]webrtc.ICEServer{}, servers...) }
//}
//
//// NewCallHandler creates a new CallHandler with default settings that can be
//// overridden by functional options.
//func NewCallHandler(cfg *config.Config, opts ...CallHandlerOption) (*CallHandler, error) {
//	// Initialize with default values.
//	h := &CallHandler{
//		MessageClientConfig: cfg,
//		MessageClient:       &calls.BaseClient{Sender: whttp.NewSender[calls.Request]()},
//		RecordingsDir:       ".recordings/",
//		PromptWAV:           "beep.wav",
//		MaxRecordDuration:   2 * time.Minute,
//		ICEServers: []webrtc.ICEServer{
//			{URLs: []string{"stun:stun.l.google.com:19302"}}, // A public STUN server.
//		},
//		peers: make(map[string]*peerState),
//	}
//
//	// Apply any custom configurations provided.
//	for _, o := range opts {
//		o(h)
//	}
//
//	// Ensure the configured recordings directory exists.
//	if err := os.MkdirAll(h.RecordingsDir, 0o755); err != nil {
//		return nil, fmt.Errorf("failed to create recordings directory %s: %w", h.RecordingsDir, err)
//	}
//
//	return h, nil
//}
//
//// HandleEvent is the entry point for processing call status updates from the webhook.
//func (h *CallHandler) HandleEvent(
//	ctx context.Context,
//	ntx *webhooks.BusinessNotificationContext,
//	notification *webhooks.CallStatusUpdate,
//) error {
//	_ = os.MkdirAll(h.RecordingsDir, 0o755)
//
//	// A single notification can contain updates for multiple calls.
//	for _, c := range notification.Calls {
//		callID := c.ID
//
//		// A webhook with an SDP "offer" signifies a new incoming call that we need to answer.
//		// This is the primary trigger to start a WebRTC connection.
//		if c.Session != nil && strings.EqualFold(c.Session.SDPType, "offer") && c.Session.SDP != "" {
//			if err := h.handleConnect(ctx, callID, c.Session.SDP); err != nil {
//				// Returning an error here would cause the webhook provider to retry,
//				// so we log it and continue.
//				fmt.Printf("Error handling connect for call %s: %v\n", callID, err)
//			}
//			continue
//		}
//
//		// Log other status updates for debugging and monitoring purposes.
//		if c.Status != "" {
//			fmt.Printf("[status] call_id=%s status=%s event=%s\n", callID, c.Status, c.Event)
//		}
//
//		// If the call has ended (terminated status, terminate event, or has an end time),
//		// we must clean up the associated peer connection and state.
//		if strings.EqualFold(c.Status, "terminated") || strings.EqualFold(c.Event, "terminate") || c.EndTime != "" {
//			h.cleanup(callID)
//		}
//	}
//	return nil
//}
//
//// handleConnect manages the multi-step process required by WhatsApp to answer an incoming call.
//// Flow: (Receive Offer) -> Create Answer -> Send PreAccept -> Send Accept
//func (h *CallHandler) handleConnect(ctx context.Context, callID, offerSDP string) error {
//	// 1. Create a new WebRTC peer connection and an outgoing audio track.
//	pc, senderTrack, err := h.newPeer(callID)
//	if err != nil {
//		return fmt.Errorf("failed to create new peer: %w", err)
//	}
//	// Store the new peer state so it can be managed.
//	h.setPeer(callID, &peerState{pc: pc, senderTrack: senderTrack})
//
//	// 2. Create an SDP answer based on the received offer.
//	answer, err := h.createAnswer(ctx, pc, offerSDP)
//	if err != nil {
//		h.cleanup(callID) // Clean up if we fail.
//		return fmt.Errorf("failed to create answer: %w", err)
//	}
//
//	// 3. Send a "PreAccept" action to WhatsApp. This signals our intent to accept the call.
//	if _, err := h.UpdateCallStatus(ctx, &calls.Request{
//		CallID: callID,
//		Action: calls.PreAcceptCallAction,
//		Session: &calls.SessionInfo{
//			SDPType: "answer",
//			SDP:     answer,
//		},
//	}); err != nil {
//		h.cleanup(callID)
//		return fmt.Errorf("failed to send pre_accept: %w", err)
//	}
//
//	// 4. Send the final "Accept" action. The SDP answer must be identical to the one in PreAccept.
//	if _, err := h.UpdateCallStatus(ctx, &calls.Request{
//		CallID: callID,
//		Action: calls.AcceptCallAction,
//		Session: &calls.SessionInfo{
//			SDPType: "answer",
//			SDP:     answer,
//		},
//	}); err != nil {
//		h.cleanup(callID)
//		return fmt.Errorf("failed to send accept: %w", err)
//	}
//
//	// Once the call is accepted, play any configured audio prompts in a separate goroutine
//	// to avoid blocking the main event handling flow.
//	go func() {
//		if h.PromptWAV != "" {
//			if err := playWAV(senderTrack, h.PromptWAV); err != nil {
//				fmt.Printf("[prompt][%s] Error playing WAV: %v\n", callID, err)
//			}
//		}
//		// Example of playing a simple beep sound.
//		if err := beep(senderTrack, 400*time.Millisecond); err != nil {
//			fmt.Printf("[beep][%s] Error playing beep: %v\n", callID, err)
//		}
//	}()
//	return nil
//}
//
//// UpdateCallStatus sends a status update to the WhatsApp API.
//func (h *CallHandler) UpdateCallStatus(ctx context.Context, request *calls.Request) (*calls.Response, error) {
//	response, err := h.MessageClient.Send(ctx, h.MessageClientConfig, request)
//	if err != nil {
//		return nil, fmt.Errorf("failed to send message: %w", err)
//	}
//	return response, nil
//}
//
//// --- Peer State Management ---
//
//// setPeer safely adds or updates a peer's state in the map.
//func (h *CallHandler) setPeer(id string, st *peerState) {
//	h.mu.Lock()
//	defer h.mu.Unlock()
//	h.peers[id] = st
//}
//
//// getPeer safely retrieves a peer's state from the map.
//func (h *CallHandler) getPeer(id string) *peerState {
//	h.mu.Lock()
//	defer h.mu.Unlock()
//	return h.peers[id]
//}
//
//// delPeer safely removes a peer's state from the map.
//func (h *CallHandler) delPeer(id string) {
//	h.mu.Lock()
//	defer h.mu.Unlock()
//	delete(h.peers, id)
//}
//
//// cleanup closes the peer connection and removes the peer state from the map.
//// This is essential to release resources and prevent memory leaks.
//func (h *CallHandler) cleanup(callID string) {
//	if st := h.getPeer(callID); st != nil {
//		if err := st.pc.Close(); err != nil {
//			fmt.Printf("[cleanup] Error closing peer connection for call %s: %v\n", callID, err)
//		}
//		h.delPeer(callID)
//		fmt.Printf("[cleanup] Cleaned up resources for call %s\n", callID)
//	}
//}
//
//// newPeer creates and configures a new WebRTC PeerConnection for a call.
//func (h *CallHandler) newPeer(callID string) (*webrtc.PeerConnection, *webrtc.TrackLocalStaticSample, error) {
//	// Create a new PeerConnection with the configured ICE servers.
//	pc, err := webrtc.NewPeerConnection(webrtc.Configuration{ICEServers: h.ICEServers})
//	if err != nil {
//		return nil, nil, fmt.Errorf("failed to create peer connection: %w", err)
//	}
//
//	// Create an outgoing audio track for the server. This track will be used to send
//	// audio (like prompts or beeps) to the user.
//	// WhatsApp expects Opus audio at a 48kHz clock rate with a single channel (mono).
//	track, err := webrtc.NewTrackLocalStaticSample(
//		webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeOpus, ClockRate: 48000, Channels: 1},
//		"server-audio", "whatsapp",
//	)
//	if err != nil {
//		_ = pc.Close()
//		return nil, nil, fmt.Errorf("failed to create local track: %w", err)
//	}
//
//	// Add the track to a transceiver that is configured to be send-only.
//	if _, err := pc.AddTransceiverFromTrack(track, webrtc.RTPTransceiverInit{
//		Direction: webrtc.RTPTransceiverDirectionSendonly,
//	}); err != nil {
//		_ = pc.Close()
//		return nil, nil, fmt.Errorf("failed to add transceiver: %w", err)
//	}
//
//	// Set up a callback to handle incoming audio tracks from the user.
//	// This is where we will receive the caller's voice.
//	pc.OnTrack(func(remote *webrtc.TrackRemote, _ *webrtc.RTPReceiver) {
//		if remote.Kind() != webrtc.RTPCodecTypeAudio {
//			return
//		}
//
//		// Prepare to save the incoming audio to an OGG file.
//		_ = os.MkdirAll(h.RecordingsDir, 0o755)
//		path := filepath.Join(h.RecordingsDir, callID+".ogg")
//		w, err := oggwriter.New(path, 48000, 1) // 48kHz sample rate, 1 channel.
//		if err != nil {
//			fmt.Printf("[ogg open][%s] Failed to create ogg writer: %v\n", callID, err)
//			return
//		}
//		fmt.Printf("[recording] -> Saving audio for call %s to %s (max duration: %s)\n", callID, path, h.MaxRecordDuration)
//
//		// Start a goroutine to read RTP packets from the remote track and write them to the OGG file.
//		// A timer enforces the maximum recording duration.
//		deadline := time.NewTimer(h.MaxRecordDuration)
//		go func() {
//			defer func() {
//				_ = w.Close()
//				fmt.Printf("[recording closed] %s\n", path)
//			}()
//			for {
//				select {
//				case <-deadline.C:
//					fmt.Printf("[recording] Max duration reached for call %s\n", callID)
//					return
//				default:
//					// ReadRTP blocks until a packet is received.
//					pkt, _, err := remote.ReadRTP()
//					if err != nil {
//						// Errors (like io.EOF) indicate the track has ended.
//						return
//					}
//					// Write the raw RTP packet to the OGG file.
//					if err := w.WriteRTP(pkt); err != nil {
//						return
//					}
//				}
//			}
//		}()
//	})
//
//	// Log connection state changes for debugging.
//	pc.OnConnectionStateChange(func(s webrtc.PeerConnectionState) {
//		fmt.Printf("[pc %s] Connection state has changed: %s\n", callID, s.String())
//	})
//
//	return pc, track, nil
//}
//
//// createAnswer performs the SDP offer/answer exchange to establish the WebRTC connection.
//func (h *CallHandler) createAnswer(ctx context.Context, pc *webrtc.PeerConnection, offerSDP string) (string, error) {
//	// Set the remote description from the offer received from WhatsApp.
//	offer := webrtc.SessionDescription{Type: webrtc.SDPTypeOffer, SDP: offerSDP}
//	if err := pc.SetRemoteDescription(offer); err != nil {
//		return "", fmt.Errorf("failed to set remote description: %w", err)
//	}
//
//	// Create an answer to the offer.
//	ans, err := pc.CreateAnswer(nil)
//	if err != nil {
//		return "", fmt.Errorf("failed to create answer: %w", err)
//	}
//
//	// Create a promise that resolves when ICE gathering is complete.
//	// This ensures that our SDP answer includes all necessary ICE candidates.
//	g := webrtc.GatheringCompletePromise(pc)
//
//	// Set the local description. This starts the ICE gathering process.
//	if err := pc.SetLocalDescription(ans); err != nil {
//		return "", fmt.Errorf("failed to set local description: %w", err)
//	}
//
//	// Wait for ICE gathering to complete or for the context to be canceled.
//	select {
//	case <-g:
//		// Gathering is complete. Return the SDP of the local description.
//		return pc.LocalDescription().SDP, nil
//	case <-ctx.Done():
//		return "", ctx.Err()
//	}
//}
//
//// --- Audio Generation and Playback ---
//
//// beep generates a sine wave tone and writes it to the outgoing audio track.
//func beep(track *webrtc.TrackLocalStaticSample, d time.Duration) error {
//	const rate = 48000
//	const frame = rate / 50 // 960 samples for a 20ms frame at 48kHz.
//	n := int(float64(rate) * d.Seconds())
//	pcm := make([]int16, n)
//
//	// Generate a 1000Hz sine wave.
//	const freq = 1000.0
//	for i := 0; i < n; i++ {
//		// Calculate the sine wave value and scale it to the 16-bit PCM range.
//		v := 0.25 * math.Sin(2*math.Pi*freq*float64(i)/rate)
//		pcm[i] = int16(v * 32767)
//	}
//	return writePCMOpus(track, pcm, rate, frame)
//}
//
//// playWAV reads a WAV file, decodes it into PCM audio, and writes it to the outgoing track.
//func playWAV(track *webrtc.TrackLocalStaticSample, path string) error {
//	b, err := os.ReadFile(path)
//	if err != nil {
//		return fmt.Errorf("failed to read WAV file %s: %w", path, err)
//	}
//	pcm, err := simpleWAVToPCM16Mono48k(b)
//	if err != nil {
//		return fmt.Errorf("failed to decode WAV file %s: %w", path, err)
//	}
//	// The frame size must be 960 for 20ms of audio at 48kHz.
//	return writePCMOpus(track, pcm, 48000, 960)
//}
//
//// writePCMOpus encodes raw PCM audio into Opus frames and writes them to the track.
//func writePCMOpus(track *webrtc.TrackLocalStaticSample, pcm []int16, rate, frame int) error {
//	// Initialize the Opus encoder.
//	enc, err := gopus.NewEncoder(rate, 1, gopus.Audio)
//	if err != nil {
//		return fmt.Errorf("failed to create opus encoder: %w", err)
//	}
//	// Set a reasonable bitrate for voice.
//	enc.SetBitrate(32000)
//
//	// Iterate through the PCM data in chunks (frames).
//	for i := 0; i+frame <= len(pcm); i += frame {
//		chunk := pcm[i : i+frame]
//		// Encode the PCM chunk into an Opus frame.
//		opus, err := enc.Encode(chunk, frame, 4000) // 4000 bytes is a safe buffer size.
//		if err != nil {
//			return fmt.Errorf("opus encoding failed: %w", err)
//		}
//
//		// Write the encoded Opus data as a sample to the WebRTC track.
//		// Each sample corresponds to 20ms of audio.
//		if err := track.WriteSample(media.Sample{Data: opus, Duration: 20 * time.Millisecond}); err != nil {
//			return fmt.Errorf("failed to write sample to track: %w", err)
//		}
//
//		// Sleep for the duration of the sample to simulate real-time streaming.
//		time.Sleep(20 * time.Millisecond)
//	}
//	return nil
//}
//
//// --- WAV Parsing Utilities ---
//
//// simpleWAVToPCM16Mono48k is a minimal WAV file parser. It makes strong assumptions
//// about the file format: it must be uncompressed PCM, 16-bit, mono, and 48kHz sample rate.
//func simpleWAVToPCM16Mono48k(b []byte) ([]int16, error) {
//	// Basic validation of the RIFF/WAVE header.
//	if len(b) < 44 || string(b[0:4]) != "RIFF" || string(b[8:12]) != "WAVE" {
//		return nil, errors.New("invalid WAV header: not a RIFF/WAVE file")
//	}
//
//	// Read format metadata from the 'fmt ' chunk.
//	audioFmt := u16(b[20:22]) // Audio format (1 for PCM).
//	chans := u16(b[22:24])    // Number of channels (1 for mono).
//	sr := u32(b[24:28])       // Sample rate (must be 48000).
//	bps := u16(b[34:36])      // Bits per sample (must be 16).
//
//	// Enforce the required format.
//	if audioFmt != 1 || chans != 1 || sr != 48000 || bps != 16 {
//		return nil, fmt.Errorf("unsupported WAV format: need 16-bit mono 48kHz PCM, but got fmt=%d ch=%d sr=%d bps=%d", audioFmt, chans, sr, bps)
//	}
//
//	// Find the 'data' chunk which contains the raw audio samples.
//	off := 12 // Start search after 'WAVE'.
//	for off+8 <= len(b) {
//		id := string(b[off : off+4])
//		sz := int(u32(b[off+4 : off+8]))
//		off += 8
//		if id == "data" {
//			if off+sz > len(b) || sz%2 != 0 {
//				return nil, errors.New("invalid or corrupt data chunk in WAV file")
//			}
//			// Convert the raw byte data into 16-bit integer samples.
//			pcm := make([]int16, sz/2)
//			for i := 0; i < len(pcm); i++ {
//				// Little-endian conversion.
//				pcm[i] = int16(uint16(b[off+2*i]) | uint16(b[off+2*i+1])<<8)
//			}
//			return pcm, nil
//		}
//		// Skip to the next chunk.
//		off += sz
//	}
//	return nil, errors.New("WAV file does not contain a 'data' chunk")
//}
//
//// u16 decodes a 2-byte little-endian unsigned integer.
//func u16(b []byte) uint16 { return uint16(b[0]) | uint16(b[1])<<8 }
//
//// u32 decodes a 4-byte little-endian unsigned integer.
//func u32(b []byte) uint32 {
//	return uint32(b[0]) | uint32(b[1])<<8 | uint32(b[2])<<16 | uint32(b[3])<<24
//}

func main() {}
