package main

import (
	"context"
	"net/url"
	"sort"
	"strconv"

	"github.com/dracory/liveflux"
	"github.com/gouniverse/hb"
)

// Node represents a tree node.
type Node struct {
	ID       int
	Label    string
	Children []*Node
}

// Tree is a Liveflux component demonstrating a simple mutable tree.
type Tree struct {
	liveflux.Base
	Roots  []*Node
	nextID int
}

func (t *Tree) GetAlias() string { return "tree" }

// Mount initializes state.
func (t *Tree) Mount(ctx context.Context, params map[string]string) error {
	if t.nextID == 0 {
		// start with one sample root
		root := &Node{ID: t.newID(), Label: "Root"}
		root.Children = []*Node{
			{ID: t.newID(), Label: "Child A"},
			{ID: t.newID(), Label: "Child B"},
		}
		t.Roots = []*Node{root}
	}
	return nil
}

func (t *Tree) newID() int {
	t.nextID++
	return t.nextID
}

// Handle processes actions: add_root, add_child, delete
func (t *Tree) Handle(ctx context.Context, action string, data url.Values) error {
	switch action {
	case "add_root":
		label := data.Get("label")
		if label == "" {
			label = "(unnamed)"
		}
		t.Roots = append(t.Roots, &Node{ID: t.newID(), Label: label})
		// keep deterministic order by label then id
		sort.SliceStable(t.Roots, func(i, j int) bool {
			if t.Roots[i].Label == t.Roots[j].Label { return t.Roots[i].ID < t.Roots[j].ID }
			return t.Roots[i].Label < t.Roots[j].Label
		})
	case "add_child":
		pidStr := data.Get("parent_id")
		label := data.Get("label")
		if label == "" { label = "(unnamed)" }
		pid, _ := strconv.Atoi(pidStr)
		if parent := t.findNode(pid); parent != nil {
			parent.Children = append(parent.Children, &Node{ID: t.newID(), Label: label})
			// sort children by label
			sort.SliceStable(parent.Children, func(i, j int) bool {
				if parent.Children[i].Label == parent.Children[j].Label { return parent.Children[i].ID < parent.Children[j].ID }
				return parent.Children[i].Label < parent.Children[j].Label
			})
		}
	case "delete":
		nidStr := data.Get("node_id")
		nid, _ := strconv.Atoi(nidStr)
		t.deleteNode(nid)
	}
	return nil
}

func (t *Tree) findNode(id int) *Node {
	var dfs func(list []*Node) *Node
	dfs = func(list []*Node) *Node {
		for _, n := range list {
			if n.ID == id { return n }
			if found := dfs(n.Children); found != nil { return found }
		}
		return nil
	}
	return dfs(t.Roots)
}

func (t *Tree) deleteNode(id int) {
	var del func(list []*Node) ([]*Node, bool)
	del = func(list []*Node) ([]*Node, bool) {
		changed := false
		res := list[:0]
		for _, n := range list {
			if n.ID == id { changed = true; continue }
			// recurse into children
			children, ch := del(n.Children)
			if ch { changed = true }
			n.Children = children
			res = append(res, n)
		}
		return res, changed
	}
	roots, _ := del(t.Roots)
	t.Roots = roots
}

// Render renders the tree UI with controls to add/delete nodes.
func (t *Tree) Render(ctx context.Context) hb.TagInterface {
	title := hb.H2().Text("Tree")

	// Toolbar with Add Root button that opens a modal
	toolbar := hb.Div().Class("mb-3").
		Child(
			hb.Button().Type("button").Class("btn btn-primary").
				Attr("onclick", "window.treeOpenModalRoot()").
				Child(hb.I().Class("bi bi-plus-lg me-1")).
				Child(hb.Span().Text("Add root")),
		)

	// Tree view
	treeView := hb.Div().Child(t.renderNodes(t.Roots, 0))

	// Simple modal (CSS-only) inside the component root
	modal := hb.Div().ID("tree-modal").
		Style("display:none;position:fixed;inset:0;background:rgba(0,0,0,.35);z-index:1050;align-items:center;justify-content:center;").
		Child(
			hb.Div().Style("background:#fff;min-width:320px;max-width:90vw;border-radius:.5rem;box-shadow:0 1rem 3rem rgba(0,0,0,.175);").
				Class("p-3").
				Child(hb.Div().Class("d-flex justify-content-between align-items-center mb-2").
					Child(hb.Div().Class("d-flex align-items-center gap-2").
						Child(hb.I().Class("bi bi-node-plus"))).
					Child(hb.Strong().Text("Add node")).
					Child(hb.Button().Type("button").Class("btn-close").Attr("aria-label", "Close").Attr("onclick", "window.treeCloseModal()")),
				).
				Child(
					hb.Form().ID("tree-modal-form").Class("d-flex flex-column gap-2").
						Child(hb.Input().Type("hidden").Name("parent_id").ID("tree-modal-parent")).
						Child(hb.Label().Attr("for", "tree-modal-label").Text("Title")).
						Child(hb.Input().Type("text").Name("label").ID("tree-modal-label").Class("form-control").Placeholder("Enter title")).
						Child(
							hb.Div().Class("d-flex justify-content-end gap-2 mt-2").
								Child(hb.Button().Type("button").Class("btn btn-light").Attr("onclick", "window.treeCloseModal()").
									Child(hb.Span().Text("Cancel"))).
								Child(hb.Button().Type("submit").ID("tree-modal-submit").Class("btn btn-primary").Data("flux-action", "add_root").
									Child(hb.I().Class("bi bi-check-lg me-1")).
									Child(hb.Span().Text("Save"))),
						),
				),
		)

	// Inline script to manage modal behavior
	script := hb.Script(`
      (function(){
        var modal = null, form = null, parentInput = null, submitBtn = null;
        function ensure(){
          modal = modal || document.getElementById('tree-modal');
          form = form || document.getElementById('tree-modal-form');
          parentInput = parentInput || document.getElementById('tree-modal-parent');
          submitBtn = submitBtn || document.getElementById('tree-modal-submit');
        }
        window.treeOpenModalRoot = function(){
          ensure();
          if(form) form.reset();
          if(parentInput) parentInput.value = '';
          if(submitBtn) submitBtn.setAttribute('data-flux-action','add_root');
          if(modal) modal.style.display = 'flex';
        };
        window.treeOpenModalChild = function(pid){
          ensure();
          if(form) form.reset();
          if(parentInput) parentInput.value = String(pid||'');
          if(submitBtn) submitBtn.setAttribute('data-flux-action','add_child');
          if(modal) modal.style.display = 'flex';
        };
        window.treeCloseModal = function(){ ensure(); if(modal) modal.style.display = 'none'; };
      })();
    `)

	content := hb.Div().
		Child(title).
		Child(toolbar).
		Child(treeView).
		Child(modal).
		Child(script)

	return t.Root(content)
}

func (t *Tree) renderNodes(nodes []*Node, depth int) hb.TagInterface {
	// base list class; indent nested lists
	ulClass := "list-unstyled"
	if depth > 0 {
		ulClass += " ms-4"
	}
	ul := hb.Ul().Class(ulClass)
	for _, n := range nodes {
		iconClass := "bi bi-file-earmark"
		if len(n.Children) > 0 {
			iconClass = "bi bi-folder"
		}
		ul.Child(
			hb.Li().Class("mb-2").
				Child(hb.Div().Class("d-flex gap-2 align-items-center mb-1").
					Child(hb.I().Class(iconClass + " me-1")).
					Child(hb.Span().Text(n.Label + " (#" + strconv.Itoa(n.ID) + ")")).
					// Delete remains a simple form submit
					Child(hb.Form().Class("d-inline-flex gap-2").
						Child(hb.Input().Type("hidden").Name("node_id").Value(strconv.Itoa(n.ID))).
						Child(hb.Button().Type("submit").Class("btn btn-sm btn-outline-danger").Data("flux-action", "delete").
							Child(hb.I().Class("bi bi-trash me-1")).
							Child(hb.Span().Text("Delete"))),
					).
					// Add child now opens a modal to enter the title
					Child(hb.Button().Type("button").Class("btn btn-sm btn-outline-primary").
						Attr("onclick", "window.treeOpenModalChild("+strconv.Itoa(n.ID)+")").
						Child(hb.I().Class("bi bi-plus-lg me-1")).
						Child(hb.Span().Text("Add child"))),
				).
				Child(t.renderNodes(n.Children, depth+1)),
		)
	}
	return ul
}

func init() {
	liveflux.Register(new(Tree))
}
