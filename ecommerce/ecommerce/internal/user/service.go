package user

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/example/ecommerce/pkg/models"
)

type Service struct {
	repository *Repository
	auth       *AuthService
}

func NewService() *Service {
	return &Service{
		repository: NewRepository(),
		auth:       NewAuthService(),
	}
}

type RegistrationRequest struct {
	Email       string `json:"email"`
	Username    string `json:"username"`
	Password    string `json:"password"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	PhoneNumber string `json:"phone_number"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type UpdateProfileRequest struct {
	FirstName   string         `json:"first_name,omitempty"`
	LastName    string         `json:"last_name,omitempty"`
	PhoneNumber string         `json:"phone_number,omitempty"`
	Address     *models.Address `json:"address,omitempty"`
}

func (s *Service) RegisterUser(ctx context.Context, req RegistrationRequest) (*models.User, error) {
	if err := s.validateRegistration(req); err != nil {
		return nil, fmt.Errorf("registration validation failed: %w", err)
	}

	existing, _ := s.repository.FindByEmail(ctx, req.Email)
	if existing != nil {
		return nil, errors.New("email already registered")
	}

	existing, _ = s.repository.FindByUsername(ctx, req.Username)
	if existing != nil {
		return nil, errors.New("username already taken")
	}

	passwordHash, err := s.auth.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := &models.User{
		ID:           s.generateUserID(),
		Email:        req.Email,
		Username:     req.Username,
		PasswordHash: passwordHash,
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		PhoneNumber:  req.PhoneNumber,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := s.repository.Save(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to save user: %w", err)
	}

	go s.sendWelcomeEmail(context.Background(), user)

	return user, nil
}

func (s *Service) Login(ctx context.Context, req LoginRequest) (string, error) {
	if req.Username == "" || req.Password == "" {
		return "", errors.New("username and password are required")
	}

	user, err := s.repository.FindByUsername(ctx, req.Username)
	if err != nil {
		return "", errors.New("invalid credentials")
	}

	if !s.auth.VerifyPassword(req.Password, user.PasswordHash) {
		go s.recordFailedLogin(context.Background(), user.ID)
		return "", errors.New("invalid credentials")
	}

	token, err := s.auth.GenerateToken(user.ID, user.Username)
	if err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}

	user.LastLoginAt = time.Now()
	s.repository.Update(ctx, user)

	go s.recordSuccessfulLogin(context.Background(), user.ID)

	return token, nil
}

func (s *Service) UpdateProfile(ctx context.Context, userID string, req UpdateProfileRequest) error {
	user, err := s.repository.FindByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	if req.FirstName != "" {
		user.FirstName = req.FirstName
	}
	if req.LastName != "" {
		user.LastName = req.LastName
	}
	if req.PhoneNumber != "" {
		if !s.validatePhoneNumber(req.PhoneNumber) {
			return errors.New("invalid phone number format")
		}
		user.PhoneNumber = req.PhoneNumber
	}
	if req.Address != nil {
		user.Address = *req.Address
	}

	user.UpdatedAt = time.Now()

	if err := s.repository.Update(ctx, user); err != nil {
		return fmt.Errorf("failed to update profile: %w", err)
	}

	return nil
}

func (s *Service) validateRegistration(req RegistrationRequest) error {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(req.Email) {
		return errors.New("invalid email format")
	}

	if len(req.Username) < 3 || len(req.Username) > 20 {
		return errors.New("username must be between 3 and 20 characters")
	}

	if len(req.Password) < 8 {
		return errors.New("password must be at least 8 characters")
	}

	if req.FirstName == "" || req.LastName == "" {
		return errors.New("first name and last name are required")
	}

	return nil
}

func (s *Service) validatePhoneNumber(phone string) bool {
	phoneRegex := regexp.MustCompile(`^\+?[1-9]\d{1,14}$`)
	return phoneRegex.MatchString(phone)
}

func (s *Service) sendWelcomeEmail(ctx context.Context, user *models.User) {
	fmt.Printf("Sending welcome email to %s\n", user.Email)
}

func (s *Service) recordFailedLogin(ctx context.Context, userID string) {
	fmt.Printf("Failed login attempt for user %s at %s\n", userID, time.Now().Format(time.RFC3339))
}

func (s *Service) recordSuccessfulLogin(ctx context.Context, userID string) {
	fmt.Printf("Successful login for user %s at %s\n", userID, time.Now().Format(time.RFC3339))
}

func (s *Service) generateUserID() string {
	timestamp := time.Now().Unix()
	return fmt.Sprintf("USR-%d-%04d", timestamp, time.Now().Nanosecond()%10000)
}