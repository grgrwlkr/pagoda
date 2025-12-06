package messenger

import (
	"encoding/json"
	"fmt"

	"github.com/mikestefanello/pagoda/pkg/ui"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// WebSocketConnection creates an HTMX WebSocket connection component with Alpine.js for message handling.
// This uses Alpine.js for WebSocket message processing (WebSocket requires JS).
// Parameters:
//   - r: UI request with context
//   - wsPath: WebSocket path (e.g., "/ws")
//   - channelID: Channel ID for joining (0 if not a channel)
//   - dmID: Direct message ID for joining (0 if not a DM)
//
// Returns:
//   - HTML node with HTMX WebSocket connection and minimal JavaScript message handler
func WebSocketConnection(r *ui.Request, wsPath string, channelID, dmID int64) Node {
	// Prepare join message data
	var joinMessage map[string]interface{}
	if channelID > 0 {
		joinMessage = map[string]interface{}{
			"type": "join_channel",
			"data": map[string]interface{}{
				"channel_id": channelID,
			},
		}
	} else if dmID > 0 {
		joinMessage = map[string]interface{}{
			"type": "join_dm",
			"data": map[string]interface{}{
				"dm_id": dmID,
			},
		}
	}

	var joinMessageJSON []byte
	if len(joinMessage) > 0 {
		joinMessageJSON, _ = json.Marshal(joinMessage)
	}

	// Create Alpine.js component for WebSocket message handling
	// This uses Alpine.js for WebSocket message processing (WebSocket requires JS)
	alpineData := fmt.Sprintf(`{
		channelID: %d,
		dmID: %d,
		handleMessage(message) {
			// Handle WebSocket messages via Alpine.js
			const msg = typeof message === 'string' ? JSON.parse(message) : message;
			switch(msg.type) {
				case 'message_new':
					this.handleMessageNew(msg.data);
					break;
				case 'message_edited':
					this.handleMessageEdited(msg.data);
					break;
				case 'message_deleted':
					this.handleMessageDeleted(msg.data);
					break;
				case 'user_typing':
					this.handleUserTyping(msg.data);
					break;
				case 'user_online':
				case 'user_offline':
					this.handleUserStatus(msg.data, msg.type === 'user_online');
					break;
				case 'reaction_added':
				case 'reaction_removed':
				case 'channel_updated':
				case 'member_joined':
				case 'member_left':
					// Reload page for these events
					window.location.reload();
					break;
			}
		},
		handleMessageNew(data) {
			// Check if thread reply and thread panel is open
			if (data.thread_id) {
				const rightPanel = document.getElementById('right-panel');
				if (rightPanel && !rightPanel.classList.contains('hidden')) {
					const panelThreadId = rightPanel.getAttribute('data-thread-id');
					if (panelThreadId && parseInt(panelThreadId) === data.thread_id) {
						// Reload thread panel via HTMX - update button URL and trigger click
						const reloadBtn = document.getElementById('ws-reload-thread-button');
						if (reloadBtn) {
							reloadBtn.setAttribute('hx-get', '/message/' + panelThreadId + '/thread/panel');
							reloadBtn.click(); // Trigger HTMX request via button click
						}
						return;
					}
				}
			}
			// Regular message - scroll to bottom
			this.scrollToBottom();
			this.clearTypingIndicator();
		},
		handleMessageEdited(data) {
			// For message edits, we could reload the message list or specific message
			// For now, just log - full reload can be done via HTMX if needed
			console.log('Message edited:', data.message_id);
		},
		handleMessageDeleted(data) {
			// Remove message element
			const el = document.getElementById('message-' + data.message_id);
			if (el) {
				el.remove();
			}
		},
		handleUserTyping(data) {
			// Trigger custom event for typing indicator
			const container = document.getElementById('typing-indicator-container');
			if (container) {
				container.dispatchEvent(new CustomEvent('show-typing', {
					detail: { userID: data.user_id, userName: data.user_name || 'Someone' }
				}));
			}
		},
		handleUserStatus(data, isOnline) {
			// Update user status indicator
			console.log('User status:', data.user_id, isOnline);
		},
		clearTypingIndicator() {
			// Trigger custom event for typing indicator
			const container = document.getElementById('typing-indicator-container');
			if (container) {
				container.dispatchEvent(new CustomEvent('clear-typing'));
			}
		},
		scrollToBottom() {
			const messageList = document.getElementById('message-list');
			if (messageList) {
				messageList.scrollTop = messageList.scrollHeight;
			}
		}
	}`, channelID, dmID)

	// Create WebSocket connection container with HTMX WebSocket extension
	return Div(
		ID("websocket-connection"),
		Attr("hx-ext", "ws"),       // Enable HTMX WebSocket extension
		Attr("ws-connect", wsPath), // Connect to WebSocket endpoint
		Attr("x-data", alpineData), // Alpine.js component for message handling
		// Handle WebSocket messages via Alpine.js
		Attr("hx-on::htmx:ws-message", "handleMessage(event.detail.message)"),
			// Hidden form for sending join message via HTMX WebSocket
			// This replaces htmx.trigger() call
			If(len(joinMessageJSON) > 0,
				Form(
					ID("ws-join-form"),
					Attr("ws-send"), // HTMX will send form data over WebSocket
					Style("display: none;"),
					Input(
						Type("hidden"),
						Name("message"),
						Value(string(joinMessageJSON)),
					),
				),
			),
			// Trigger form submission on WebSocket connect
			Attr("hx-on::htmx:ws-connect", `
			// Send join message via HTMX form (replaces htmx.trigger)
			const joinForm = document.getElementById('ws-join-form');
			if (joinForm) {
				joinForm.requestSubmit();
			}
		`),
			// Hidden container for WebSocket connection
			Style("display: none;"),
		),
	}
}
