package websocket

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {
	// Registered clients.
	clients map[*Client]bool

	// Inbound messages from the clients.
	broadcast chan []byte

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client

	//
	namespace string

	//
	room string
}

func newHub() *Hub {
	return &Hub{
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
	}
}

var Instance *Hub = newHub()

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)

				// Remove empty hubs from namespace where applicable
				if h.namespace != nil&len(h.clients) == 0 {
					removeHub(h.namespace, h.room)
				}
			}
		case message := <-h.broadcast:
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}

// Structure to store namespace
var hubs = map[string]map[string]*Hub{}

// Returns hub from namespace. Client should be registered to hub immediately.
func GetHub(namespace string, room string) *Hub {
	if _, ok := hubs[namespace]; !ok {
		hubs[namespace] = map[string]*Hub{}
	}
	namespaceHubs := hubs[namespace]

	if _, ok := namespaceHubs[room]; !ok {
		hub = &NewHub()
		hub.namespace = namespace
		hub.room = room

		namespaceHubs[room] = &hub
	}
	return namespaceHubs[room]
}

// Unexported method to remove hub from namespace table
func removeHub(namespace string, room string) {
	delete(hubs[namespace], room)

	if len(hubs[namespace]) == 0 {
		delete(hubs[namespace])
	}
}
