package main

import (
    "context"
    "log"
    "net/url"

    "github.com/dracory/hb"
    "github.com/dracory/liveflux"
)

// Post represents a single post entry.
type Post struct {
    Title     string
    Timestamp string
}

// PostList is a component that listens for post-created events.
type PostList struct {
    liveflux.Base
    Posts []Post
}

func (pl *PostList) Mount(ctx context.Context, params map[string]string) error {
    pl.Posts = []Post{}
    liveflux.RegisterEventListeners(pl, pl.GetEventDispatcher())
    return nil
}

func (pl *PostList) Handle(ctx context.Context, action string, data url.Values) error {
    log.Printf("[PostList] Handle action: %s", action)
    switch action {
    case "clear":
        pl.Posts = []Post{}
    case "add-post":
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

// OnPostCreated listens for the "post-created" event.
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
