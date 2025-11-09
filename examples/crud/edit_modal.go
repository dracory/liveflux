package main

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"github.com/dracory/hb"
	"github.com/dracory/liveflux"
)

const crudEditModalScript = `(function() {
    window.crudEditModal = {
        open: function(id, name, email, role) {
            const modal = document.getElementById('crud-edit-modal');
            const idEl = document.getElementById('crud-edit-id');
            const nameEl = document.getElementById('crud-edit-name');
            const emailEl = document.getElementById('crud-edit-email');
            const roleEl = document.getElementById('crud-edit-role');
            if (modal && idEl && nameEl && emailEl && roleEl) {
                idEl.value = id;
                nameEl.value = name;
                emailEl.value = email;
                roleEl.value = role;
                modal.style.display = 'flex';
            }
        },
        close: function() {
            const modal = document.getElementById('crud-edit-modal');
            if (modal) modal.style.display = 'none';
        }
    };
})();`

type EditUserModal struct {
	liveflux.Base
	UpdatedEvent map[string]any
}

func (c *EditUserModal) GetKind() string { return "users.edit_modal" }

func (c *EditUserModal) Mount(ctx context.Context, params map[string]string) error {
	return nil
}

func (c *EditUserModal) Handle(ctx context.Context, action string, form url.Values) error {
	if action == "update" {
		id, _ := strconv.Atoi(form.Get("id"))
		name := form.Get("name")
		email := form.Get("email")
		role := form.Get("role")
		if user, ok := repo.Update(id, name, email, role); ok {
			c.DispatchToKind("users.list", "user-updated", map[string]interface{}{
				"id":    user.ID,
				"name":  user.Name,
				"email": user.Email,
				"role":  user.Role,
				"flash": "Updated " + user.Name,
			})
			// prepare client-side browser event payload
			c.UpdatedEvent = map[string]any{
				"id":    user.ID,
				"name":  user.Name,
				"email": user.Email,
				"role":  user.Role,
				"flash": "Updated " + user.Name,
			}
		}
	}
	return nil
}

func (c *EditUserModal) Render(ctx context.Context) hb.TagInterface {
	// Header components
	headerTitle := hb.H4().Text("Edit User")
	closeButton := hb.Button().
		Type("button").
		Class("btn-close").
		Attr("onclick", "window.crudEditModal && window.crudEditModal.close();")
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
	editIdInput := hb.Input().
		Type("hidden").
		Name("id").
		ID("crud-edit-id")

	// Form inputs
	nameLabel := hb.Label().Class("form-label").Text("Name")
	nameInput := hb.Input().
		Type("text").
		Class("form-control").
		Name("name").
		ID("crud-edit-name").
		Required(true)
	nameField := hb.Div().Class("mb-3").
		Child(nameLabel).
		Child(nameInput)

	emailLabel := hb.Label().Class("form-label").Text("Email")
	emailInput := hb.Input().
		Type("email").
		Class("form-control").
		Name("email").
		ID("crud-edit-email").
		Required(true)
	emailField := hb.Div().Class("mb-3").
		Child(emailLabel).
		Child(emailInput)

	roleLabel := hb.Label().Class("form-label").Text("Role")
	roleInput := hb.Input().
		Type("text").
		Class("form-control").
		Name("role").
		ID("crud-edit-role").
		Required(true)
	roleField := hb.Div().Class("mb-3").
		Child(roleLabel).
		Child(roleInput)

	// Form body
	formBody := hb.Div().Class("crud-modal__body").
		Child(typeInput).
		Child(idInput).
		Child(editIdInput).
		Child(nameField).
		Child(emailField).
		Child(roleField)

	// Footer buttons
	cancelBtn := hb.Button().
		Type("button").
		Class("btn btn-secondary").
		Attr("onclick", "window.crudEditModal && window.crudEditModal.close();").
		Text("Cancel")
	updateBtn := hb.Button().
		Type("submit").
		Class("btn btn-primary").
		Data("flux-action", "update").
		Text("Update")
	footer := hb.Div().Class("crud-modal__footer").
		Child(cancelBtn).
		Child(updateBtn)

	// Form
	form := hb.Form().Method("post").
		Child(formBody).
		Child(footer)

	// Card
	card := hb.Div().Class("crud-modal__card").
		Child(header).
		Child(form)

	// Modal
	modal := hb.Div().ID("crud-edit-modal").Class("crud-modal").
		Child(card).
		Child(hb.Script(crudEditModalScript))

	// If a user was just updated, emit a browser event so lists can refresh immediately
	if c.UpdatedEvent != nil {
		p := c.UpdatedEvent
		script := hb.NewScript(fmt.Sprintf(`(function(){
  var data = { id: %d, name: '%s', email: '%s', role: '%s', flash: '%s' };
  if(window.liveflux && window.liveflux.dispatch){ window.liveflux.dispatch('user-updated', data); }
  else { window.dispatchEvent(new CustomEvent('user-updated', { detail: data })); }
})();`,
			p["id"].(int),
			jsString(p["name"].(string)),
			jsString(p["email"].(string)),
			jsString(p["role"].(string)),
			jsString(p["flash"].(string)),
		))
		modal = modal.Child(script)
		c.UpdatedEvent = nil
	}

	return c.Root(modal)
}
