package messenger

import (
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// ModalStore creates an Alpine.js store for managing modal windows.
// This replaces direct getElementById().close() calls with Alpine.js store methods.
// The store provides methods for opening and closing modals.
func ModalStore() Node {
	return Script(
		Raw(`
		// Alpine.js store for modal management
		// This replaces direct getElementById().close() calls
		document.addEventListener('alpine:init', () => {
			Alpine.store('modal', {
				openModal(modalId) {
					const modal = document.getElementById(modalId);
					if (modal && modal.tagName === 'DIALOG') {
						modal.showModal();
					}
				},
				
				closeModal(modalId) {
					const modal = document.getElementById(modalId);
					if (modal && modal.tagName === 'DIALOG') {
						modal.close();
					}
				}
			});
		});
		`),
	)
}
