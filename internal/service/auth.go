package service

import (
	"errors"
	"goaltrack/internal/model"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthService struct {
	db *gorm.DB
}

func NewAuthService(db *gorm.DB) *AuthService {
	return &AuthService{db: db}
}

func (s *AuthService) Register(nickname, email, phone, password string) (*model.User, error) {
	var count int64
	s.db.Model(&model.User{}).Where("email = ? OR phone = ?", email, phone).Count(&count)
	if count > 0 {
		return nil, errors.New("邮箱或手机号已被注册")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &model.User{
		Nickname:     nickname,
		Email:        email,
		Phone:        phone,
		PasswordHash: string(hash),
	}
	if err := s.db.Create(user).Error; err != nil {
		return nil, err
	}
	return user, nil
}

func (s *AuthService) Login(account, password string) (*model.User, error) {
	var user model.User
	err := s.db.Where("email = ? OR phone = ?", account, account).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("账号不存在")
		}
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, errors.New("密码错误")
	}
	return &user, nil
}

func (s *AuthService) GetProfile(userID uint) (*model.User, error) {
	var user model.User
	err := s.db.First(&user, userID).Error
	return &user, err
}

func (s *AuthService) UpdateProfile(userID uint, updates map[string]interface{}) error {
	return s.db.Model(&model.User{}).Where("id = ?", userID).Updates(updates).Error
}
