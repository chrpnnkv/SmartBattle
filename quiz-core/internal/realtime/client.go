package realtime

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Client interface {
	CreateRoom(ctx context.Context, hostID uuid.UUID, req CreateRoomRequest) (CreateRoomResponse, error)

	GetRoom(ctx context.Context, pin string) (RoomInfo, error)

	GetParticipants(ctx context.Context, pin string) (RoomParticipants, error)
}

type HTTPClient struct {
	baseURL   string
	jwtSecret string
	http      *http.Client
}

func NewHTTPClient(baseURL, jwtSecret string) *HTTPClient {
	return &HTTPClient{
		baseURL:   baseURL,
		jwtSecret: jwtSecret,
		http:      &http.Client{Timeout: DefaultCreateRoomTimeout},
	}
}

var ErrRoomNotFound = errors.New("realtime: room not found")

func (c *HTTPClient) CreateRoom(ctx context.Context, hostID uuid.UUID, req CreateRoomRequest) (CreateRoomResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return CreateRoomResponse{}, fmt.Errorf("marshal create room request: %w", err)
	}

	tokenString, err := c.signServiceToken(hostID)
	if err != nil {
		return CreateRoomResponse{}, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/rooms", bytes.NewBuffer(body))
	if err != nil {
		return CreateRoomResponse{}, fmt.Errorf("build create room request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+tokenString)

	resp, err := c.http.Do(httpReq)
	if err != nil {
		return CreateRoomResponse{}, fmt.Errorf("call realtime create room: %w", err)
	}
	defer resp.Body.Close()

	var rtResp CreateRoomResponse
	if err := json.NewDecoder(resp.Body).Decode(&rtResp); err != nil {
		return CreateRoomResponse{}, fmt.Errorf("decode realtime response: %w", err)
	}
	if resp.StatusCode != http.StatusCreated {
		if rtResp.Error == "" {
			rtResp.Error = fmt.Sprintf("status %d", resp.StatusCode)
		}
		return rtResp, fmt.Errorf("realtime create room: %s", rtResp.Error)
	}
	return rtResp, nil
}

func (c *HTTPClient) GetRoom(ctx context.Context, pin string) (RoomInfo, error) {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/api/rooms/"+pin, nil)
	if err != nil {
		return RoomInfo{}, fmt.Errorf("build get room request: %w", err)
	}

	resp, err := c.http.Do(httpReq)
	if err != nil {
		return RoomInfo{}, fmt.Errorf("call realtime get room: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return RoomInfo{}, ErrRoomNotFound
	}
	if resp.StatusCode != http.StatusOK {
		return RoomInfo{}, fmt.Errorf("realtime get room status %d", resp.StatusCode)
	}

	var info RoomInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return RoomInfo{}, fmt.Errorf("decode room info: %w", err)
	}
	return info, nil
}

func (c *HTTPClient) GetParticipants(ctx context.Context, pin string) (RoomParticipants, error) {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/api/rooms/"+pin+"/participants", nil)
	if err != nil {
		return RoomParticipants{}, fmt.Errorf("build participants request: %w", err)
	}

	resp, err := c.http.Do(httpReq)
	if err != nil {
		return RoomParticipants{}, fmt.Errorf("call realtime participants: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return RoomParticipants{}, fmt.Errorf("realtime participants status %d", resp.StatusCode)
	}

	var body RoomParticipants
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return RoomParticipants{}, fmt.Errorf("decode participants: %w", err)
	}
	if body.Participants == nil {
		body.Participants = []Participant{}
	}
	return body, nil
}

func (c *HTTPClient) signServiceToken(hostID uuid.UUID) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": hostID.String(),
		"role":    "teacher",
		"exp":     time.Now().Add(time.Minute * 5).Unix(),
	})
	signed, err := token.SignedString([]byte(c.jwtSecret))
	if err != nil || signed == "" {
		return "", fmt.Errorf("sign service token: %w", err)
	}
	return signed, nil
}
