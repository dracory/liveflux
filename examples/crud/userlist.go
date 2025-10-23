package main

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/dracory/hb"
	"github.com/dracory/liveflux"
)

type UserList struct {
	liveflux.Base
	Query string
	Flash string
}

func (c *UserList) GetAlias() string { return "users.list" }

func (c *UserList) Mount(ctx context.Context, params map[string]string) error {
	if q, ok := params["q"]; ok && c.Query == "" {
		c.Query = strings.TrimSpace(q)
	}
	liveflux.RegisterEventListeners(c, c.GetEventDispatcher())
	return nil
}

func (c *UserList) Handle(ctx context.Context, action string, form url.Values) error {
	switch action {
	case "filter":
		c.Query = strings.TrimSpace(form.Get("search"))
	case "clear":
		c.Query = ""
	case "dismiss_flash":
		c.Flash = ""
	case "create_modal_open":
		c.DispatchToAlias("users.create_modal", "open")
	}
	return nil
}

func (c *UserList) Render(ctx context.Context) hb.TagInterface {
	users := repo.List()
	if c.Query != "" {
		q := strings.ToLower(c.Query)
		filtered := users[:0]
		for _, u := range users {
			if strings.Contains(strings.ToLower(u.Name), q) ||
				strings.Contains(strings.ToLower(u.Email), q) ||
				strings.Contains(strings.ToLower(u.Role), q) {
				filtered = append(filtered, u)
			}
		}
		users = filtered
	}

	searchInput := hb.Input().Type("search").Class("form-control").
		ID("crud-search").Name("search").Placeholder("Search...").Value(c.Query)

	filterBtn := hb.Button().Type("submit").Class("btn btn-primary").
		Attr(liveflux.DataFluxAction, "filter").Text("Search")

	clearBtn := hb.Button().Type("button").Class("btn btn-outline-secondary").
		Attr("onclick", "document.getElementById('crud-search').value='';window.__lw && window.__lw.clickSubmit && window.__lw.clickSubmit(this);").
		Attr(liveflux.DataFluxAction, "clear").
		Text("Clear")

	createBtn := hb.Button().
		Type("button").
		Class("btn btn-success").
		Attr(liveflux.DataFluxAction, "create_modal_open").
		//Attr("onclick", "window.crudCreateModal && window.crudCreateModal.open();").
		Text("Add User")

	form := hb.Form().Class("mb-3").
		Method("post").
		Child(
			hb.Div().Class("input-group").
				Child(searchInput).
				Child(
					hb.Div().Class("input-group-append").
						Child(filterBtn).
						Child(clearBtn).
						Child(createBtn),
				),
		)

	table := hb.Table().Class("table").
		Child(hb.Thead().Child(
			hb.Tr().
				Child(hb.Th().Text("Name")).
				Child(hb.Th().Text("Email")).
				Child(hb.Th().Text("Role")).
				Child(hb.Th().Text("Actions")),
		)).
		Child(hb.Tbody().Child(
			c.renderUsers(users),
		))

	body := hb.Div()
	if c.Flash != "" {
		body = body.Child(
			hb.Div().Class("alert alert-success d-flex align-items-center justify-content-between").
				Child(hb.Span().Text(c.Flash)).
				Child(hb.Button().
					Type("button").Class("btn-close").Attr("aria-label", "Close").
					Attr(liveflux.DataFluxAction, "dismiss_flash")),
		)
	}
	body = body.
		Child(form).
		Child(table)

	return c.Root(body)
}

func (c *UserList) renderUsers(users []User) hb.TagInterface {
	tbody := hb.Tbody()
	for _, u := range users {
		// Create edit button
		editBtn := hb.Button().
			Type("button").
			Class("btn btn-sm btn-outline-primary").
			Attr("onclick", fmt.Sprintf(
				"window.crudEditModal && window.crudEditModal.open(%d, '%s', '%s', '%s');",
				u.ID, jsString(u.Name), jsString(u.Email), jsString(u.Role))).
			Text("Edit")

		// Create delete button
		deleteBtn := hb.Button().
			Type("button").
			Class("btn btn-sm btn-outline-danger").
			Attr("onclick", fmt.Sprintf(
				"window.crudDeleteModal && window.crudDeleteModal.open(%d, '%s');", u.ID, jsString(u.Name))).
			Text("Delete")

		// Action cell
		actionsCell := hb.Td().
			Child(editBtn).
			Child(deleteBtn)

		// Table row
		row := hb.Tr().
			Child(hb.Td().Text(u.Name)).
			Child(hb.Td().Text(u.Email)).
			Child(hb.Td().Text(u.Role)).
			Child(actionsCell)

		tbody.Child(row)
	}
	return tbody
}

func (c *UserList) OnUserCreated(ctx context.Context, event liveflux.Event) error {
	c.applyUserEvent(event)
	return nil
}

func (c *UserList) OnUserUpdated(ctx context.Context, event liveflux.Event) error {
	c.applyUserEvent(event)
	return nil
}

func (c *UserList) OnUserDeleted(ctx context.Context, event liveflux.Event) error {
	c.applyUserEvent(event)
	return nil
}

func (c *UserList) applyUserEvent(event liveflux.Event) {
	if flash, ok := event.Data["flash"].(string); ok {
		c.Flash = flash
	} else {
		c.Flash = ""
	}
}
