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

// DirectMessage renders the direct message page
func DirectMessage(ctx echo.Context, dmID int64, otherUserName string, messages []messengerComponents.MessageData) error {
	r := ui.NewRequest(ctx)
	r.Title = otherUserName

	content := Div(
		Class("flex flex-col h-full"),
		// DM header
		dmHeader(r, dmID, otherUserName),
		// Messages list
		messengerComponents.MessageList(r, messages),
		// Typing indicator (will be updated via WebSocket)
		Div(
			ID("typing-indicator-container"),
			// TypingIndicator will be added via WebSocket updates
		),
		// Message input
		messengerComponents.MessageInput(r, dmID, true), // true = is a DM
		// WebSocket connection script
		dmWebsocketScript(r, dmID),
	)

	return r.Render(messengerLayouts.Messenger, content)
}

func dmHeader(r *ui.Request, dmID int64, userName string) Node {
	return Div(
		Class("border-b border-base-300 p-4 bg-base-100 flex items-center justify-between"),
		Div(
			Class("flex items-center gap-3"),
			Div(
				ID(fmt.Sprintf("dm-avatar-%d", dmID)),
				messengerComponents.UserAvatar(r, dmID, userName, "sm", nil), // nil = unknown online status, will be updated via WebSocket
			),
			H2(
				Class("text-xl font-semibold"),
				Text(userName),
			),
		),
		Div(
			Class("flex items-center gap-2"),
			// User info button
			Button(
				Class("btn btn-sm btn-ghost"),
				Text("ℹ️"),
				Attr("title", "User info"),
				// TODO: Open right panel with user info
			),
		),
	)
}

func dmWebsocketScript(r *ui.Request, dmID int64) Node {
	return Script(
		Raw(`
(function() {
	let ws = null;
	let reconnectAttempts = 0;
	const maxReconnectAttempts = 5;
	const reconnectDelay = 3000; // 3 seconds
	const dmID = ` + fmt.Sprintf("%d", dmID) + `;
	
	// Track typing users
	const typingUsers = new Map(); // userID -> timeout
	
	function connect() {
		const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
		ws = new WebSocket(protocol + '//' + window.location.host + '/ws');
		
		ws.onopen = function() {
			console.log('WebSocket connected for DM');
			reconnectAttempts = 0;
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
			console.log('WebSocket disconnected', 'code', event.code);
			clearTypingIndicator();
			
			if (event.code !== 1000 && reconnectAttempts < maxReconnectAttempts) {
				reconnectAttempts++;
				console.log('Attempting to reconnect...', 'attempt', reconnectAttempts);
				setTimeout(connect, reconnectDelay);
			}
		};
	}
	
	function handleWebSocketMessage(message) {
		switch(message.type) {
			case 'message_new':
				if (message.data.dm_id === dmID) {
					scrollToBottom();
					clearTypingIndicator();
				}
				break;
			case 'user_typing':
				if (message.data.dm_id === dmID) {
					showTypingIndicator(message.data.user_id, message.data.user_name || 'Someone');
				}
				break;
			case 'user_online':
				updateUserStatus(message.data.user_id, true);
				break;
			case 'user_offline':
				updateUserStatus(message.data.user_id, false);
				break;
		}
	}
	
	function showTypingIndicator(userID, userName) {
		if (typingUsers.has(userID)) {
			clearTimeout(typingUsers.get(userID));
		}
		
		const container = document.getElementById('typing-indicator-container');
		if (container) {
			container.innerHTML = '<div id="typing-indicator" class="px-4 py-2 text-sm text-base-content/60 italic">' + 
				userName + ' is typing<span class="inline-block ml-1">...</span></div>';
		}
		
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
		if (!container || typingUsers.size === 0) {
			if (container) container.innerHTML = '';
			return;
		}
		container.innerHTML = '<div id="typing-indicator" class="px-4 py-2 text-sm text-base-content/60 italic">' + 
			'Someone is typing<span class="inline-block ml-1">...</span></div>';
	}
	
	function updateUserStatus(userID, isOnline) {
		// Update online status in DM avatar
		const avatarEl = document.getElementById('dm-avatar-' + dmID);
		if (avatarEl) {
			const statusEl = avatarEl.querySelector('[data-online]');
			if (statusEl) {
				statusEl.setAttribute('data-online', isOnline ? 'true' : 'false');
				statusEl.className = statusEl.className.replace(/bg-(success|base-300)/g, '');
				statusEl.className += ' ' + (isOnline ? 'bg-success' : 'bg-base-300');
			}
		}
	}
	
	function scrollToBottom() {
		const messageList = document.getElementById('message-list');
		if (messageList) {
			messageList.scrollTop = messageList.scrollHeight;
		}
	}
	
	// Handle typing events for DM
	const messageInput = document.getElementById('message-input');
	let typingTimeout = null;
	
	if (messageInput) {
		messageInput.addEventListener('input', function() {
			// For DM, typing events are handled differently (would need DM-specific events)
			// For now, just clear timeout
			if (typingTimeout) {
				clearTimeout(typingTimeout);
			}
		});
	}
	
	connect();
	
	window.addEventListener('beforeunload', function() {
		if (ws) {
			ws.close(1000, 'Page unloading');
		}
	});
})();
		`),
	)
}
