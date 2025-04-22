package service

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/krez3f4l/audit_logger/pkg/domain/audit"
	"golang.org/x/exp/rand"

	"hippo/internal/domain"
	"hippo/internal/platform/logger"
	"hippo/internal/repository"
)

type PasswordHasher interface {
	Hash(password string) (string, error)
	Compare(hash, pwd string) error
}

type UsersRepository interface {
	Create(ctx context.Context, user domain.User) error
	GetByEmail(ctx context.Context, email string) (domain.User, error)
}

type SessionRepository interface {
	Create(ctx context.Context, token domain.RefreshSession) error
	Get(ctx context.Context, token string) (domain.RefreshSession, error)
}

type Users struct {
	repo        UsersRepository
	sessionRepo SessionRepository
	hasher      PasswordHasher

	auditClient AuditClient

	hmacSecret       []byte
	refreshTokenLife time.Duration
	accessTokenLife  time.Duration
	log              logger.Logger
}

func NewUsers(
	repo UsersRepository,
	sessionRepo SessionRepository,
	hasher PasswordHasher,
	auditClient AuditClient,
	secret []byte,
	refreshTokenLife time.Duration,
	accessTokenLife time.Duration,
	log logger.Logger,
) *Users {

	return &Users{
		repo:             repo,
		sessionRepo:      sessionRepo,
		hasher:           hasher,
		auditClient:      auditClient,
		hmacSecret:       secret,
		refreshTokenLife: refreshTokenLife,
		accessTokenLife:  accessTokenLife,
		log:              log,
	}
}

func (s *Users) SignUp(ctx context.Context, sInfo domain.SignUpInfo) (int64, error) {
	password, err := s.hasher.Hash(sInfo.Password)
	if err != nil {
		return -1, err
	}

	user := domain.User{
		Name:      sInfo.Name,
		Email:     sInfo.Email,
		Password:  password,
		CreatedAt: time.Now(),
	}

	if err = s.repo.Create(ctx, user); err != nil {
		var duplicateEmail *repository.ErrDuplicateEmail
		if errors.As(err, &duplicateEmail) {
			return -1, NewErrDuplicateEmail(err)
		}
		return -1, err
	}

	user, err = s.repo.GetByEmail(ctx, sInfo.Email)
	if err != nil {
		return -1, err
	}

	go s.runAuditCall(ctx, audit.ACTION_REGISTER, audit.ENTITY_USER, user.ID)

	return user.ID, nil
}

func (s *Users) SignIn(ctx context.Context, sInfo domain.SignInInfo) (string, string, error) {
	user, err := s.repo.GetByEmail(ctx, sInfo.Email)
	if err != nil {
		var invalidCred *repository.ErrInvalidCredential
		if errors.As(err, &invalidCred) {
			return "", "", NewErrInvalidCredential(err)
		}
		return "", "", err
	}

	if err = s.hasher.Compare(user.Password, sInfo.Password); err != nil {
		return "", "", NewErrInvalidCredential(err)
	}

	accessToken, refreshToken, err := s.generateTokens(ctx, user.ID)
	if err != nil {
		return "", "", err
	}

	go s.runAuditCall(ctx, audit.ACTION_LOGIN, audit.ENTITY_USER, user.ID)

	return accessToken, refreshToken, nil
}

func (s *Users) ParseToken(ctx context.Context, token string) (int64, error) {
	t, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.hmacSecret, nil
	})

	if err != nil {
		return 0, err
	}

	if !t.Valid {
		return 0, errors.New("invalid token")
	}

	claims, ok := t.Claims.(jwt.MapClaims)
	if !ok {
		return 0, errors.New("invalid claims")
	}

	subject, ok := claims["sub"].(string)
	if !ok {
		return 0, errors.New("invalid subject")
	}

	id, err := strconv.Atoi(subject)
	if err != nil {
		return 0, errors.New("invalid subject")
	}

	return int64(id), nil
}

func (s *Users) generateTokens(ctx context.Context, userId int64) (string, string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.StandardClaims{
		Subject:   strconv.Itoa(int(userId)),
		IssuedAt:  time.Now().Unix(),
		ExpiresAt: time.Now().Add(s.accessTokenLife).Unix(),
	})

	accessToken, err := token.SignedString(s.hmacSecret)
	if err != nil {
		return "", "", err
	}

	refreshToken, err := newRefreshToken()
	if err != nil {
		return "", "", err
	}

	if err := s.sessionRepo.Create(ctx, domain.RefreshSession{
		UserID:    userId,
		Token:     refreshToken,
		ExpiresAt: time.Now().Add(s.refreshTokenLife),
	}); err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

func newRefreshToken() (string, error) {
	b := make([]byte, 32)

	s := rand.NewSource(uint64(time.Now().Unix()))
	r := rand.New(s)

	if _, err := r.Read(b); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", b), nil
}

func (s *Users) RefreshToken(ctx context.Context, refreshToken string) (string, string, error) {
	session, err := s.sessionRepo.Get(ctx, refreshToken)
	if err != nil {
		return "", "", err
	}

	if session.ExpiresAt.Unix() < time.Now().Unix() {
		return "", "", NewErrRefreshTokenExpired()
	}

	return s.generateTokens(ctx, session.UserID)
}

func (s *Users) runAuditCall(ctx context.Context, entity, action string, id int64) {
	logErr := s.auditClient.SendLogRequest(ctx, audit.LogItem{
		Entity:    entity,
		Action:    action,
		EntityID:  id,
		Timestamp: time.Now(),
	})
	if logErr != nil {
		s.log.Warn("audit log failed", logger.Err(logErr))
	}
}
