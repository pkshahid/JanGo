package app

import (
	"github.com/godjango/godjango/forms"
)

// CommentForm uses the ModelForm pattern to bind to Comment.
type CommentForm struct {
	forms.ModelForm
}

func NewCommentForm() *CommentForm {
	mf, _ := forms.NewModelForm(&Comment{}, []string{"Author", "Body"}, nil)
	return &CommentForm{ModelForm: *mf}
}
