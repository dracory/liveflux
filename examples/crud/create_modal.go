package main

import (
	"context"
	"log"
	"net/url"

	"github.com/dracory/hb"
	"github.com/dracory/liveflux"
)

type CreateUserModal struct {
	liveflux.Base
	Open bool
}

func (c *CreateUserModal) GetAlias() string {
	return "users.create_modal"
}

func (c *CreateUserModal) Mount(ctx context.Context, params map[string]string) error {
	liveflux.RegisterEventListeners(c, c.GetEventDispatcher())
	return nil
}

func (c *CreateUserModal) OnOpen(ctx context.Context, event liveflux.Event) error {
	log.Println("OnOpen event received")
	c.Open = true
	return nil
}

func (c *CreateUserModal) Handle(ctx context.Context, action string, form url.Values) error {
	if action == "create" {
		name := form.Get("name")
		email := form.Get("email")
		role := form.Get("role")
		user := repo.Create(name, email, role)
		c.DispatchToAlias("users.list", "user-created", map[string]any{
			"id":    user.ID,
			"name":  user.Name,
			"email": user.Email,
			"role":  user.Role,
			"flash": "Added " + user.Name,
		})
		c.Open = false
	}
	if action == "open" {
		log.Println("Opening modal")
		c.Open = true
	}
	if action == "close" {
		log.Println("Closing modal")
		c.Open = false
	}
	return nil
}

func (c *CreateUserModal) initScript() hb.TagInterface {
	alias := c.GetAlias()
	id := c.GetID()
	return hb.Script(`(function(){
const expectedAlias = '` + alias + `';
const expectedId = '` + id + `';

function setup(root){
	if(!root.$wire){
		console.warn('Modal create script: $wire not yet available on root');
		return;
	}
	// Listen for events
	root.$wire.on('open', function(){ root.$wire.call('open'); });
	root.$wire.on('close', function(){ root.$wire.call('close'); });
}

function init() {
	const root = liveflux.findComponent(expectedAlias, expectedId);
	if(!root){
		console.error('Modal create script: matching root not found for', expectedAlias, expectedId);
		return;
	}
	if(root.$wire){ setup(root); }
	else { document.addEventListener('livewire:init', setup, { once: true }); }
}

// wait for other components to initialize
setTimeout(init, 100);
    })();`)
}

func (c *CreateUserModal) Render(ctx context.Context) hb.TagInterface {
	// Header components
	headerTitle := hb.H4().Text("Add New User")
	closeButton := hb.Button().
		Type("button").
		Class("btn-close").
		Attr(liveflux.DataFluxAction, "close")

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
		Attr(liveflux.DataFluxAction, "close").
		Text("Cancel")

	submitBtn := hb.Button().
		Type("submit").
		Class("btn btn-primary").
		Attr(liveflux.DataFluxAction, "create").
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
	modal := hb.Div().ID("crud-create-modal").Class("crud-modal")
	if c.Open {
		modal = modal.Attr("style", "display: flex;")
	} else {
		modal = modal.Attr("style", "display: none;")
	}
	modal = modal.
		Child(card).
		Child(c.initScript())

	return c.Root(modal)
}
