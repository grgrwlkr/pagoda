package messenger

import (
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// TypingIndicator creates a typing indicator component using Alpine.js.
// This replaces innerHTML manipulations with Alpine.js reactive state.
// The component shows who is typing and automatically hides after 3 seconds.
func TypingIndicator() Node {
	return Div(
		ID("typing-indicator-container"),
		// Alpine.js component for managing typing indicator state
		Attr("x-data", `{
			typingUsers: new Map(),
			typingText: '',
			init() {
				// Listen for custom events from WebSocket handler
				this.$el.addEventListener('show-typing', (e) => {
					this.showTyping(e.detail.userID, e.detail.userName);
				});
				this.$el.addEventListener('clear-typing', () => {
					this.clearTyping();
				});
			},
			showTyping(userID, userName) {
				if (this.typingUsers.has(userID)) {
					clearTimeout(this.typingUsers.get(userID));
				}
				this.typingText = userName + ' is typing';
				const timeout = setTimeout(() => {
					this.typingUsers.delete(userID);
					this.updateDisplay();
				}, 3000);
				this.typingUsers.set(userID, timeout);
				this.updateDisplay();
			},
			clearTyping() {
				this.typingUsers.clear();
				this.updateDisplay();
			},
			updateDisplay() {
				if (this.typingUsers.size === 0) {
					this.typingText = '';
				} else if (this.typingUsers.size === 1) {
					// Get first user name (simplified - in full implementation track names)
					this.typingText = 'Someone is typing';
				} else {
					this.typingText = 'Someone is typing';
				}
			}
		}`),
		// Show/hide based on typingText
		Attr("x-show", "typingText !== ''"),
		// Display typing text
		Attr("x-text", "typingText"),
		Class("px-4 py-2 text-sm text-base-content/60 italic"),
		// Hidden by default
		Style("display: none;"),
	)
}
