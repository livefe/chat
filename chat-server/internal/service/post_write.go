package service

import (
	"math"
	"strings"
	"time"
	"unicode/utf8"

	"app/internal/model"
	"app/internal/types"
	"app/pkg/errs"
)

type normalizedPostFields struct {
	Title   string
	Content string
	StartAt int64
	EndAt   int64
	Extras  *model.PostExtra
	Pinned  bool
}

func validPostCategory(c string) bool {
	switch c {
	case model.PostCategoryEvent, model.PostCategoryNotice, model.PostCategoryIdle, model.PostCategoryHelp:
		return true
	default:
		return false
	}
}

func roundMoney(v float64) float64 {
	return math.Round(v*100) / 100
}

func postExtrasHasContent(e *model.PostExtra) bool {
	if e == nil {
		return false
	}
	return e.Amount > 0 ||
		e.Negotiable ||
		e.NeedSignup ||
		e.NeedCheckIn ||
		e.MaxParticipants > 0 ||
		e.AllowWaitlist ||
		e.SignupDeadlineAt > 0
}

func normalizePostWrite(
	category string,
	roomID uint,
	pinned bool,
	title, content string,
	startAt, endAt int64,
	extras *types.PostExtrasPayload,
) (*normalizedPostFields, error) {
	title = strings.TrimSpace(title)
	content = strings.TrimSpace(content)
	if title == "" {
		return nil, errs.Invalid("请填写标题")
	}
	if utf8.RuneCountInString(title) > 100 {
		return nil, errs.Invalid("标题过长")
	}

	out := &normalizedPostFields{
		Title:   title,
		Content: content,
		Pinned:  pinned && roomID > 0,
	}

	switch category {
	case model.PostCategoryEvent:
		if content == "" {
			return nil, errs.Invalid("请填写活动说明")
		}
		if startAt > 0 && endAt > 0 && endAt < startAt {
			return nil, errs.Invalid("结束时间不能早于开始时间")
		}
		out.StartAt = startAt
		out.EndAt = endAt
		if extras == nil {
			extras = &types.PostExtrasPayload{}
		}
		if err := validateEventExtras(startAt, endAt, extras); err != nil {
			return nil, err
		}
		ex := &model.PostExtra{}
		applyEventExtras(ex, extras)
		if postExtrasHasContent(ex) {
			out.Extras = ex
		}

	case model.PostCategoryNotice:
		if content == "" {
			return nil, errs.Invalid("请填写公告内容")
		}

	case model.PostCategoryIdle:
		if extras != nil {
			if extras.Amount < 0 {
				return nil, errs.Invalid("价格无效")
			}
			if extras.Amount == 0 && !extras.Negotiable {
				return nil, errs.Invalid("请填写价格或勾选面议")
			}
			ex := &model.PostExtra{
				Amount:     roundMoney(extras.Amount),
				Negotiable: extras.Negotiable,
			}
			out.Extras = ex
		}

	case model.PostCategoryHelp:
		if content == "" {
			return nil, errs.Invalid("请说明求助内容")
		}
		if extras == nil {
			break
		}
		switch extras.FeeType {
		case model.PostFeeReward:
			if extras.Amount <= 0 {
				return nil, errs.Invalid("请填写悬赏金额")
			}
			out.Extras = &model.PostExtra{
				FeeType:        model.PostFeeReward,
				Amount:         roundMoney(extras.Amount),
				FeeDescription: strings.TrimSpace(extras.FeeDescription),
			}
		case model.PostFeeCharge:
			return nil, errs.Invalid("求助不支持收费")
		}

	default:
		return nil, errs.Invalid("无效的动态分类")
	}

	return out, nil
}

func applyEventExtras(dest *model.PostExtra, src *types.PostExtrasPayload) {
	dest.NeedSignup = src.NeedSignup
	dest.NeedCheckIn = src.NeedSignup && src.NeedCheckIn
	if src.NeedSignup {
		dest.MaxParticipants = src.MaxParticipants
		dest.AllowWaitlist = src.MaxParticipants > 0 && src.AllowWaitlist
		dest.SignupDeadlineAt = src.SignupDeadlineAt
	}
	switch src.FeeType {
	case model.PostFeeCharge, model.PostFeeReward:
		dest.FeeType = src.FeeType
		dest.Amount = roundMoney(src.Amount)
		dest.FeeDescription = strings.TrimSpace(src.FeeDescription)
	default:
		dest.FeeType = model.PostFeeNone
	}
}

func validateEventExtras(startAt, endAt int64, extras *types.PostExtrasPayload) error {
	if extras.FeeType != model.PostFeeNone &&
		extras.FeeType != model.PostFeeCharge &&
		extras.FeeType != model.PostFeeReward {
		return errs.Invalid("无效的费用类型")
	}
	if (extras.FeeType == model.PostFeeCharge || extras.FeeType == model.PostFeeReward) && extras.Amount <= 0 {
		return errs.Invalid("请填写金额")
	}
	if !extras.NeedSignup && !extras.NeedCheckIn {
		return nil
	}
	if endAt <= 0 {
		return errs.Invalid("开启报名或核验时请设置结束时间")
	}
	if !extras.NeedSignup {
		return nil
	}
	deadline := extras.SignupDeadlineAt
	if deadline <= 0 {
		return nil
	}
	now := time.Now().Unix()
	if isSignupDeadlinePassed(deadline, now) {
		return errs.Invalid("报名截止时间不能早于当前时间")
	}
	if startAt > 0 && deadline > startAt {
		return errs.Invalid("报名截止不能晚于开始时间")
	}
	return nil
}
