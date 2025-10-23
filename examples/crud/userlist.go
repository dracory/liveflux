package main

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/dracory/hb"
	"github.com/dracory/liveflux"
)

const userListRefreshScript = `(function(){
  window.crudRefreshUsers = function(message){
    var form = document.querySelector('[data-crud-userlist-form]');
    if(!form){ return; }
    var flashInput = form.querySelector('input[name="flash"]');
    if(!flashInput){
      flashInput = document.createElement('input');
      flashInput.type = 'hidden';
      flashInput.name = 'flash';
      form.appendChild(flashInput);
    }
    flashInput.value = message || '';
    var submitBtn = form.querySelector('[data-flux-action="refresh"]');
    if(!submitBtn){ return; }
    window.__lw && window.__lw.clickSubmit
      ? window.__lw.clickSubmit(submitBtn)
      : submitBtn.click();
  };
})();`

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
	return nil
}

func (c *UserList) Handle(ctx context.Context, action string, form url.Values) error {
	switch action {
	case "filter":
		c.Query = strings.TrimSpace(form.Get("search"))
	case "clear":
		c.Query = ""
	case "refresh":
		c.Query = strings.TrimSpace(form.Get("search"))
		c.Flash = strings.TrimSpace(form.Get("flash"))
	case "dismiss_flash":
		c.Flash = ""
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
		Data("flux-action", "filter").Text("Search")

	clearBtn := hb.Button().Type("button").Class("btn btn-outline-secondary").
		Attr("onclick", "document.getElementById('crud-search').value='';window.__lw && window.__lw.clickSubmit && window.__lw.clickSubmit(this);").Data("flux-action", "clear").Text("Clear")

	createBtn := hb.Button().Type("button").
		Class("btn btn-success").Attr("onclick", "window.crudCreateModal && window.crudCreateModal.open();").Text("Add User")

	refreshBtn := hb.Button().Type("button").Class("d-none").
		Data("flux-action", "refresh")

	form := hb.Form().Class("mb-3").
		Method("post").
		Attr("data-crud-userlist-form", "1").
		Child(
			hb.Div().Class("input-group").
				Child(searchInput).
				Child(
					hb.Div().Class("input-group-append").
						Child(filterBtn).
						Child(clearBtn).
						Child(createBtn).
						Child(refreshBtn),
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

	return c.Root(
		hb.Div().
			Child(form).
			Child(table).
			Child(hb.Script(userListRefreshScript)),
	)
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
