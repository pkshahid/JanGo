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
	form := NewSetPasswordForm(nil)
	postData := make(map[string]any)
	for k, v := range req.POST {
		if len(v) > 0 {
			postData[k] = v[0]
		}
	}
	form.Bind(postData, nil)

	if !form.IsValid() {
		ctx := map[string]any{
			"form":        form,
			"form_errors": FormErrorsToString(&form.Form),
		}
		return godjangohttp.Render(req, v.TemplateName, ctx)
	}

	// Update password logic via ORM here.
	// A full implementation would resolve the user from uidb64/token,
	// then set: user.(*AbstractUser).Password = hashers.MakePassword(newPassword1)

	return godjangohttp.NewRedirectResponse("/login/", false)
}
