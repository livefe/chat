package handler

import (
	"app/internal/middleware"
	"app/internal/service"
	"app/internal/types"
	"app/pkg/logger"
	"app/pkg/response"

	"github.com/gin-gonic/gin"
)

type PostHandler struct {
	log     *logger.Logger
	postSvc *service.PostService
}

func NewPostHandler(log *logger.Logger, postSvc *service.PostService) *PostHandler {
	return &PostHandler{log: log, postSvc: postSvc}
}

func (h *PostHandler) Create(c *gin.Context) {
	var req types.CreatePostRequest
	if !bindJSON(c, &req) {
		return
	}
	id, err := h.postSvc.Create(c.Request.Context(), middleware.CurrentUserID(c), &req)
	if err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "已提交，等待审核", types.IDResponse{ID: id})
}

func (h *PostHandler) Update(c *gin.Context) {
	var req types.UpdatePostRequest
	if !bindJSON(c, &req) {
		return
	}
	if err := h.postSvc.Update(c.Request.Context(), middleware.CurrentUserID(c), &req); err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "已提交，等待审核", nil)
}

func (h *PostHandler) Withdraw(c *gin.Context) {
	var req types.IDRequest
	if !bindJSON(c, &req) {
		return
	}
	if err := h.postSvc.Withdraw(c.Request.Context(), middleware.CurrentUserID(c), req.ID); err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "已下架", nil)
}

func (h *PostHandler) DeleteMine(c *gin.Context) {
	var req types.IDRequest
	if !bindJSON(c, &req) {
		return
	}
	if err := h.postSvc.DeleteMine(c.Request.Context(), middleware.CurrentUserID(c), req.ID); err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "已删除", nil)
}

func (h *PostHandler) HideJoined(c *gin.Context) {
	var req types.IDRequest
	if !bindJSON(c, &req) {
		return
	}
	if err := h.postSvc.HideJoined(c.Request.Context(), middleware.CurrentUserID(c), req.ID); err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "已移出", nil)
}

func (h *PostHandler) List(c *gin.Context) {
	var req types.PostListRequest
	if !bindQuery(c, &req) {
		return
	}
	result, err := h.postSvc.ListPublic(c.Request.Context(), middleware.CurrentUserID(c), &req)
	if err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "", result)
}

func (h *PostHandler) MyList(c *gin.Context) {
	var req types.MyPostListRequest
	if !bindQuery(c, &req) {
		return
	}
	result, err := h.postSvc.ListMine(c.Request.Context(), middleware.CurrentUserID(c), &req)
	if err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "", result)
}

func (h *PostHandler) MyJoinedList(c *gin.Context) {
	var req types.MyPostListRequest
	if !bindQuery(c, &req) {
		return
	}
	result, err := h.postSvc.ListJoined(c.Request.Context(), middleware.CurrentUserID(c), &req)
	if err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "", result)
}

func (h *PostHandler) UserList(c *gin.Context) {
	var req types.UserPostListRequest
	if !bindQuery(c, &req) {
		return
	}
	result, err := h.postSvc.ListPublicByUser(c.Request.Context(), middleware.CurrentUserID(c), req.UserID, &req)
	if err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "", result)
}

func (h *PostHandler) Detail(c *gin.Context) {
	var req types.IDRequest
	if !bindQuery(c, &req) {
		return
	}
	detail, err := h.postSvc.Get(c.Request.Context(), middleware.CurrentUserID(c), req.ID)
	if err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "", detail)
}

func (h *PostHandler) Signup(c *gin.Context) {
	var req types.PostSignupRequest
	if !bindJSON(c, &req) {
		return
	}
	result, err := h.postSvc.Signup(c.Request.Context(), middleware.CurrentUserID(c), req.PostID)
	if err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "报名成功", result)
}

func (h *PostHandler) CancelSignup(c *gin.Context) {
	var req types.PostSignupRequest
	if !bindJSON(c, &req) {
		return
	}
	if err := h.postSvc.CancelSignup(c.Request.Context(), middleware.CurrentUserID(c), req.PostID); err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "已取消报名", nil)
}

func (h *PostHandler) GetCheckInCode(c *gin.Context) {
	var req types.PostSignupCheckInCodeRequest
	if !bindQuery(c, &req) {
		return
	}
	result, err := h.postSvc.GetCheckInCode(c.Request.Context(), middleware.CurrentUserID(c), req.PostID)
	if err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "", result)
}

func (h *PostHandler) CheckInSignup(c *gin.Context) {
	var req types.PostSignupCheckInRequest
	if !bindJSON(c, &req) {
		return
	}
	result, err := h.postSvc.CheckInSignup(c.Request.Context(), middleware.CurrentUserID(c), &req)
	if err != nil {
		response.WriteError(c, err)
		return
	}
	msg := "核验成功"
	if result != nil && result.AlreadyCheckedIn {
		msg = "已核验"
	}
	response.Success(c, msg, result)
}

func (h *PostHandler) PublisherSignupList(c *gin.Context) {
	var req types.AdminPostSignupListRequest
	if !bindQuery(c, &req) {
		return
	}
	result, err := h.postSvc.ListSignups(c.Request.Context(), middleware.CurrentUserID(c), &req)
	if err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "", result)
}

func (h *PostHandler) UpdateSignupRemark(c *gin.Context) {
	var req types.UpdateSignupRemarkRequest
	if !bindJSON(c, &req) {
		return
	}
	if err := h.postSvc.UpdateSignupRemark(c.Request.Context(), middleware.CurrentUserID(c), &req); err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "已保存", nil)
}

func (h *PostHandler) AdminList(c *gin.Context) {
	var req types.AdminPostListRequest
	if !bindQuery(c, &req) {
		return
	}
	result, err := h.postSvc.AdminList(c.Request.Context(), &req)
	if err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "", result)
}

func (h *PostHandler) AdminMap(c *gin.Context) {
	var req types.AdminPostMapQuery
	if !bindQuery(c, &req) {
		return
	}
	result, err := h.postSvc.AdminMap(c.Request.Context(), &req)
	if err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "", result)
}

func (h *PostHandler) AdminDetail(c *gin.Context) {
	var req types.IDRequest
	if !bindQuery(c, &req) {
		return
	}
	result, err := h.postSvc.AdminGet(c.Request.Context(), req.ID)
	if err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "", result)
}

func (h *PostHandler) AdminSignupList(c *gin.Context) {
	var req types.AdminPostSignupListRequest
	if !bindQuery(c, &req) {
		return
	}
	result, err := h.postSvc.AdminListSignups(c.Request.Context(), &req)
	if err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "", result)
}

func (h *PostHandler) AdminUserSignupList(c *gin.Context) {
	var req types.AdminUserSignupListRequest
	if !bindQuery(c, &req) {
		return
	}
	result, err := h.postSvc.AdminListUserSignups(c.Request.Context(), &req)
	if err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "", result)
}

func (h *PostHandler) AdminUpdateStatus(c *gin.Context) {
	var req types.UpdatePostStatusRequest
	if !bindJSON(c, &req) {
		return
	}
	if err := h.postSvc.AdminUpdateStatus(c.Request.Context(), middleware.CurrentEmployeeID(c), &req); err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "更新成功", nil)
}

func (h *PostHandler) AdminBatchUpdateStatus(c *gin.Context) {
	var req types.BatchUpdatePostStatusRequest
	if !bindJSON(c, &req) {
		return
	}
	result, err := h.postSvc.AdminBatchUpdateStatus(c.Request.Context(), middleware.CurrentEmployeeID(c), &req)
	if err != nil {
		response.WriteError(c, err)
		return
	}
	response.Success(c, "批量操作完成", result)
}
