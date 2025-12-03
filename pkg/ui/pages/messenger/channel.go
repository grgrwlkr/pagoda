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
	let ws = null;
	let reconnectAttempts = 0;
	const maxReconnectAttempts = 5;
	const reconnectDelay = 3000; // 3 seconds
	const channelID = ` + fmt.Sprintf("%d", channelID) + `;
	
	// Track typing users
	const typingUsers = new Map(); // userID -> timeout
	
	function connect() {
		const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
		ws = new WebSocket(protocol + '//' + window.location.host + '/ws');
		
		ws.onopen = function() {
			console.log('WebSocket connected');
			reconnectAttempts = 0;
			// Join channel
			ws.send(JSON.stringify({
				type: 'join_channel',
				data: { channel_id: channelID }
			}));
		};
		
		ws.onmessage = function(event) {
			try {
				const message = JSON.parse(event.data);
				handleWebSocketMessage(message);
			} catch (e) {
				console.error('Failed to parse WebSocket message:', e);
			}
		};
		
		ws.onerror = function(error) {
			console.error('WebSocket error:', error);
		};
		
		ws.onclose = function(event) {
			console.log('WebSocket disconnected', 'code', event.code, 'reason', event.reason);
			// Clear typing indicators
			clearTypingIndicator();
			
			// Attempt to reconnect if not a normal closure
			if (event.code !== 1000 && reconnectAttempts < maxReconnectAttempts) {
				reconnectAttempts++;
				console.log('Attempting to reconnect...', 'attempt', reconnectAttempts);
				setTimeout(connect, reconnectDelay);
			} else if (reconnectAttempts >= maxReconnectAttempts) {
				console.error('Max reconnection attempts reached');
			}
		};
	}
	
	function handleWebSocketMessage(message) {
		switch(message.type) {
			case 'message_new':
				// Message is already added via HTMX, just scroll to bottom
				scrollToBottom();
				clearTypingIndicator();
				break;
			case 'message_edited':
				updateMessage(message.data);
				break;
			case 'message_deleted':
				removeMessage(message.data.message_id);
				break;
			case 'user_typing':
				showTypingIndicator(message.data.user_id, message.data.user_name || 'Someone');
				break;
			case 'user_online':
				updateUserStatus(message.data.user_id, true);
				break;
			case 'user_offline':
				updateUserStatus(message.data.user_id, false);
				break;
			case 'reaction_added':
			case 'reaction_removed':
				// Reload the message to update reactions
				location.reload();
				break;
			case 'channel_updated':
			case 'member_joined':
			case 'member_left':
				// Reload page to update sidebar
				location.reload();
				break;
			case 'error':
				console.error('WebSocket error:', message.data);
				break;
		}
	}
	
	function updateMessage(data) {
		const messageEl = document.getElementById('message-' + data.message_id);
		if (messageEl) {
			const contentEl = messageEl.querySelector('.message-content');
			if (contentEl) {
				contentEl.textContent = data.content;
			}
		}
	}
	
	function removeMessage(messageID) {
		const messageEl = document.getElementById('message-' + messageID);
		if (messageEl) {
			messageEl.remove();
		}
	}
	
	function showTypingIndicator(userID, userName) {
		// Clear existing timeout for this user
		if (typingUsers.has(userID)) {
			clearTimeout(typingUsers.get(userID));
		}
		
		// Update typing indicator
		const container = document.getElementById('typing-indicator-container');
		if (container) {
			container.innerHTML = '<div id="typing-indicator" class="px-4 py-2 text-sm text-base-content/60 italic">' + 
				userName + ' is typing<span class="inline-block ml-1">...</span></div>';
		}
		
		// Set timeout to hide typing indicator after 3 seconds
		const timeout = setTimeout(() => {
			typingUsers.delete(userID);
			updateTypingIndicator();
		}, 3000);
		
		typingUsers.set(userID, timeout);
	}
	
	function clearTypingIndicator() {
		typingUsers.clear();
		const container = document.getElementById('typing-indicator-container');
		if (container) {
			container.innerHTML = '';
		}
	}
	
	function updateTypingIndicator() {
		const container = document.getElementById('typing-indicator-container');
		if (!container) return;
		
		if (typingUsers.size === 0) {
			container.innerHTML = '';
			return;
		}
		
		// For now, just show "Someone is typing..."
		// In full implementation, we'd track user names
		container.innerHTML = '<div id="typing-indicator" class="px-4 py-2 text-sm text-base-content/60 italic">' + 
			'Someone is typing<span class="inline-block ml-1">...</span></div>';
	}
	
	function updateUserStatus(userID, isOnline) {
		// Update online status indicator in sidebar or user list
		// This would require tracking user elements
		console.log('User status updated', 'user_id', userID, 'online', isOnline);
	}
	
	function scrollToBottom() {
		const messageList = document.getElementById('message-list');
		if (messageList) {
			messageList.scrollTop = messageList.scrollHeight;
		}
	}
	
	// Handle typing events
	const messageInput = document.getElementById('message-input');
	let typingTimeout = null;
	
	if (messageInput) {
		// Send typing_start when user starts typing
		messageInput.addEventListener('input', function() {
			if (ws && ws.readyState === WebSocket.OPEN) {
				ws.send(JSON.stringify({
					type: 'typing_start',
					data: { channel_id: channelID }
				}));
				
				// Clear existing timeout
				if (typingTimeout) {
					clearTimeout(typingTimeout);
				}
				
				// Send typing_stop after 2 seconds of inactivity
				typingTimeout = setTimeout(function() {
					if (ws && ws.readyState === WebSocket.OPEN) {
						ws.send(JSON.stringify({
							type: 'typing_stop',
							data: { channel_id: channelID }
						}));
					}
				}, 2000);
			}
		});
		
		// Send typing_stop when message is sent
		messageInput.addEventListener('keypress', function(e) {
			if (e.key === 'Enter' && !e.shiftKey) {
				if (typingTimeout) {
					clearTimeout(typingTimeout);
					typingTimeout = null;
				}
				if (ws && ws.readyState === WebSocket.OPEN) {
					ws.send(JSON.stringify({
						type: 'typing_stop',
						data: { channel_id: channelID }
					}));
				}
			}
		});
	}
	
	// Mark channel as read when page becomes visible
	document.addEventListener('visibilitychange', function() {
		if (!document.hidden && ws && ws.readyState === WebSocket.OPEN) {
			ws.send(JSON.stringify({
				type: 'mark_read',
				data: { channel_id: channelID }
			}));
		}
	});
	
	// Mark as read on page load
	if (ws && ws.readyState === WebSocket.OPEN) {
		ws.send(JSON.stringify({
			type: 'mark_read',
			data: { channel_id: channelID }
		}));
	}
	
	// Initial connection
	connect();
	
	// Cleanup on page unload
	window.addEventListener('beforeunload', function() {
		if (ws) {
			ws.close(1000, 'Page unloading');
		}
	});
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
