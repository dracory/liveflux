package main

import (
	"context"
	"log"
	"net/url"

	"github.com/dracory/hb"
	"github.com/dracory/liveflux"
)

// NotificationBanner shows notifications when events are dispatched.
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

	script := hb.Script(`
        (function(){
            var root = liveflux.findComponent('` + nb.GetKind() + `', '` + nb.GetID() + `');
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
