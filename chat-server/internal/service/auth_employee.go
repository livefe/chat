package service

import (
	"context"
	"errors"
	"fmt"

	"app/internal/model"
	"app/internal/types"
	"app/pkg/auth"
	"app/pkg/errs"
	"app/pkg/logger"

	"gorm.io/gorm"
)

func (s *AuthService) LoginEmployee(ctx context.Context, req *types.EmployeeLoginRequest, ip, ua string) (*types.EmployeeLoginResponse, error) {
	record := &model.EmployeeLoginRecord{
		Mobile:    req.Mobile,
		IP:        ip,
		UserAgent: ua,
		LoginType: types.LoginTypeSMS,
	}
	defer func() {
		if err := s.employeeLoginRepo.Create(ctx, record); err != nil {
			s.log.Error(ctx, "创建员工登录记录失败", logger.String("mobile", req.Mobile), logger.Error(err))
		}
	}()

	if err := s.smsSvc.VerifyCode(ctx, req.Mobile, req.Code, types.SourceAdminLogin); err != nil {
		record.Remark = err.Error()
		s.log.Warn(ctx, "员工登录失败", logger.String("reason", err.Error()))
		return nil, err
	}

	employee, err := s.employeeRepo.GetByMobile(ctx, req.Mobile)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			record.Remark = "员工不存在"
			s.log.Warn(ctx, "员工登录失败", logger.String("reason", "员工不存在"))
			return nil, errs.Forbidden("无后台访问权限")
		}
		record.Remark = fmt.Sprintf("查询员工失败: %v", err)
		s.log.Error(ctx, "员工登录失败", logger.String("reason", "查询员工失败"), logger.Error(err))
		return nil, fmt.Errorf("系统错误，请稍后重试: %w", err)
	}

	if employee.Status != model.EmployeeStatusEnabled {
		record.Remark = "员工已被禁用"
		s.log.Warn(ctx, "员工登录失败", logger.String("reason", "员工已被禁用"), logger.Uint("employee_id", employee.ID))
		return nil, errs.Forbidden("账号已被禁用")
	}

	hasRole, err := s.employeeRoleRepo.HasAnyRole(ctx, employee.ID)
	if err != nil {
		record.Remark = fmt.Sprintf("查询员工角色失败: %v", err)
		s.log.Error(ctx, "员工登录失败", logger.String("reason", "查询员工角色失败"), logger.Uint("employee_id", employee.ID), logger.Error(err))
		return nil, fmt.Errorf("系统错误，请稍后重试: %w", err)
	}
	if !hasRole {
		record.Remark = "员工未分配角色"
		s.log.Warn(ctx, "员工登录失败", logger.String("reason", "员工未分配角色"), logger.Uint("employee_id", employee.ID))
		return nil, errs.Forbidden("无后台访问权限")
	}

	record.EmployeeID = employee.ID
	record.Status = model.SMSStatusSuccess

	pair, err := s.tokenAuth.GenerateTokenPair(auth.SubjectTypeEmployee, employee.ID)
	if err != nil {
		s.log.Error(ctx, "员工登录失败",
			logger.String("reason", "生成令牌失败"),
			logger.Uint("employee_id", employee.ID),
			logger.Error(err),
		)
		return nil, fmt.Errorf("登录失败，请稍后重试: %w", err)
	}
	if err := s.storeRefreshJTI(ctx, auth.SubjectTypeEmployee, employee.ID, pair.RefreshJTI); err != nil {
		s.log.Error(ctx, "员工登录失败",
			logger.String("reason", "存储刷新令牌失败"),
			logger.Uint("employee_id", employee.ID),
			logger.Error(err),
		)
		return nil, fmt.Errorf("登录失败，请稍后重试: %w", err)
	}

	departmentName := ""
	if dept, err := s.deptRepo.GetByID(ctx, employee.DepartmentID); err == nil {
		departmentName = dept.Name
	}

	return &types.EmployeeLoginResponse{
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
		Employee:     toEmployeeInfo(employee, departmentName),
	}, nil
}

func (s *AuthService) RefreshEmployee(ctx context.Context, refreshToken string) (*types.RefreshTokenResponse, error) {
	return s.refreshTokenPair(ctx, refreshToken, auth.SubjectTypeEmployee,
		func(ctx context.Context, subjectID uint) error {
			employee, err := s.employeeRepo.GetByID(ctx, subjectID)
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return errs.Unauthorized("登录已失效，请重新登录")
				}
				return fmt.Errorf("系统错误，请稍后重试: %w", err)
			}
			if employee.Status != model.EmployeeStatusEnabled {
				return errs.Forbidden("账号已被禁用")
			}
			return nil
		},
		func(ctx context.Context, subjectID uint) error {
			hasRole, err := s.employeeRoleRepo.HasAnyRole(ctx, subjectID)
			if err != nil {
				return fmt.Errorf("系统错误，请稍后重试: %w", err)
			}
			if !hasRole {
				return errs.Forbidden("无后台访问权限")
			}
			return nil
		},
	)
}
