package components

import (
	"fmt"

	"github.com/mikestefanello/pagoda/pkg/ui"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

func HtmxListeners(r *ui.Request) Node {
	const htmxErr = `
		document.body.addEventListener('htmx:beforeSwap', function(evt) {
			if (evt.detail.xhr.status >= 400){
				evt.detail.shouldSwap = true;
				evt.detail.target = htmx.find("body");
			}
		});
	`

	const htmxCSRF = `
		document.body.addEventListener('htmx:configRequest', function(evt)  {
			if (evt.detail.verb !== "get") {
				// Get CSRF token from form (if available) - this ensures we use the latest token
				var csrfToken = null;
				var form = evt.detail.elt;
				
				// First: try to get token from the form being submitted
				if (form && form.tagName === 'FORM') {
					var csrfInput = form.querySelector('input[name="csrf"]');
					if (csrfInput && csrfInput.value) {
						csrfToken = csrfInput.value;
					}
				}
				
				// Fallback: get token from _csrf cookie (source of truth for Echo)
				if (!csrfToken) {
					var cookies = document.cookie.split(';');
					for (var i = 0; i < cookies.length; i++) {
						var cookie = cookies[i].trim();
						if (cookie.startsWith('_csrf=')) {
							csrfToken = cookie.substring(7);
							break;
						}
					}
				}
				
				// Final fallback: use initial token from page load
				if (!csrfToken) {
					csrfToken = '%s';
				}
				
				// Set CSRF token in request parameters
				evt.detail.parameters['csrf'] = csrfToken;
			}
		})
	`

	const htmxModal = `
		document.body.addEventListener('htmx:afterSwap', function(evt) {
			// Автоматически открываем модальное окно, если был вставлен dialog элемент
			// Use Alpine.js store for modal management
			if (window.Alpine && window.Alpine.store && window.Alpine.store('modal')) {
				// Check if channel-create-modal was inserted
				const modal = document.getElementById('channel-create-modal');
				if (modal && modal.tagName === 'DIALOG') {
					requestAnimationFrame(function() {
						window.Alpine.store('modal').openModal('channel-create-modal');
					});
				}
				// Check if workspace-create-modal was inserted
				const workspaceModal = document.getElementById('workspace-create-modal');
				if (workspaceModal && workspaceModal.tagName === 'DIALOG') {
					requestAnimationFrame(function() {
						window.Alpine.store('modal').openModal('workspace-create-modal');
					});
				}
			} else {
				// Fallback if Alpine.js store not available
				const modal = document.getElementById('channel-create-modal');
				if (modal && modal.tagName === 'DIALOG' && typeof modal.showModal === 'function') {
					requestAnimationFrame(function() {
						modal.showModal();
					});
				}
			}
		});
	`

	return Group{
		Script(Raw(htmxErr)),
		Iff(len(r.CSRF) > 0, func() Node {
			return Script(Raw(fmt.Sprintf(htmxCSRF, r.CSRF)))
		}),
		Script(Raw(htmxModal)),
	}
}

func HxBoost() Node {
	return Attr("hx-boost", "true")
}
