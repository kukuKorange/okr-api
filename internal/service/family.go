package service

import (
	"crypto/rand"
	"errors"
	"math/big"
	"time"

	"goaltrack/internal/model"

	"gorm.io/gorm"
)

type FamilyService struct {
	db *gorm.DB
}

func NewFamilyService(db *gorm.DB) *FamilyService {
	return &FamilyService{db: db}
}

func (s *FamilyService) Create(userID uint, name string) (*model.Family, error) {
	var count int64
	s.db.Model(&model.Family{}).Where("creator_id = ?", userID).Count(&count)
	if count > 0 {
		return nil, errors.New("每个用户只能创建一个家庭")
	}

	code, err := s.generateInviteCode()
	if err != nil {
		return nil, err
	}

	family := &model.Family{
		CreatorID:  userID,
		Name:       name,
		InviteCode: code,
	}
	if err := s.db.Create(family).Error; err != nil {
		return nil, err
	}

	// Creator auto-joins as admin
	member := &model.FamilyMember{
		FamilyID: family.ID,
		UserID:   userID,
		Role:     "admin",
		JoinedAt: time.Now(),
	}
	s.db.Create(member)

	return family, nil
}

func (s *FamilyService) GetByUserID(userID uint) (*model.Family, error) {
	var member model.FamilyMember
	err := s.db.Where("user_id = ?", userID).First(&member).Error
	if err != nil {
		return nil, errors.New("你还没有加入任何家庭")
	}

	var family model.Family
	err = s.db.Preload("Members.User").First(&family, member.FamilyID).Error
	return &family, err
}

func (s *FamilyService) JoinByCode(userID uint, code string) error {
	var family model.Family
	if err := s.db.Where("invite_code = ?", code).First(&family).Error; err != nil {
		return errors.New("邀请码无效")
	}

	var count int64
	s.db.Model(&model.FamilyMember{}).Where("family_id = ? AND user_id = ?", family.ID, userID).Count(&count)
	if count > 0 {
		return errors.New("你已经是这个家庭的成员了")
	}

	member := &model.FamilyMember{
		FamilyID: family.ID,
		UserID:   userID,
		Role:     "member",
		JoinedAt: time.Now(),
	}
	return s.db.Create(member).Error
}

func (s *FamilyService) GetMembers(familyID uint) ([]model.FamilyMember, error) {
	var members []model.FamilyMember
	err := s.db.Preload("User").Where("family_id = ?", familyID).Find(&members).Error
	return members, err
}

func (s *FamilyService) RemoveMember(adminUserID, memberID uint) error {
	var member model.FamilyMember
	if err := s.db.First(&member, memberID).Error; err != nil {
		return errors.New("成员不存在")
	}

	// Check if requester is admin
	var admin model.FamilyMember
	err := s.db.Where("family_id = ? AND user_id = ? AND role = ?", member.FamilyID, adminUserID, "admin").
		First(&admin).Error
	if err != nil {
		return errors.New("无权操作")
	}

	if member.UserID == adminUserID {
		return errors.New("不能移除自己")
	}

	return s.db.Delete(&member).Error
}

func (s *FamilyService) RegenerateInviteCode(userID uint) (string, error) {
	var family model.Family
	if err := s.db.Where("creator_id = ?", userID).First(&family).Error; err != nil {
		return "", errors.New("你不是家庭创建者")
	}

	code, err := s.generateInviteCode()
	if err != nil {
		return "", err
	}

	s.db.Model(&family).Update("invite_code", code)
	return code, nil
}

func (s *FamilyService) generateInviteCode() (string, error) {
	const charset = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	code := make([]byte, 6)
	for i := range code {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		code[i] = charset[n.Int64()]
	}
	return string(code), nil
}
