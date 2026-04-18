package auth

import (
	godjangohttp "github.com/pkshahid/JanGo/http"
	"github.com/pkshahid/JanGo/http/views"
)

var (
	TemplatePasswordReset        = "registration/password_reset_form.html"
	TemplatePasswordResetDone    = "registration/password_reset_done.html"
	TemplatePasswordResetConfirm = "registration/password_reset_confirm.html"
)

// PasswordResetView renders the form to enter an email for password reset.
type PasswordResetView struct {
	views.TemplateView
}

func NewPasswordResetView() *PasswordResetView {
	return &PasswordResetView{
		TemplateView: views.TemplateView{
			TemplateName: TemplatePasswordReset,
		},
	}
}

func (v *PasswordResetView) Dispatch(req *godjangohttp.Request) godjangohttp.Response {
	return v.BaseView.Dispatch(req, v)
}

func (v *PasswordResetView) Post(req *godjangohttp.Request) godjangohttp.Response {
	email := req.POST.Get("email")
	_ = email // Look up user, generate token, send email in full implementation

	return godjangohttp.NewRedirectResponse("/password_reset/done/", false)
}

// PasswordResetDoneView displays success message for sent email.
func PasswordResetDoneView(req *godjangohttp.Request) godjangohttp.Response {
	return godjangohttp.Render(req, TemplatePasswordResetDone, nil)
}

// PasswordResetConfirmView renders the form for entering a new password using a token.
type PasswordResetConfirmView struct {
	views.TemplateView
}

func NewPasswordResetConfirmView() *PasswordResetConfirmView {
	return &PasswordResetConfirmView{
		TemplateView: views.TemplateView{
			TemplateName: TemplatePasswordResetConfirm,
		},
	}
}

func (v *PasswordResetConfirmView) Dispatch(req *godjangohttp.Request) godjangohttp.Response {
	// A full implementation requires parsing `uidb64` and `token` from kwargs/ResolverMatch.
	return v.BaseView.Dispatch(req, v)
}

func (v *PasswordResetConfirmView) Post(req *godjangohttp.Request) godjangohttp.Response {
	newPassword1 := req.POST.Get("new_password1")
	newPassword2 := req.POST.Get("new_password2")

	if newPassword1 != newPassword2 {
		ctx := map[string]any{"form_errors": "The two password fields didn't match."}
		return godjangohttp.Render(req, v.TemplateName, ctx)
	}

	// Update password logic via ORM here.

	return godjangohttp.NewRedirectResponse("/login/", false)
}
