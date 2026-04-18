package signals

import (
	"github.com/pkshahid/JanGo/core/signals"
)

var (
	// PreInit is called at the beginning of a model's __init__() method.
	PreInit = signals.NewSignal("pre_init")

	// PostInit is called at the end of a model's __init__() method.
	PostInit = signals.NewSignal("post_init")

	// PreSave is called at the beginning of a model's save() method.
	// kwargs: "instance", "created" (bool), "update_fields"
	PreSave = signals.NewSignal("pre_save")

	// PostSave is called at the end of a model's save() method.
	// kwargs: "instance", "created" (bool), "update_fields"
	PostSave = signals.NewSignal("post_save")

	// PreDelete is called at the beginning of a model's delete() method or a QuerySet's delete().
	// kwargs: "instance"
	PreDelete = signals.NewSignal("pre_delete")

	// PostDelete is called at the end of a model's delete() method or a QuerySet's delete().
	// kwargs: "instance"
	PostDelete = signals.NewSignal("post_delete")

	// M2MChanged is sent when a ManyToManyField is changed on a model instance.
	// kwargs: "action" (string), "pk_set" ([]any), "model"
	M2MChanged = signals.NewSignal("m2m_changed")

	// PreMigrate is sent by the migrate command before it starts to migrate an app.
	PreMigrate = signals.NewSignal("pre_migrate")

	// PostMigrate is sent by the migrate command after it completes migrating an app.
	PostMigrate = signals.NewSignal("post_migrate")
)

// Connect is a shortcut to attach a receiver to an ORM signal.
// It allows Django-style decorator simulation:
// e.g. signals.Connect(signals.PostSave, myFunc, MyModel{})
func Connect(sig *signals.Signal, receiver signals.ReceiverFunc, sender any) error {
	return sig.Connect(receiver, sender, false) // Default synchronous execution
}
