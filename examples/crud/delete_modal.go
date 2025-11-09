package main

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"github.com/dracory/hb"
	"github.com/dracory/liveflux"
)

const crudDeleteModalScript = `(function() {
    window.crudDeleteModal = {
        open: function(id, name) {
            const modal = document.getElementById('crud-delete-modal');
            const idEl = document.getElementById('crud-delete-id');
            const nameEl = document.getElementById('crud-delete-name');
            if (modal && idEl && nameEl) {
                idEl.value = id;
                nameEl.textContent = name;
                modal.style.display = 'flex';
            }
        },
        close: function() {
            const modal = document.getElementById('crud-delete-modal');
            if (modal) modal.style.display = 'none';
        }
    };
})();`

type DeleteUserModal struct {
	liveflux.Base
	DeletedEvent map[string]any
}

func (c *DeleteUserModal) GetKind() string { return "users.delete_modal" }

func (c *DeleteUserModal) Mount(ctx context.Context, params map[string]string) error {
	return nil
}

func (c *DeleteUserModal) Handle(ctx context.Context, action string, form url.Values) error {
	if action == "delete" {
		id, _ := strconv.Atoi(form.Get("id"))
		if user, ok := repo.Delete(id); ok {
			c.DispatchToKind("users.list", "user-deleted", map[string]interface{}{
				"id":    user.ID,
				"name":  user.Name,
				"email": user.Email,
				"role":  user.Role,
				"flash": "Removed " + user.Name,
			})
			// prepare client-side browser event payload
			c.DeletedEvent = map[string]any{
				"id":    user.ID,
				"name":  user.Name,
				"email": user.Email,
				"role":  user.Role,
				"flash": "Removed " + user.Name,
			}
		}
	}
	return nil
}

func (c *DeleteUserModal) Render(ctx context.Context) hb.TagInterface {
	// Header components
	headerTitle := hb.H4().Text("Delete User")
	closeButton := hb.Button().
		Type("button").
		Class("btn-close").
		Attr("onclick", "window.crudDeleteModal && window.crudDeleteModal.close();")
	header := hb.Div().Class("crud-modal__header").
		Child(headerTitle).
		Child(closeButton)

	// Hidden inputs
	typeInput := hb.Input().
		Type("hidden").
		Name("liveflux_component_kind").
		Value(c.GetKind())
	idInput := hb.Input().
		Type("hidden").
		Name("liveflux_component_id").
		Value(c.GetID())
	deleteIdInput := hb.Input().
		Type("hidden").
		Name("id").
		ID("crud-delete-id")

	// Form body
	formBody := hb.Div().Class("crud-modal__body").
		Child(typeInput).
		Child(idInput).
		Child(deleteIdInput).
		Child(hb.P().Text("Are you sure you want to delete this user?")).
		Child(hb.P().Class("fw-bold").ID("crud-delete-name"))

	// Footer buttons
	cancelBtn := hb.Button().
		Type("button").
		Class("btn btn-secondary").
		Attr("onclick", "window.crudDeleteModal && window.crudDeleteModal.close();").
		Text("Cancel")
	deleteBtn := hb.Button().
		Type("submit").
		Class("btn btn-danger").
		Attr(liveflux.DataFluxAction, "delete").
		Text("Delete")
	footer := hb.Div().Class("crud-modal__footer").
		Child(cancelBtn).
		Child(deleteBtn)

	// Form
	form := hb.Form().Method("post").
		Child(formBody).
		Child(footer)

	// Card
	card := hb.Div().Class("crud-modal__card").
		Child(header).
		Child(form)

	// Modal
	modal := hb.Div().ID("crud-delete-modal").Class("crud-modal").
		Child(card).
		Child(hb.Script(crudDeleteModalScript))

	// If a user was just deleted, emit a browser event so lists can refresh immediately
	if c.DeletedEvent != nil {
		p := c.DeletedEvent
		script := hb.NewScript(fmt.Sprintf(`(function(){
  var data = { id: %d};
  liveflux.dispatch('user-deleted', data);
})();`,
			p["id"].(int),
		))
		modal = modal.Child(script)
		c.DeletedEvent = nil
	}

	return c.Root(modal)
}
