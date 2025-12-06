package messenger

import (
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// ThreadPanelStore creates an Alpine.js store for managing thread panel state.
// This replaces global window.* functions with Alpine.js store.
// The store provides methods for opening, closing, and switching thread panels.
func ThreadPanelStore() Node {
	return Script(
		Raw(`
		// Alpine.js store for thread panel management
		// This replaces global window.* functions
		document.addEventListener('alpine:init', () => {
			Alpine.store('threadPanel', {
				currentThreadId: null,
				scrollPositions: new Map(),
				
				openThreadPanel(threadId) {
					const rightPanel = document.getElementById('right-panel');
					const backdrop = document.getElementById('right-panel-backdrop');
					
					if (!rightPanel) return;
					
					// Set thread ID
					this.currentThreadId = threadId;
					rightPanel.setAttribute('data-thread-id', threadId);
					
					// Show panel (remove hidden class)
					rightPanel.classList.remove('hidden');
					
					// Show backdrop on mobile
					if (backdrop) {
						backdrop.classList.remove('hidden');
					}
					
					// Trigger animation (slide in from right on mobile, fade in on desktop)
					requestAnimationFrame(() => {
						rightPanel.classList.remove('translate-x-full');
						if (backdrop) {
							backdrop.style.opacity = '1';
						}
					});
					
					// Restore scroll position if available
					const savedScroll = this.scrollPositions.get(threadId);
					if (savedScroll !== undefined) {
						const contentEl = rightPanel.querySelector('#thread-panel-content');
						if (contentEl) {
							setTimeout(() => {
								contentEl.scrollTop = savedScroll;
							}, 100);
						}
					} else {
						// Scroll to bottom if no saved position
						setTimeout(() => {
							const contentEl = rightPanel.querySelector('#thread-panel-content');
							if (contentEl) {
								contentEl.scrollTop = contentEl.scrollHeight;
							}
						}, 100);
					}
				},
				
				closeThreadPanel() {
					const rightPanel = document.getElementById('right-panel');
					const backdrop = document.getElementById('right-panel-backdrop');
					
					if (!rightPanel) return;
					
					// Save scroll position before closing
					const threadId = rightPanel.getAttribute('data-thread-id');
					if (threadId) {
						const contentEl = rightPanel.querySelector('#thread-panel-content');
						if (contentEl) {
							this.scrollPositions.set(threadId, contentEl.scrollTop);
						}
					}
					
					// Hide backdrop first (fade out)
					if (backdrop) {
						backdrop.style.opacity = '0';
						setTimeout(() => {
							backdrop.classList.add('hidden');
						}, 300);
					}
					
					// Slide out on mobile, fade out on desktop
					rightPanel.classList.add('translate-x-full');
					
					// Hide panel after animation completes
					setTimeout(() => {
						rightPanel.classList.add('hidden');
						rightPanel.removeAttribute('data-thread-id');
						this.currentThreadId = null;
					}, 300);
				},
				
				switchThreadPanel(newThreadId) {
					const rightPanel = document.getElementById('right-panel');
					if (!rightPanel) return;
					
					const currentThreadId = rightPanel.getAttribute('data-thread-id');
					
					// If same thread, do nothing
					if (currentThreadId && currentThreadId === String(newThreadId)) {
						return;
					}
					
					// Save current scroll position
					if (currentThreadId) {
						const contentEl = rightPanel.querySelector('#thread-panel-content');
						if (contentEl) {
							this.scrollPositions.set(currentThreadId, contentEl.scrollTop);
						}
					}
					
					// Close current panel (without animation for smooth transition)
					rightPanel.classList.add('hidden');
					
					// Open new thread panel
					setTimeout(() => {
						this.openThreadPanel(newThreadId);
					}, 50);
				},
				
				scrollToNewReply() {
					const rightPanel = document.getElementById('right-panel');
					if (!rightPanel || rightPanel.classList.contains('hidden')) return;
					
					const contentEl = rightPanel.querySelector('#thread-panel-content');
					if (!contentEl) return;
					
					// Smooth scroll to bottom
					contentEl.scrollTo({
						top: contentEl.scrollHeight,
						behavior: 'smooth'
					});
				}
			});
		});
		`),
	)
}
