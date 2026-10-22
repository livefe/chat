package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"app/internal/model"
	"app/internal/types"
	"app/pkg/database"
	"app/pkg/logger"

	"gorm.io/gorm"
)

const noticeCloseReason = "system_notice"

// NotifyUser 以系统消息写入用户最近一次客服会话；用户没有客服会话时创建一个已结案的会话承载通知，不进入排队。
func (s *SupportService) NotifyUser(ctx context.Context, userID uint, text string) error {
	st, err := s.supportStateRepo.GetLatestByUserID(ctx, userID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("查询客服会话失败: %w", err)
	}

	var convID uint
	var item *types.IMMessageItem
	if err := database.Transaction(ctx, s.db, func(ctx context.Context) error {
		if st != nil {
			convID = st.ConversationID
		} else {
			id, err := s.createNoticeConversation(ctx, userID)
			if err != nil {
				return err
			}
			convID = id
		}
		written, err := s.writeSystemMessage(ctx, convID, text)
		item = written
		return err
	}); err != nil {
		s.log.Error(ctx, "写入用户通知失败", logger.Uint("user_id", userID), logger.Error(err))
		return fmt.Errorf("发送通知失败: %w", err)
	}

	conv, err := s.imSvc.GetConversation(ctx, convID)
	if err == nil && conv.Status != model.IMConvStatusClosed {
		s.fanoutMessage(ctx, convID, model.IMPrincipalSystem, 0, item)
		return nil
	}
	_ = s.incrUserUnread(ctx, userID)
	if s.users != nil {
		payload, _ := json.Marshal(map[string]any{
			"type":  types.MsgTypeMessage,
			"scope": types.ScopeSupport,
			"data":  item,
		})
		s.users.DeliverToUser(ctx, userID, payload)
	}
	return nil
}

func (s *SupportService) createNoticeConversation(ctx context.Context, userID uint) (uint, error) {
	title := "官方客服"
	if cfg, err := s.loadConfig(ctx); err == nil && cfg.BrandName != "" {
		title = cfg.BrandName
	}
	now := time.Now().Unix()
	conv, err := s.imSvc.CreateSupportConversation(ctx, userID, title, now)
	if err != nil {
		return 0, err
	}
	if err := s.imSvc.SetConversationStatus(ctx, conv.ID, model.IMConvStatusClosed); err != nil {
		return 0, err
	}
	if err := s.supportStateRepo.Create(ctx, &model.IMSupportState{
		ConversationID: conv.ID,
		UserID:         userID,
		CloseAt:        now,
		Source:         "system",
		CloseReason:    noticeCloseReason,
	}); err != nil {
		return 0, err
	}
	return conv.ID, nil
}
