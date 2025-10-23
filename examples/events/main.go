package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/dracory/hb"
	"github.com/dracory/liveflux"
)

// PostCreator is a component that creates posts and dispatches events
type PostCreator struct {
	liveflux.Base
	Title string
}

func (pc *PostCreator) Mount(ctx context.Context, params map[string]string) error {
	pc.Title = ""
	return nil
}

func (pc *PostCreator) Handle(ctx context.Context, action string, data url.Values) error {
	if action == "create" {
		pc.Title = data.Get("title")
		log.Printf("[PostCreator] Create action with title: %s", pc.Title)
		if pc.Title != "" {
			// Dispatch event to notify other components
			log.Printf("[PostCreator] Dispatching post-created event")
			pc.Dispatch("post-created", map[string]interface{}{
				"title":     pc.Title,
				"timestamp": time.Now().Format("15:04:05"),
			})
			pc.Title = "" // Clear input after creating
		}
	}
	return nil
}

func (pc *PostCreator) Render(ctx context.Context) hb.TagInterface {
	return pc.Root(
		hb.Div().Class("card").Children([]hb.TagInterface{
			hb.H2().Text("Create Post"),
			hb.Form().Children([]hb.TagInterface{
				hb.Input().
					Type("text").
					Name("title").
					Value(pc.Title).
					Placeholder("Enter post title...").
					Class("input"),
				hb.Button().
					Type("submit").
					Attr("data-flux-action", "create").
					Class("btn btn-primary").
					Text("Create Post"),
			}),
		}),
	)
}

// PostList is a component that listens for post-created events
type PostList struct {
	liveflux.Base
	Posts []Post
}

type Post struct {
	Title     string
	Timestamp string
}

func (pl *PostList) Mount(ctx context.Context, params map[string]string) error {
	pl.Posts = []Post{}
	// Register event listener using method naming convention
	liveflux.RegisterEventListeners(pl, pl.GetEventDispatcher())
	return nil
}

func (pl *PostList) Handle(ctx context.Context, action string, data url.Values) error {
	log.Printf("[PostList] Handle action: %s", action)
	if action == "clear" {
		pl.Posts = []Post{}
	} else if action == "add-post" {
		// Add post from event data
		title := data.Get("title")
		timestamp := data.Get("timestamp")
		log.Printf("[PostList] Adding post from event: title=%s, timestamp=%s", title, timestamp)
		if title != "" {
			pl.Posts = append(pl.Posts, Post{
				Title:     title,
				Timestamp: timestamp,
			})
			log.Printf("[PostList] Post added. Total posts: %d", len(pl.Posts))
		}
	}
	return nil
}

// OnPostCreated is automatically registered as a listener for "post-created" event
func (pl *PostList) OnPostCreated(ctx context.Context, event liveflux.Event) error {
	title, _ := event.Data["title"].(string)
	timestamp, _ := event.Data["timestamp"].(string)

	pl.Posts = append(pl.Posts, Post{
		Title:     title,
		Timestamp: timestamp,
	})
	return nil
}

func (pl *PostList) Render(ctx context.Context) hb.TagInterface {
	postItems := []hb.TagInterface{}
	for _, post := range pl.Posts {
		postItems = append(postItems, hb.Li().Class("post-item").Children([]hb.TagInterface{
			hb.Strong().Text(post.Title),
			hb.Span().Text(" - " + post.Timestamp).Class("timestamp"),
		}))
	}

	if len(postItems) == 0 {
		postItems = append(postItems, hb.Li().Class("empty").Text("No posts yet..."))
	}

	// Add JavaScript to listen for post-created events and call the server action via $wire
	script := hb.Script(`
		(function(){
			var root = document.currentScript.closest('[data-flux-root]');
			if(!root) {
				console.error('[PostList] Could not find component root');
				return;
			}

			console.log('[PostList] Initializing post-created listener');

			function setupListener(){
				if(!root.$wire){
					console.log('[PostList] $wire not ready yet, retrying...');
					setTimeout(setupListener, 50);
					return;
				}

				console.log('[PostList] $wire ready, registering post-created listener');
				root.$wire.on('post-created', function(event){
					var data = event && event.data ? event.data : {};
					var title = data.title || '';
					var timestamp = data.timestamp || '';
					console.log('[PostList] Event received, calling add-post with', title, timestamp);
					root.$wire.call('add-post', {
						title: title,
						timestamp: timestamp,
					});
				});
			}

			setupListener();
		})();
	`)

	return pl.Root(
		hb.Div().Class("card").Children([]hb.TagInterface{
			hb.H2().Text("Post List"),
			hb.Ul().Class("post-list").Children(postItems),
			hb.Button().
				Attr("data-flux-action", "clear").
				Class("btn btn-secondary").
				Text("Clear All"),
			script,
		}),
	)
}

// NotificationBanner shows notifications when events are dispatched
type NotificationBanner struct {
	liveflux.Base
	Message   string
	Timestamp string
}

func (nb *NotificationBanner) Mount(ctx context.Context, params map[string]string) error {
	nb.Message = ""
	return nil
}

func (nb *NotificationBanner) Handle(ctx context.Context, action string, data url.Values) error {
	log.Printf("[NotificationBanner] Handle action: %s", action)
	switch action {
	case "show":
		nb.Message = data.Get("title")
		nb.Timestamp = data.Get("timestamp")
	case "clear":
		nb.Message = ""
		nb.Timestamp = ""
	}
	return nil
}

func (nb *NotificationBanner) Render(ctx context.Context) hb.TagInterface {
	content := hb.Div().Class("notification-content")

	if nb.Message != "" {
		content.Children([]hb.TagInterface{
			hb.Div().Class("notification active").Children([]hb.TagInterface{
				hb.Span().Text(nb.Message),
				hb.Span().Text(" - " + nb.Timestamp).Class("timestamp"),
			}),
		})
	} else {
		content.Child(hb.Div().Class("notification").Text("Waiting for events..."))
	}

	// Add JavaScript to listen for events
	script := hb.Script(`
		(function(){
			var root = document.currentScript.closest('[data-flux-root]');
			if(!root) {
				console.error('[NotificationBanner] Could not find component root');
				return;
			}

			function setupListener(){
				if(!root.$wire){
					setTimeout(setupListener, 50);
					return;
				}

				root.$wire.on('post-created', function(event){
					var data = event && event.data ? event.data : {};
					console.log('[NotificationBanner] Event received, calling show with', data);
					root.$wire.call('show', {
						title: data.title || '',
						timestamp: data.timestamp || '',
					});
				});
			}

			setupListener();
		})();
	`)

	return nb.Root(
		hb.Div().Class("card notification-card").Children([]hb.TagInterface{
			hb.H2().Text("Notifications"),
			content,
			script,
		}),
	)
}

func main() {
	// Register components
	if err := liveflux.RegisterByAlias("post-creator", &PostCreator{}); err != nil {
		log.Fatal(err)
	}
	if err := liveflux.RegisterByAlias("post-list", &PostList{}); err != nil {
		log.Fatal(err)
	}
	if err := liveflux.RegisterByAlias("notification-banner", &NotificationBanner{}); err != nil {
		log.Fatal(err)
	}

	// Setup HTTP server
	mux := http.NewServeMux()

	// Liveflux handler
	mux.Handle("/liveflux", liveflux.NewHandler(nil))

	// Main page
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		page := hb.Webpage().
			SetTitle("Liveflux Events Example").
			SetCharset("utf-8").
			Style(`
				body { font-family: Arial, sans-serif; max-width: 1200px; margin: 0 auto; padding: 20px; background: #f5f5f5; }
				.container { display: grid; grid-template-columns: 1fr 1fr; gap: 20px; margin-bottom: 20px; }
					.card { background: white; padding: 20px; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
					.notification-card { grid-column: 1 / -1; }
					h2 { margin-top: 0; color: #333; }
					.input { width: 100%; padding: 10px; margin: 10px 0; border: 1px solid #ddd; border-radius: 4px; box-sizing: border-box; }
					.btn { padding: 10px 20px; border: none; border-radius: 4px; cursor: pointer; font-size: 14px; }
					.btn-primary { background: #007bff; color: white; }
					.btn-primary:hover { background: #0056b3; }
					.btn-secondary { background: #6c757d; color: white; margin-top: 10px; }
					.btn-secondary:hover { background: #545b62; }
					.post-list { list-style: none; padding: 0; margin: 15px 0; }
					.post-item { padding: 10px; margin: 5px 0; background: #f8f9fa; border-radius: 4px; }
					.timestamp { color: #666; font-size: 0.9em; }
					.empty { color: #999; font-style: italic; }
					.notification { padding: 15px; background: #e9ecef; border-radius: 4px; text-align: center; }
					.notification.active { background: #d4edda; color: #155724; border: 1px solid #c3e6cb; }
					form { margin: 0; }
				`).
			Children([]hb.TagInterface{
				hb.H1().Text("Liveflux Events Example"),
				hb.P().Text("This example demonstrates the event system. Create a post in the left component, and watch it appear in the list on the right via events."),
				hb.Div().Class("container").Children([]hb.TagInterface{
					liveflux.PlaceholderByAlias("post-creator"),
					liveflux.PlaceholderByAlias("post-list"),
				}),
				liveflux.PlaceholderByAlias("notification-banner"),
			}).
			Script(liveflux.JS())

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, page.ToHTML())
	})

	addr := ":8084"
	fmt.Printf("Server running at http://localhost%s\n", addr)
	fmt.Println("Open your browser and create posts to see events in action!")
	log.Fatal(http.ListenAndServe(addr, mux))
}
