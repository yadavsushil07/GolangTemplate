package service

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/yadavsushil07/GolangTemplate/internal/model"
	"github.com/yadavsushil07/GolangTemplate/internal/repository"
)

type otpEntry struct {
	code      string
	expiresAt time.Time
}

type AuthService struct {
	userRepo      *repository.UserRepository
	jwtSecret     []byte
	otpExpiry     time.Duration
	mu            sync.Mutex
	otpStore      map[string]otpEntry
}

func NewAuthService(userRepo *repository.UserRepository, jwtSecret string, otpExpiryMinutes int) *AuthService {
	return &AuthService{
		userRepo:  userRepo,
		jwtSecret: []byte(jwtSecret),
		otpExpiry: time.Duration(otpExpiryMinutes) * time.Minute,
		otpStore:  make(map[string]otpEntry),
	}
}

func (s *AuthService) RequestOTP(identifier string) (string, error) {
	identifier = strings.TrimSpace(identifier)
	if identifier == "" {
		return "", fmt.Errorf("identifier is required")
	}

	b := make([]byte, 3)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate OTP")
	}
	code := fmt.Sprintf("%06d", int(b[0])<<16|int(b[1])<<8|int(b[2]))
	code = code[len(code)-6:]

	s.mu.Lock()
	s.otpStore[identifier] = otpEntry{code: code, expiresAt: time.Now().Add(s.otpExpiry)}
	s.mu.Unlock()

	return code, nil
}

func (s *AuthService) VerifyOTP(ctx context.Context, identifier, code string) (string, *model.User, error) {
	identifier = strings.TrimSpace(identifier)
	code = strings.TrimSpace(code)

	s.mu.Lock()
	entry, ok := s.otpStore[identifier]
	if ok {
		delete(s.otpStore, identifier)
	}
	s.mu.Unlock()

	if !ok {
		return "", nil, fmt.Errorf("no active OTP for this identifier")
	}
	if time.Now().After(entry.expiresAt) {
		return "", nil, fmt.Errorf("OTP has expired")
	}
	if entry.code != code {
		return "", nil, fmt.Errorf("invalid OTP")
	}

	user, err := s.userRepo.FindByIdentifier(ctx, identifier)
	if err != nil {
		return "", nil, err
	}
	if user == nil {
		user, err = s.userRepo.Create(ctx, identifier, model.RoleCustomer)
		if err != nil {
			return "", nil, err
		}
	}

	token, err := s.issueJWT(user)
	if err != nil {
		return "", nil, err
	}
	return token, user, nil
}

func (s *AuthService) issueJWT(user *model.User) (string, error) {
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))
	payloadJSON, err := json.Marshal(map[string]any{
		"sub":  fmt.Sprintf("%d", user.ID),
		"id":   user.ID,
		"role": user.Role,
		"exp":  time.Now().Add(24 * time.Hour).Unix(),
	})
	if err != nil {
		return "", err
	}
	payload := base64.RawURLEncoding.EncodeToString(payloadJSON)
	signingInput := header + "." + payload
	sig := s.hmacSHA256(signingInput)
	return signingInput + "." + base64.RawURLEncoding.EncodeToString(sig), nil
}

func (s *AuthService) ValidateJWT(token string) (int64, string, bool) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return 0, "", false
	}
	expectedSig := base64.RawURLEncoding.EncodeToString(s.hmacSHA256(parts[0] + "." + parts[1]))
	if !hmac.Equal([]byte(parts[2]), []byte(expectedSig)) {
		return 0, "", false
	}
	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return 0, "", false
	}
	var claims map[string]any
	if err := json.Unmarshal(payloadBytes, &claims); err != nil {
		return 0, "", false
	}
	exp, _ := claims["exp"].(float64)
	if time.Unix(int64(exp), 0).Before(time.Now()) {
		return 0, "", false
	}
	idFloat, _ := claims["id"].(float64)
	role, _ := claims["role"].(string)
	return int64(idFloat), role, true
}

func (s *AuthService) hmacSHA256(message string) []byte {
	h := hmac.New(sha256.New, s.jwtSecret)
	_, _ = h.Write([]byte(message))
	return h.Sum(nil)
}
