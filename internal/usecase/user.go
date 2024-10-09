package service

import (
	"context"
	"errors"
	"github.com/Enthreeka/tg-question-bot/internal/entity"
	"github.com/Enthreeka/tg-question-bot/internal/repo"
	"github.com/Enthreeka/tg-question-bot/pkg/logger"
	"github.com/Enthreeka/tg-question-bot/pkg/postgres"
	customMsg "github.com/Enthreeka/tg-question-bot/pkg/tg_bot_api"
)

type UserService interface {
	GetUserByID(ctx context.Context, id int64) (*entity.User, error)
	GetAllUsers(ctx context.Context) ([]entity.User, error)
	GetAllAdmin(ctx context.Context) ([]entity.User, error)

	CreateUserIFNotExist(ctx context.Context, user *entity.User) error

	UpdateRoleByUsername(ctx context.Context, role entity.UserRole, username string) error

	// QUESTION domain
	CreateQuestion(ctx context.Context, userID int64, text string) error
}

type userService struct {
	userRepo repo.UserRepo
	log      *logger.Logger
	pg       *postgres.Postgres
	tgMsg    customMsg.Message
}

func NewUserService(
	userRepo repo.UserRepo,
	log *logger.Logger,
	tgMsg customMsg.Message,
	pg *postgres.Postgres,
) (UserService, error) {
	if userRepo == nil {
		return nil, errors.New("userRepo is nil")
	}
	if log == nil {
		return nil, errors.New("log is nil")
	}
	if pg == nil {
		return nil, errors.New("pg is nil")
	}
	if tgMsg == nil {
		return nil, errors.New("tgMsg is nil")
	}

	return &userService{
		userRepo: userRepo,
		log:      log,
		pg:       pg,
		tgMsg:    tgMsg,
	}, nil
}

func (u *userService) GetUserByID(ctx context.Context, id int64) (*entity.User, error) {
	return u.userRepo.GetUserByID(ctx, id)
}

func (u *userService) GetAllAdmin(ctx context.Context) ([]entity.User, error) {
	return u.userRepo.GetAllAdmin(ctx)
}

func (u *userService) CreateUserIFNotExist(ctx context.Context, user *entity.User) error {
	isExist, err := u.userRepo.IsUserExistByUserID(ctx, user.ID)
	if err != nil {
		u.log.Error("userRepo.IsUserExistByUsernameTg: failed to check user: %v", err)
		return err
	}

	if !isExist {
		u.log.Info("Get user: %s", user.String())
		err := u.userRepo.CreateUser(ctx, user)
		if err != nil {
			u.log.Error("userRepo.CreateUser: failed to create user: %v", err)
			return err
		}
	}

	return nil
}

func (u *userService) GetAllUsers(ctx context.Context) ([]entity.User, error) {
	return u.userRepo.GetAllUsers(ctx)
}

func (u *userService) UpdateRoleByUsername(ctx context.Context, role entity.UserRole, username string) error {
	return u.userRepo.UpdateRoleByUsername(ctx, role, username)
}

func (u *userService) CreateQuestion(ctx context.Context, userID int64, text string) error {
	_, err := u.pg.Pool.Exec(ctx,
		"INSERT INTO question ( user_id,question) VALUES ($1, $2)",
		userID, text)
	if err != nil {
		u.log.Error("failed to insert question: ", err)
		return nil
	}

	if _, err := u.tgMsg.SendNewMessage(userID, nil, "Я получил ваше сообщение и отправил его аналитикам"); err != nil {
		u.log.Error("failed to send new message: ", err)
		return nil
	}

	return nil
}
