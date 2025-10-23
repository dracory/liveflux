package main

import (
    "context"
    "log"
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
