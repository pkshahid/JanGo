package auth

import (
	"fmt"

	godjangohttp "github.com/pkshahid/JanGo/http"
	"github.com/pkshahid/JanGo/http/views"
)

// Standard implementation templates mapping
var (
	TemplateLogin             = "registration/login.html"
	TemplateLoggedOut         = "registration/logged_out.html"
	TemplatePasswordChange    = "registration/password_change_form.html"
	TemplatePasswordChangeDone = "registration/password_change_done.html"
)

// LoginView displays the login form and handles the login action.
type LoginView struct {
	views.TemplateView
}

func NewLoginView() *LoginView {
	return &LoginView{
		TemplateView: views.TemplateView{
			TemplateName: TemplateLogin,
		},
	}
}

func (v *LoginView) Dispatch(req *godjangohttp.Request) godjangohttp.Response {
	return v.BaseView.Dispatch(req, v)
}

func (v *LoginView) Get(req *godjangohttp.Request) godjangohttp.Response {
	if req.User != nil && req.User.IsAuthenticated() {
		// Default redirect to /
		return godjangohttp.NewRedirectResponse("/", false)
	}
	return v.TemplateView.Get(req)
}

func (v *LoginView) Post(req *godjangohttp.Request) godjangohttp.Response {
	username := req.POST.Get("username")
	password := req.POST.Get("password")

	user, err := Authenticate(username, password)
	if err == nil && user != nil {
		err = Login(req, user)
		if err == nil {
			next := req.POST.Get("next")
			if next == "" {
				next = req.GET.Get("next")
			}
			if next == "" {
				next = "/" // settings.LOGIN_REDIRECT_URL in a real app
			}
			return godjangohttp.NewRedirectResponse(next, false)
		}
	}

	// Login failed
	ctx := map[string]any{
		"form_errors": fmt.Sprintf("Please enter a correct username and password. Note that both fields may be case-sensitive."),
	}
	return godjangohttp.Render(req, v.TemplateName, ctx)
}

// LogoutView logs out the user and displays the 'You are logged out' message.
func LogoutView(req *godjangohttp.Request) godjangohttp.Response {
	err := Logout(req)
	if err != nil {
		return views.ServerError(req)
	}

	next := req.GET.Get("next")
	if next != "" {
		return godjangohttp.NewRedirectResponse(next, false)
	}

	return godjangohttp.Render(req, TemplateLoggedOut, nil)
}

// PasswordChangeView handles changing a user's password.
type PasswordChangeView struct {
	views.TemplateView
}

func NewPasswordChangeView() *PasswordChangeView {
	return &PasswordChangeView{
		TemplateView: views.TemplateView{
			TemplateName: TemplatePasswordChange,
		},
	}
}

func (v *PasswordChangeView) Dispatch(req *godjangohttp.Request) godjangohttp.Response {
	// Ensure user is logged in
	if req.User == nil || !req.User.IsAuthenticated() {
		return godjangohttp.NewRedirectResponse("/login/?next="+req.URL.RequestURI(), false)
	}
	return v.BaseView.Dispatch(req, v)
}

func (v *PasswordChangeView) Post(req *godjangohttp.Request) godjangohttp.Response {
	oldPassword := req.POST.Get("old_password")
	newPassword1 := req.POST.Get("new_password1")
	newPassword2 := req.POST.Get("new_password2")

	// 1. Verify old password
	user, err := Authenticate(req.User.Username(), oldPassword)
	if err != nil || user == nil {
		ctx := map[string]any{"form_errors": "Your old password was entered incorrectly."}
		return godjangohttp.Render(req, v.TemplateName, ctx)
	}

	// 2. Verify new passwords match
	if newPassword1 != newPassword2 {
		ctx := map[string]any{"form_errors": "The two password fields didn't match."}
		return godjangohttp.Render(req, v.TemplateName, ctx)
	}

	// 3. Update password in the database
	// A full implementation requires updating the model instance via the ORM.
	// We'd use `hashers.MakePassword(newPassword1)`
	// e.g.: user.(*AbstractUser).Password = hashers.MakePassword(newPassword1)
	//       orm.NewQuerySet[AbstractUser]().Filter(...).Update(...)

	return godjangohttp.NewRedirectResponse("/password_change/done/", false)
}

// PasswordChangeDoneView displays success message.
func PasswordChangeDoneView(req *godjangohttp.Request) godjangohttp.Response {
	if req.User == nil || !req.User.IsAuthenticated() {
		return godjangohttp.NewRedirectResponse("/login/?next="+req.URL.RequestURI(), false)
	}
	return godjangohttp.Render(req, TemplatePasswordChangeDone, nil)
}
