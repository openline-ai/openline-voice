package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"golang.org/x/net/websocket"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

type AudioTranscription struct {
	Prediction []struct {
		TimeBegin     int     `json:"time_begin"`
		TimeEnd       float64 `json:"time_end"`
		Transcription string  `json:"transcription"`
		Language      string  `json:"language"`
		Probability   float64 `json:"probability"`
		Speaker       string  `json:"speaker"`
		Channel       string  `json:"channel"`
	} `json:"prediction"`
	PredictionRaw struct {
		Metadata struct {
			TotalSpeechDuration         float64 `json:"total_speech_duration"`
			TotalSpeechDurationChannel0 float64 `json:"total_speech_duration_channel_0"`
			TotalSpeechDurationChannel1 float64 `json:"total_speech_duration_channel_1"`
			AudioConversionTime         float64 `json:"audio_conversion_time"`
			VadTime                     float64 `json:"vad_time"`
			InferenceTime               float64 `json:"inference_time"`
			DiarizationTime             float64 `json:"diarization_time"`
			TotalTranscriptionTime      float64 `json:"total_transcription_time"`
			OriginalFileType            string  `json:"original_file_type"`
			OriginalNbChannels          int     `json:"original_nb_channels"`
			OriginalSampleRate          int     `json:"original_sample_rate"`
			OriginalSampleWidth         int     `json:"original_sample_width"`
			OriginalNbSilentChannels    int     `json:"original_nb_silent_channels"`
			OriginalNbSimilarChannels   int     `json:"original_nb_similar_channels"`
			OriginalMediainfo           struct {
				Index            string `json:"index"`
				CodecName        string `json:"codec_name"`
				CodecLongName    string `json:"codec_long_name"`
				Profile          string `json:"profile"`
				CodecType        string `json:"codec_type"`
				CodecTimeBase    string `json:"codec_time_base"`
				CodecTagString   string `json:"codec_tag_string"`
				CodecTag         string `json:"codec_tag"`
				SampleFmt        string `json:"sample_fmt"`
				SampleRate       string `json:"sample_rate"`
				Channels         string `json:"channels"`
				ChannelLayout    string `json:"channel_layout"`
				BitsPerSample    string `json:"bits_per_sample"`
				ID               string `json:"id"`
				RFrameRate       string `json:"r_frame_rate"`
				AvgFrameRate     string `json:"avg_frame_rate"`
				TimeBase         string `json:"time_base"`
				StartPts         string `json:"start_pts"`
				StartTime        string `json:"start_time"`
				DurationTs       string `json:"duration_ts"`
				Duration         string `json:"duration"`
				BitRate          string `json:"bit_rate"`
				MaxBitRate       string `json:"max_bit_rate"`
				BitsPerRawSample string `json:"bits_per_raw_sample"`
				NbFrames         string `json:"nb_frames"`
				NbReadFrames     string `json:"nb_read_frames"`
				NbReadPackets    string `json:"nb_read_packets"`
				Disposition      struct {
					Default         string `json:"default"`
					Dub             string `json:"dub"`
					Original        string `json:"original"`
					Comment         string `json:"comment"`
					Lyrics          string `json:"lyrics"`
					Karaoke         string `json:"karaoke"`
					Forced          string `json:"forced"`
					HearingImpaired string `json:"hearing_impaired"`
					VisualImpaired  string `json:"visual_impaired"`
					CleanEffects    string `json:"clean_effects"`
					AttachedPic     string `json:"attached_pic"`
					TimedThumbnails string `json:"timed_thumbnails"`
				} `json:"DISPOSITION"`
				Tag struct {
					Comment string `json:"Comment"`
				} `json:"TAG"`
				NbStreams      string `json:"nb_streams"`
				NbPrograms     string `json:"nb_programs"`
				FormatName     string `json:"format_name"`
				FormatLongName string `json:"format_long_name"`
				Size           string `json:"size"`
				ProbeScore     string `json:"probe_score"`
			} `json:"original_mediainfo"`
		} `json:"metadata"`
		Transcription []struct {
			TimeBegin     int     `json:"time_begin"`
			TimeEnd       float64 `json:"time_end"`
			Transcription string  `json:"transcription"`
			Language      string  `json:"language"`
			Probability   float64 `json:"probability"`
			Speaker       string  `json:"speaker"`
			Channel       string  `json:"channel"`
		} `json:"transcription"`
	} `json:"prediction_raw"`
}

type gladiaPayload struct {
	Frames     string `json:"frames"`
	SampleRate int    `json:"sample_rate"`
}

type GladiaClient struct {
	conn           *websocket.Conn
	currentText    string
	channel        chan string
	audioChannel   chan []byte
	completed      chan interface{}
	audioCompleted chan interface{}
	bytes          *bytes.Buffer
	sampleRate     int
	conf           *RecordServiceConfig
}

func swapBytes(b []byte) []byte {
	for i := 0; i < len(b); i += 2 {
		b[i], b[i+1] = b[i+1], b[i]
	}
	return b
}
func (g *GladiaClient) processPacket(payload []byte) {
	g.bytes.Write(payload)
	if g.bytes.Len() >= 15000 {
		msgBytes := make([]byte, 15000)
		_, _ = g.bytes.Read(msgBytes)
		msgBytes = swapBytes(msgBytes)
		msgString := base64.StdEncoding.EncodeToString(msgBytes)

		msg, _ := json.Marshal(gladiaPayload{Frames: msgString, SampleRate: g.sampleRate})
		//log.Printf("Sending audio: %v", string(msg))
		g.conn.Write(msg)
	}
}

func (g *GladiaClient) AudioLoop() {
	log.Printf("Starting AudioLoop")
	silence := make([]byte, 1920)
	nextPacket := time.Now().Add(20 * time.Second) // don't generate silence until first packet arrives
	for {
		select {
		case <-time.After(nextPacket.Sub(time.Now())):
			log.Printf("Silence detected!")
			g.processPacket(silence)
			nextPacket = time.Now().Add(20 * time.Millisecond)
		case payload := <-g.audioChannel:
			nextPacket = time.Now().Add(25 * time.Millisecond) // allow 5 seconds of jitter
			g.processPacket(payload)
		case <-g.audioCompleted:
			log.Printf("Shutting down AudioLoop")
			return
		}
	}
}

func (g *GladiaClient) SendAudio(payload []byte) {
	g.audioChannel <- payload
}

func (g *GladiaClient) ReadText() {
	log.Printf("Starting ReadText")
	for {
		var msg string
		err := websocket.Message.Receive(g.conn, &msg)
		if err != nil {
			log.Printf("Error reading from websocket: %v", err)
			g.completed <- struct{}{}
			g.audioCompleted <- struct{}{}
			return
		}
		if msg == "" {
			if g.currentText != "" {
				g.channel <- g.currentText
			}
		}
		g.currentText = msg
	}
}

func (g *GladiaClient) Close() {
	g.conn.Close()
}

func TranscribeAudio(conf *RecordServiceConfig, filename string, person1 string, person2 string) (string, error) {
	file, _ := os.Open(filename)
	defer file.Close()
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("audio", filepath.Base(file.Name()))
	io.Copy(part, file)
	writer.WriteField("language", "english")
	writer.WriteField("language_behaviour", "automatic single language")
	writer.WriteField("output_format", "json")
	writer.WriteField("toggle_diarization", "false")
	writer.Close()
	r, _ := http.NewRequest("POST", "https://api.gladia.io/audio/text/audio-transcription/", body)
	r.Header.Add("Content-Type", writer.FormDataContentType())
	r.Header.Add("Accept", "application/json")
	r.Header.Add("x-gladia-key", conf.GladiaApiKey)

	client := &http.Client{}
	res, err := client.Do(r)

	if err != nil {
		log.Printf("TranscribeAudio: could not send request: %s\n", err)
		return "", err
	}

	log.Printf("TranscribeAudio: got response!\n")
	log.Printf("TranscribeAudio: status code: %d\n", res.StatusCode)
	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		log.Printf("TranscribeAudio: could not read response body: %s\n", err)
		return "", err

	}
	transcription := &AudioTranscription{}
	err = json.Unmarshal(resBody, &transcription)
	if err != nil {
		log.Printf("TranscribeAudio: could not unmarshal response body: %s\n", err)
		return "", err
	}

	transcriptionString := ""
	for _, t := range transcription.Prediction {
		if t.Channel == "channel_0" {
			transcriptionString += person1 + ": " + t.Transcription + "\n"
		} else if t.Channel == "channel_1" {
			transcriptionString += person2 + ": " + t.Transcription + "\n"
		}
	}
	return transcriptionString, nil
}

func NewGladiaClient(sampleRate int, conf *RecordServiceConfig) *GladiaClient {
	conn, err := websocket.Dial("wss://aipi-riva-ws.k0s.gladia.io/audio/text/audio-transcription", "", "https://app.gladia.io")
	if err != nil {
		log.Printf("Error connecting to websocket: %v", err)
		return nil
	}
	log.Printf("Gladia Client: Connected to websocket: %v", conn)
	return &GladiaClient{conn: conn,
		currentText:    "",
		channel:        make(chan string),
		audioChannel:   make(chan []byte),
		completed:      make(chan interface{}),
		audioCompleted: make(chan interface{}),
		bytes:          bytes.NewBuffer(make([]byte, 2000)),
		sampleRate:     sampleRate,
		conf:           conf,
	}
}
