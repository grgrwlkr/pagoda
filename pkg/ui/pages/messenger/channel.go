package messenger

import (
	"fmt"

	"github.com/labstack/echo/v4"
	"github.com/mikestefanello/pagoda/pkg/ui"
	messengerComponents "github.com/mikestefanello/pagoda/pkg/ui/components/messenger"
	messengerLayouts "github.com/mikestefanello/pagoda/pkg/ui/layouts"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// Channel renders the channel page with messages
func Channel(ctx echo.Context, channelID int64, channelName string, messages []messengerComponents.MessageData) error {
	r := ui.NewRequest(ctx)
	r.Title = "#" + channelName

	content := Div(
		Class("flex flex-col h-full"),
		// Channel header
		channelHeader(r, channelID, channelName),
		// Messages list
		messengerComponents.MessageList(r, messages),
		// Typing indicator (will be updated via WebSocket)
		Div(
			ID("typing-indicator-container"),
			// TypingIndicator will be added via WebSocket updates
		),
		// Message input
		messengerComponents.MessageInput(r, channelID, false), // false = not a DM
		// WebSocket connection script
		websocketScript(r, channelID),
	)

	return r.Render(messengerLayouts.Messenger, content)
}

func websocketScript(r *ui.Request, channelID int64) Node {
	return Script(
		Raw(`
(function() {
	const ws = new WebSocket('ws://' + window.location.host + '/ws');
	const channelID = ` + fmt.Sprintf("%d", channelID) + `;
	
	ws.onopen = function() {
		console.log('WebSocket connected');
		// Join channel
		ws.send(JSON.stringify({
			type: 'join_channel',
			data: { channel_id: channelID }
		}));
	};
	
	ws.onmessage = function(event) {
		const message = JSON.parse(event.data);
		handleWebSocketMessage(message);
	};
	
	ws.onerror = function(error) {
		console.error('WebSocket error:', error);
	};
	
	ws.onclose = function() {
		console.log('WebSocket disconnected');
	};
	
	function handleWebSocketMessage(message) {
		switch(message.type) {
			case 'message_new':
				addMessage(message.data);
				break;
			case 'message_edited':
				updateMessage(message.data);
				break;
			case 'message_deleted':
				removeMessage(message.data.message_id);
				break;
			case 'user_typing':
				showTypingIndicator(message.data.user_id);
				break;
			case 'reaction_added':
				updateReaction(message.data);
				break;
			case 'reaction_removed':
				updateReaction(message.data);
				break;
		}
	}
	
	function addMessage(data) {
		// Add message to list
		const messageList = document.getElementById('message-list');
		if (messageList) {
			// Create message element and append
			// This is a simplified version - full implementation would create proper HTML
		}
	}
	
	function updateMessage(data) {
		const messageEl = document.getElementById('message-' + data.message_id);
		if (messageEl) {
			// Update message content
		}
	}
	
	function removeMessage(messageID) {
		const messageEl = document.getElementById('message-' + messageID);
		if (messageEl) {
			messageEl.remove();
		}
	}
	
	function showTypingIndicator(userID) {
		// Show typing indicator
	}
	
	function updateReaction(data) {
		// Update reaction on message
	}
	
	// Handle message input
	const messageInput = document.getElementById('message-input');
	const sendButton = document.getElementById('message-send');
	
	if (messageInput && sendButton) {
		sendButton.addEventListener('click', function() {
			const content = messageInput.value.trim();
			if (content) {
				ws.send(JSON.stringify({
					type: 'message_send',
					data: {
						channel_id: channelID,
						content: content
					}
				}));
				messageInput.value = '';
			}
		});
		
		messageInput.addEventListener('keypress', function(e) {
			if (e.key === 'Enter' && !e.shiftKey) {
				e.preventDefault();
				sendButton.click();
			}
		});
	}
})();
		`),
	)
}

func channelHeader(r *ui.Request, channelID int64, channelName string) Node {
	return Div(
		Class("border-b border-base-300 p-4 bg-base-100 flex items-center justify-between"),
		Div(
			Class("flex items-center gap-3"),
			H2(
				Class("text-xl font-semibold"),
				Span(Class("text-base-content/60"), Text("#")),
				Text(channelName),
			),
		),
		Div(
			Class("flex items-center gap-2"),
			// Channel info button
			Button(
				Class("btn btn-sm btn-ghost"),
				Text("ℹ️"),
				Attr("title", "Channel info"),
				// TODO: Open right panel with channel info
			),
		),
	)
}
