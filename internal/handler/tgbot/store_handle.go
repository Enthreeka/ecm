package tgbot

import (
	"context"
	"github.com/Enthreeka/tg-question-bot/internal/entity"
	store "github.com/Enthreeka/tg-question-bot/pkg/local_storage"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (b *Bot) isStateExist(userID int64) (*store.Data, bool) {
	data, exist := b.store.Read(userID)
	return data, exist
}

func (b *Bot) isStoreProcessing(ctx context.Context, update *tgbotapi.Update) (bool, error) {
	userID := update.Message.From.ID
	storeData, isExist := b.isStateExist(userID)
	if !isExist || storeData == nil {
		return false, nil
	}
	defer b.store.Delete(userID)

	return b.switchStoreData(ctx, update, storeData)
}

func (b *Bot) switchStoreData(ctx context.Context, update *tgbotapi.Update, storeData *store.Data) (bool, error) {
	var (
		err error
	)

	switch storeData.OperationType {
	case store.AdminCreate:
		err = b.userService.UpdateRoleByUsername(ctx, entity.AdminType, update.Message.Text)
		if err != nil {
			b.log.Error("isStoreExist::store.AdminCreate:UpdateRoleByUsername: %v", err)
		}
	case store.AdminDelete:
		err = b.userService.UpdateRoleByUsername(ctx, entity.UserType, update.Message.Text)
		if err != nil {
			b.log.Error("isStoreExist::store.AdminDelete:userRepo.UpdateRoleByUsername: %v", err)
		}
	default:
		return false, nil
	}

	if err == nil {
		b.response(storeData.OperationType, storeData.CurrentMsgID, storeData.PreferMsgID, update)
	}
	return true, err
}
