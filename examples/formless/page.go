package main

import (
	"github.com/dracory/hb"
	"github.com/dracory/liveflux"
)

func buildPage() hb.TagInterface {
	// Shared global filters (outside any component)
	globalFilters := hb.Div().ID("global-filters").Class("card mb-4").
		Child(hb.Div().Class("card-header").
			Child(hb.H5().Class("mb-0").Text("🔍 Shared Filters"))).
		Child(hb.Div().Class("card-body").
			Child(hb.Div().Class("row g-3").
				Child(hb.Div().Class("col-md-6").
					Child(hb.Label().Attr("for", "search").Class("form-label").Text("Search")).
					Child(hb.Input().Type("text").Name("search").ID("search").Class("form-control").Placeholder("Search..."))).
				Child(hb.Div().Class("col-md-6").
					Child(hb.Label().Attr("for", "category").Class("form-label").Text("Category")).
					Child(hb.Select().Name("category").ID("category").Class("form-select").
						Child(hb.Option().Value("").Text("All")).
						Child(hb.Option().Value("tech").Text("Tech")).
						Child(hb.Option().Value("docs").Text("Docs"))))))

	// Multi-step form sections (outside component)
	step1 := hb.Div().ID("step-1").Class("card mb-3").
		Child(hb.Div().Class("card-header").
			Child(hb.H6().Class("mb-0").Text("Step 1: Personal Info"))).
		Child(hb.Div().Class("card-body").
			Child(hb.Div().Class("row g-3").
				Child(hb.Div().Class("col-md-6").
					Child(hb.Label().Attr("for", "first_name").Class("form-label").Text("First Name")).
					Child(hb.Input().Type("text").Name("first_name").ID("first_name").Class("form-control"))).
				Child(hb.Div().Class("col-md-6").
					Child(hb.Label().Attr("for", "last_name").Class("form-label").Text("Last Name")).
					Child(hb.Input().Type("text").Name("last_name").ID("last_name").Class("form-control")))))

	step2 := hb.Div().ID("step-2").Class("card mb-3").
		Child(hb.Div().Class("card-header").
			Child(hb.H6().Class("mb-0").Text("Step 2: Contact Info"))).
		Child(hb.Div().Class("card-body").
			Child(hb.Div().Class("row g-3").
				Child(hb.Div().Class("col-md-6").
					Child(hb.Label().Attr("for", "email").Class("form-label").Text("Email")).
					Child(hb.Input().Type("email").Name("email").ID("email").Class("form-control"))).
				Child(hb.Div().Class("col-md-6").
					Child(hb.Label().Attr("for", "phone").Class("form-label").Text("Phone")).
					Child(hb.Input().Type("tel").Name("phone").ID("phone").Class("form-control")))))

	// User form for exclude example (outside component)
	userForm := hb.Div().ID("user-form").Class("card mb-3").
		Child(hb.Div().Class("card-header").
			Child(hb.H6().Class("mb-0").Text("User Profile Form"))).
		Child(hb.Div().Class("card-body").
			Child(hb.Div().Class("mb-3").
				Child(hb.Label().Attr("for", "username").Class("form-label").Text("Username")).
				Child(hb.Input().Type("text").Name("username").ID("username").Class("form-control").Value("john_doe"))).
			Child(hb.Div().Class("mb-3").
				Child(hb.Label().Attr("for", "password").Class("form-label").Text("Password (sensitive)")).
				Child(hb.Input().Type("password").Name("password").ID("password").Class("form-control sensitive").Placeholder("••••••••"))).
			Child(hb.Div().Class("mb-3").
				Child(hb.Label().Attr("for", "bio").Class("form-label").Text("Bio")).
				Child(hb.Textarea().Name("bio").ID("bio").Class("form-control").Attr("rows", "3").Text("Software developer"))))

	return hb.Webpage().
		SetTitle("Form-less Submission Example").
		SetCharset("utf-8").
		Meta(hb.Meta().Attr("name", "viewport").Attr("content", "width=device-width, initial-scale=1")).
		StyleURL("https://cdn.jsdelivr.net/npm/bootstrap@5.3.0/dist/css/bootstrap.min.css").
		StyleURL("https://cdn.jsdelivr.net/npm/bootstrap-icons@1.11.0/font/bootstrap-icons.css").
		Children([]hb.TagInterface{
			hb.Div().Class("container py-5").
				Child(hb.H1().Class("mb-4").Text("🚀 Form-less Submission Examples")).
				Child(hb.P().Class("lead mb-5").
					Text("Demonstrates data-flux-include and data-flux-exclude for flexible field collection.")).

				// Example 1: Shared filters
				Child(hb.H2().Class("mt-5 mb-3").Text("Example 1: Shared Filters")).
				Child(hb.P().Class("text-muted").
					Text("Both components below use data-flux-include=\"#global-filters\" to share the same filter inputs.")).
				Child(globalFilters).
				Child(hb.Div().Class("row g-4 mb-5").
					Child(hb.Div().Class("col-md-6").
						Child(liveflux.PlaceholderByAlias("formless.product-list", nil))).
					Child(hb.Div().Class("col-md-6").
						Child(liveflux.PlaceholderByAlias("formless.article-list", nil)))).

				// Example 2: Multi-step form
				Child(hb.H2().Class("mt-5 mb-3").Text("Example 2: Multi-Step Form")).
				Child(hb.P().Class("text-muted").
					Text("The submit button includes fields from both step sections using data-flux-include=\"#step-1, #step-2\".")).
				Child(step1).
				Child(step2).
				Child(liveflux.PlaceholderByAlias("formless.multi-step", nil)).
				Child(hb.Div().Class("mb-5")).

				// Example 3: Exclude sensitive fields
				Child(hb.H2().Class("mt-5 mb-3").Text("Example 3: Exclude Sensitive Fields")).
				Child(hb.P().Class("text-muted").
					Text("The button uses data-flux-include=\"#user-form\" and data-flux-exclude=\".sensitive\" to omit the password field.")).
				Child(userForm).
				Child(liveflux.PlaceholderByAlias("formless.exclude-example", nil)),
		}).
		Script(liveflux.JS())
}
