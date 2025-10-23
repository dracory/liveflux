package main

import (
	"context"
	"net/url"

	"github.com/dracory/hb"
	"github.com/dracory/liveflux"
)

const crudCreateModalScript = `(function() {
    window.crudCreateModal = {
        open: function() {
            const modal = document.getElementById('crud-create-modal');
            if (modal) modal.style.display = 'flex';
        },
        close: function() {
            const modal = document.getElementById('crud-create-modal');
            if (modal) modal.style.display = 'none';
        }
    };
})();`

type CreateUserModal struct {
	liveflux.Base
}

func (c *CreateUserModal) GetAlias() string { return "users.create_modal" }

func (c *CreateUserModal) Mount(ctx context.Context, params map[string]string) error {
	return nil
}

func (c *CreateUserModal) Handle(ctx context.Context, action string, form url.Values) error {
	if action == "create" {
		name := form.Get("name")
		email := form.Get("email")
		role := form.Get("role")
		user := repo.Create(name, email, role)
		c.DispatchTo("users.list", "user-created", map[string]interface{}{
			"id":    user.ID,
			"name":  user.Name,
			"email": user.Email,
			"role":  user.Role,
			"flash": "Added " + user.Name,
		})
	}
	return nil
}

func (c *CreateUserModal) Render(ctx context.Context) hb.TagInterface {
	// Header components
	headerTitle := hb.H4().Text("Add New User")
	closeButton := hb.Button().
		Type("button").
		Class("btn-close").
		Attr("onclick", "window.crudCreateModal && window.crudCreateModal.close();")
	header := hb.Div().Class("crud-modal__header").
		Child(headerTitle).
		Child(closeButton)

	// Form inputs
	nameLabel := hb.Label().Class("form-label").Text("Name")
	nameInput := hb.Input().
		Type("text").
		Class("form-control").
		Name("name").
		Required(true)
	nameField := hb.Div().Class("mb-3").
		Child(nameLabel).
		Child(nameInput)

	emailLabel := hb.Label().Class("form-label").Text("Email")
	emailInput := hb.Input().
		Type("email").
		Class("form-control").
		Name("email").
		Required(true)
	emailField := hb.Div().Class("mb-3").
		Child(emailLabel).
		Child(emailInput)

	roleLabel := hb.Label().Class("form-label").Text("Role")
	roleInput := hb.Input().
		Type("text").
		Class("form-control").
		Name("role").
		Required(true)
	roleField := hb.Div().Class("mb-3").
		Child(roleLabel).
		Child(roleInput)

	// Hidden inputs
	typeInput := hb.Input().
		Type("hidden").
		Name("liveflux_component_type").
		Value(c.GetAlias())
	idInput := hb.Input().
		Type("hidden").
		Name("liveflux_component_id").
		Value(c.GetID())

	// Form body
	formBody := hb.Div().Class("crud-modal__body").
		Child(typeInput).
		Child(idInput).
		Child(nameField).
		Child(emailField).
		Child(roleField)

	// Footer buttons
	cancelBtn := hb.Button().
		Type("button").
		Class("btn btn-secondary").
		Attr("onclick", "window.crudCreateModal && window.crudCreateModal.close();").
		Text("Cancel")
	submitBtn := hb.Button().
		Type("submit").
		Class("btn btn-primary").
		Data("flux-action", "create").
		Text("Create")
	footer := hb.Div().Class("crud-modal__footer").
		Child(cancelBtn).
		Child(submitBtn)

	// Form
	form := hb.Form().Method("post").
		Child(formBody).
		Child(footer)

	// Card
	card := hb.Div().Class("crud-modal__card").
		Child(header).
		Child(form)

	// Modal
	modal := hb.Div().ID("crud-create-modal").Class("crud-modal").
		Child(card).
		Child(hb.Script(crudCreateModalScript))

	return c.Root(modal)
}
