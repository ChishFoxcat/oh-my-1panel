package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/ChishFoxcat/oh-my-1panel/backend/app/dto"
	"github.com/ChishFoxcat/oh-my-1panel/backend/buserr"
	"github.com/ChishFoxcat/oh-my-1panel/backend/constant"
	"github.com/ChishFoxcat/oh-my-1panel/backend/global"
	"github.com/ChishFoxcat/oh-my-1panel/backend/init/session"
	"github.com/ChishFoxcat/oh-my-1panel/backend/utils/panel"
)

// AuthService 负责登录链路：面板设置读取、账号密码登录、动态验证码、会话与登出。
type AuthService struct{}

// NewAuthService 构造认证服务。
func NewAuthService() *AuthService {
	return &AuthService{}
}

// SystemInfo 读取面板公开信息，并在已登录时附带当前用户。
func (u *AuthService) SystemInfo(ctx context.Context, current *session.Session) (*dto.SystemInfo, error) {
	client, err := newPanelClient()
	if err != nil {
		return nil, err
	}
	defer client.Close()

	setting, err := client.LoginSetting(ctx)
	if err != nil {
		return nil, wrapPanelError(err)
	}
	info := &dto.SystemInfo{
		PanelName:   setting.PanelName,
		PanelURL:    global.CONF.PanelURL,
		Language:    setting.Language,
		Theme:       setting.Theme,
		NeedCaptcha: setting.NeedCaptcha,
		Version:     constant.Version,
		Logged:      current != nil,
	}
	if current != nil {
		info.User = toUserInfo(current.Name, current.Role)
	}
	return info, nil
}

// Captcha 获取面板登录验证码。
func (u *AuthService) Captcha(ctx context.Context) (*dto.CaptchaResponse, error) {
	client, err := newPanelClient()
	if err != nil {
		return nil, err
	}
	defer client.Close()

	captcha, err := client.Captcha(ctx)
	if err != nil {
		return nil, wrapPanelError(err)
	}
	return &dto.CaptchaResponse{CaptchaID: captcha.CaptchaID, ImagePath: captcha.ImagePath}, nil
}

// Login 使用面板账号密码登录；开启动态验证码时返回中间会话，等待 MFA 校验。
func (u *AuthService) Login(ctx context.Context, req dto.LoginRequest) (*dto.LoginResponse, *session.Session, error) {
	client, err := newPanelClient()
	if err != nil {
		return nil, nil, err
	}
	setting, err := client.LoginSetting(ctx)
	if err != nil {
		client.Close()
		return nil, nil, wrapPanelError(err)
	}
	info, err := client.Login(ctx, panel.LoginRequest{
		Name:      req.Name,
		Password:  req.Password,
		Captcha:   req.Captcha,
		CaptchaID: req.CaptchaID,
		Language:  setting.Language,
	})
	if err != nil {
		client.Close()
		return nil, nil, wrapPanelError(err)
	}
	if info.MfaStatus == constant.StatusEnable {
		pending, err := global.SESSIONS.CreatePending(client, info.Name, info.MfaSession)
		if err != nil {
			client.Close()
			return nil, nil, buserr.Internal("创建动态验证码会话失败：" + err.Error())
		}
		return &dto.LoginResponse{MfaRequired: true, MfaSession: pending.ID}, nil, nil
	}
	return u.establish(client, info)
}

// MFA 校验动态验证码并建立会话。
func (u *AuthService) MFA(ctx context.Context, req dto.MFARequest) (*dto.LoginResponse, *session.Session, error) {
	pending, ok := global.SESSIONS.GetPending(req.SessionID)
	if !ok {
		return nil, nil, buserr.Unauthorized("动态验证码会话已过期，请重新登录！")
	}
	info, err := pending.Panel.MFALogin(ctx, pending.MFASession, req.Code)
	if err != nil {
		return nil, nil, wrapPanelError(err)
	}
	global.SESSIONS.DropPending(pending.ID)
	return u.establish(pending.Panel, info)
}

// Current 读取面板当前用户信息，面板会话失效时同步销毁本地会话。
func (u *AuthService) Current(ctx context.Context, current *session.Session) (*dto.CurrentUser, error) {
	info, err := current.Panel.Current(ctx)
	if err != nil {
		if panel.IsSessionExpired(err) {
			global.SESSIONS.Drop(current.ID)
			return nil, buserr.Unauthorized("面板会话已过期，请重新登录！")
		}
		return nil, wrapPanelError(err)
	}
	return &dto.CurrentUser{
		Name:        info.Name,
		Role:        info.Role,
		IsAdmin:     info.Role == constant.RoleAdmin,
		MfaStatus:   info.MfaStatus,
		AuthSource:  info.AuthSource,
		Permissions: info.Permissions,
	}, nil
}

// Logout 注销面板会话并销毁本地会话。
func (u *AuthService) Logout(ctx context.Context, current *session.Session) {
	if err := current.Panel.Logout(ctx); err != nil {
		global.LOGGER.Warn("注销面板会话失败", "error", err)
	}
	global.SESSIONS.Drop(current.ID)
}

func (u *AuthService) establish(client *panel.Client, info *panel.UserLoginInfo) (*dto.LoginResponse, *session.Session, error) {
	current, err := global.SESSIONS.Create(client, info.Name, info.Role)
	if err != nil {
		client.Close()
		return nil, nil, buserr.Internal("创建会话失败：" + err.Error())
	}
	return &dto.LoginResponse{
		ExpiresAt: current.ExpiresAt.Format(time.RFC3339),
		User:      toUserInfo(current.Name, current.Role),
	}, current, nil
}

func newPanelClient() (*panel.Client, error) {
	client, err := panel.New(global.CONF.PanelURL, global.CONF.PanelEntrance, global.CONF.PanelSkipTLSVerify)
	if err != nil {
		return nil, buserr.Internal(err.Error())
	}
	return client, nil
}

func toUserInfo(name, role string) *dto.UserInfo {
	return &dto.UserInfo{Name: name, Role: role, IsAdmin: role == constant.RoleAdmin}
}

// wrapPanelError 把面板错误转换为可直接展示的业务错误。
func wrapPanelError(err error) error {
	var apiErr *panel.APIError
	if errors.As(err, &apiErr) {
		status := apiErr.Code
		if status < 400 || status > 599 {
			status = http.StatusBadGateway
		}
		return (&buserr.BusinessError{Status: status, Message: apiErr.Error()}).WithDetail(apiErr.Key)
	}
	return buserr.Internal(fmt.Sprintf("请求面板失败：%v", err)).WithStatus(http.StatusBadGateway)
}
