package service

import (
	"context"
	"encoding/json"
	"strings"

	"app/pkg/errs"
)

type messagePostContent struct {
	PostID uint   `json:"postId"`
	Title  string `json:"title"`
	Cover  string `json:"cover,omitempty"`
}

func parsePostCardID(raw string) (uint, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" || len(raw) > 1000 {
		return 0, errs.Invalid("动态无效")
	}
	var in struct {
		PostID uint `json:"postId"`
	}
	if err := json.Unmarshal([]byte(raw), &in); err != nil || in.PostID == 0 {
		return 0, errs.Invalid("动态无效")
	}
	return in.PostID, nil
}

func (s *MessageService) normalizePostContent(ctx context.Context, raw string) (string, error) {
	postID, err := parsePostCardID(raw)
	if err != nil {
		return "", err
	}
	id, title, cover, err := s.postSvc.MessageCardSnapshot(ctx, postID)
	if err != nil {
		return "", err
	}
	out, err := json.Marshal(messagePostContent{
		PostID: id,
		Title:  title,
		Cover:  cover,
	})
	if err != nil {
		return "", errs.Invalid("动态无效")
	}
	return string(out), nil
}

func messagePostQuotePreview(content string) string {
	var p messagePostContent
	if err := json.Unmarshal([]byte(strings.TrimSpace(content)), &p); err != nil {
		return "[动态]"
	}
	title := strings.TrimSpace(p.Title)
	if title == "" {
		return "[动态]"
	}
	label := "[动态] " + title
	runes := []rune(label)
	if len(runes) > 80 {
		return string(runes[:80]) + "…"
	}
	return label
}
