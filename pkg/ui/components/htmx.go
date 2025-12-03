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
				evt.detail.parameters['csrf'] = '%s';
			}
		})
	`

	const htmxModal = `
		document.body.addEventListener('htmx:afterSwap', function(evt) {
			// Автоматически открываем модальное окно, если был вставлен dialog элемент
			const modal = document.getElementById('channel-create-modal');
			if (modal && modal.tagName === 'DIALOG' && typeof modal.showModal === 'function') {
				// Используем requestAnimationFrame для гарантии, что элемент полностью в DOM
				requestAnimationFrame(function() {
					modal.showModal();
				});
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
