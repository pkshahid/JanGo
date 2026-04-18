package admin

import (
	"time"

	"github.com/pkshahid/JanGo/auth"
	"github.com/pkshahid/JanGo/orm"
)

func init() {
	orm.Register(&LogEntry{})
}

const (
	Addition = 1
	Change   = 2
	Deletion = 3
)

// LogEntry tracks adds, changes, and deletions in the admin interface.
type LogEntry struct {
	orm.Model
	ActionTime    time.Time          `gd:"DateTimeField,auto_now_add=true"`
	User          *auth.AbstractUser `gd:"ForeignKey,to=auth.User,on_delete=CASCADE"`
	ContentTypeID uint64             `gd:"IntegerField,blank=true,null=true"` // Would point to a contenttypes model in full implementation
	ObjectID      string             `gd:"TextField,blank=true,null=true"`
	ObjectRepr    string             `gd:"CharField,max_length=200"`
	ActionFlag    int                `gd:"SmallIntegerField"`
	ChangeMessage string             `gd:"TextField,blank=true"`
}

func (l *LogEntry) ModelMeta() *orm.Meta {
	return &orm.Meta{
		DbTable:           "django_admin_log",
		Ordering:          []string{"-ActionTime"},
		VerboseName:       "log entry",
		VerboseNamePlural: "log entries",
	}
}

// IsAddition returns true if the log entry is an addition.
func (l *LogEntry) IsAddition() bool {
	return l.ActionFlag == Addition
}

// IsChange returns true if the log entry is a change.
func (l *LogEntry) IsChange() bool {
	return l.ActionFlag == Change
}

// IsDeletion returns true if the log entry is a deletion.
func (l *LogEntry) IsDeletion() bool {
	return l.ActionFlag == Deletion
}
