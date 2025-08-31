package main

import (
	"context"
	"net/url"
	"sort"
	"strconv"
	"strings"

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
	Title  string
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
			if t.Roots[i].Label == t.Roots[j].Label {
				return t.Roots[i].ID < t.Roots[j].ID
			}
			return t.Roots[i].Label < t.Roots[j].Label
		})
	case "add_child":
		pidStr := data.Get("parent_id")
		label := data.Get("label")
		if label == "" {
			label = "(unnamed)"
		}
		pid, _ := strconv.Atoi(pidStr)
		if parent := t.findNode(pid); parent != nil {
			parent.Children = append(parent.Children, &Node{ID: t.newID(), Label: label})
			// sort children by label
			sort.SliceStable(parent.Children, func(i, j int) bool {
				if parent.Children[i].Label == parent.Children[j].Label {
					return parent.Children[i].ID < parent.Children[j].ID
				}
				return parent.Children[i].Label < parent.Children[j].Label
			})
		}
	case "edit":
		nidStr := data.Get("node_id")
		label := data.Get("label")
		if label == "" {
			label = "(unnamed)"
		}
		nid, _ := strconv.Atoi(nidStr)
		if node := t.findNode(nid); node != nil {
			node.Label = label
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
			if n.ID == id {
				return n
			}
			if found := dfs(n.Children); found != nil {
				return found
			}
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
			if n.ID == id {
				changed = true
				continue
			}
			// recurse into children
			children, ch := del(n.Children)
			if ch {
				changed = true
			}
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
	title := hb.H2().Class("mt-2 mb-3").Text(t.Title)

	// Toolbar with Add Root button that opens a modal
	toolbar := hb.Div().Class("mb-4").
		Child(
			hb.Button().Type("button").Class("btn btn-primary btn-lg rounded-pill shadow-sm").
				Attr("onclick", "window.treeOpenModalRoot()").
				Child(hb.I().Class("bi bi-plus-lg me-2")).
				Child(hb.Span().Text("Add root")),
		)

	// Tree view
	treeView := hb.Div().Child(t.renderNodes(t.Roots, 0))

	// Simple modal (CSS-only) inside the component root
	modalHeader := hb.Div().Class("d-flex justify-content-between align-items-center px-3 py-2 bg-light border-bottom").
		Child(hb.Div().Class("d-flex align-items-center gap-2 fw-semibold").
			Child(hb.I().Class("bi bi-pencil-square text-primary")),
		).
		Child(hb.Strong().ID("tree-modal-title").Text("Add node")).
		Child(hb.Button().Type("button").Class("btn-close").Attr("aria-label", "Close").Attr("onclick", "window.treeCloseModal()"))

	modalButtons := hb.Div().Class("d-flex justify-content-end gap-2 mt-3").
		Child(
			hb.Button().Type("button").Class("btn btn-light").Attr("onclick", "window.treeCloseModal()").
				Child(hb.Span().Text("Cancel")),
		).
		Child(
			hb.Button().Type("button").ID("tree-modal-submit").Class("btn btn-primary rounded-pill px-4").Data("flux-action", "add_root").
				Child(hb.I().Class("bi bi-check-lg me-1")).
				Child(hb.Span().Text("Save")),
		)

	modalForm := hb.Div().ID("tree-modal-form").Class("d-flex flex-column gap-2 p-3").
		Child(hb.Input().Type("hidden").Name("node_id").ID("tree-modal-node")).
		Child(hb.Input().Type("hidden").Name("parent_id").ID("tree-modal-parent")).
		Child(hb.Label().Attr("for", "tree-modal-label").Class("form-label mb-1").Text("Title")).
		Child(hb.Input().Type("text").Name("label").ID("tree-modal-label").Class("form-control form-control-lg").Placeholder("Enter title")).
		Child(modalButtons)

	modalCard := hb.Div().
		Style("background:#fff;min-width:360px;max-width:92vw;border-radius:.75rem;box-shadow:0 1rem 3rem rgba(0,0,0,.2);").
		Class("overflow-hidden").
		Child(modalHeader).
		Child(modalForm)

	modal := hb.Div().ID("tree-modal").
		Style("display:none;position:fixed;inset:0;background:rgba(0,0,0,.35);z-index:1050;align-items:center;justify-content:center;").
		Child(modalCard)

	// Inline script to manage modal behavior (no inner form; outer component form will be serialized)
	script := hb.Script(`
      (function(){
        var modal = null, parentInput = null, nodeInput = null, submitBtn = null, titleEl = null, labelInput = null;
        function ensure(){
          modal = modal || document.getElementById('tree-modal');
          parentInput = parentInput || document.getElementById('tree-modal-parent');
          nodeInput = nodeInput || document.getElementById('tree-modal-node');
          submitBtn = submitBtn || document.getElementById('tree-modal-submit');
          titleEl = titleEl || document.getElementById('tree-modal-title');
          labelInput = labelInput || document.getElementById('tree-modal-label');
        }
        function resetFields(){ if(parentInput) parentInput.value=''; if(nodeInput) nodeInput.value=''; if(labelInput) labelInput.value=''; }
        window.treeOpenModalRoot = function(){
          ensure();
          resetFields();
          if(titleEl) titleEl.textContent = 'Add root';
          if(submitBtn) submitBtn.setAttribute('data-flux-action','add_root');
          if(modal) modal.style.display = 'flex';
        };
        window.treeOpenModalChild = function(pid){
          ensure();
          resetFields();
          if(parentInput) parentInput.value = String(pid||'');
          if(titleEl) titleEl.textContent = 'Add child';
          if(submitBtn) submitBtn.setAttribute('data-flux-action','add_child');
          if(modal) modal.style.display = 'flex';
        };
        window.treeOpenModalEdit = function(id,label){
          ensure();
          resetFields();
          if(nodeInput) nodeInput.value = String(id||'');
          if(labelInput) labelInput.value = (label||'');
          if(titleEl) titleEl.textContent = 'Edit node';
          if(submitBtn) submitBtn.setAttribute('data-flux-action','edit');
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
		iconClass := "bi bi-file-earmark text-secondary"
		if len(n.Children) > 0 {
			iconClass = "bi bi-folder text-warning"
		}
		ul.Child(
			hb.Li().
				Child(
					hb.Div().Class("d-flex align-items-center justify-content-between border rounded-3 px-3 py-2 mb-2 bg-white").
						Child(
							hb.Div().Class("d-flex align-items-center gap-2").
								Child(hb.I().Class(iconClass + " fs-5")).
								Child(hb.Span().Class("fw-semibold").Text(n.Label)).
								Child(hb.Span().Class("text-muted small").Text(" (#" + strconv.Itoa(n.ID) + ")")),
						).
						Child(
							hb.Div().Class("btn-group btn-group-sm d-inline-flex align-items-stretch").
								// Edit
								Child(hb.Button().Type("button").Class("btn btn-outline-secondary").
									Attr("onclick", "window.treeOpenModalEdit("+strconv.Itoa(n.ID)+", '"+escapeJS(n.Label)+"')").
									Child(hb.I().Class("bi bi-pencil")).
									Child(hb.Span().Class("ms-1 d-none d-sm-inline").Text("Edit")),
								).
								// Add child
								Child(hb.Button().Type("button").Class("btn btn-outline-primary").
									Attr("onclick", "window.treeOpenModalChild("+strconv.Itoa(n.ID)+")").
									Child(hb.I().Class("bi bi-plus-lg")).
									Child(hb.Span().Class("ms-1 d-none d-sm-inline").Text("Add child")),
								).
								// Delete action via data attributes
								Child(hb.Button().Type("button").Class("btn btn-outline-danger").
									Data("flux-action", "delete").
									Data("flux-param-node_id", strconv.Itoa(n.ID)).
									Child(hb.I().Class("bi bi-trash")).
									Child(hb.Span().Class("ms-1 d-none d-sm-inline").Text("Delete")),
								),
						),
				).
				Child(t.renderNodes(n.Children, depth+1)),
		)
	}
	return ul
}

func escapeJS(s string) string {
	// Escape backslashes and single quotes for safe embedding inside single-quoted JS strings
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "'", "\\'")
	return s
}

func init() {
	liveflux.Register(new(Tree))
}
