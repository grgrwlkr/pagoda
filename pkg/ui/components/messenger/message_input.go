package messenger

import (
	"github.com/mikestefanello/pagoda/pkg/ui"
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// MessageInput renders the message input form at the bottom of the chat
// channelIDOrDMID: channel ID for channels, DM ID for direct messages
// isDM: true if this is for a direct message, false for a channel
func MessageInput(r *ui.Request, channelIDOrDMID int64, isDM bool) Node {
	var actionRoute string
	if isDM {
		actionRoute = r.Path("messenger.dm.message.create", channelIDOrDMID)
	} else {
		actionRoute = r.Path("messenger.message.create", channelIDOrDMID)
	}

	return Div(
		Class("border-t border-base-300 p-4 bg-base-100"),
		// File previews container (hidden by default, shown via JavaScript)
		Div(
			ID("file-preview-container"),
			Class("mb-2 flex flex-wrap gap-2 hidden"),
		),
		Form(
			Class("flex gap-2"),
			ID("message-form"),
			Method("POST"),
			Action(actionRoute),
			Attr("enctype", "multipart/form-data"),
			Attr("hx-post", actionRoute),
			Attr("hx-target", "#message-list"),
			Attr("hx-swap", "beforeend"),
			Attr("hx-encoding", "multipart/form-data"),
			Attr("hx-on::after-request", "this.querySelector('textarea').value = ''; this.querySelector('textarea').style.height = 'auto'; document.getElementById('file-preview-container').innerHTML = ''; document.getElementById('file-preview-container').classList.add('hidden'); document.getElementById('file-input').value = '';"),
			// CSRF token
			If(r.CSRF != "", Input(
				Type("hidden"),
				Name("csrf"),
				Value(r.CSRF),
			)),
			Div(
				Class("flex-1"),
				Textarea(
					ID("message-input"),
					Name("content"),
					Class("textarea textarea-bordered w-full resize-none"),
					Placeholder("Type a message... (drag & drop files here)"),
					Rows("1"),
					Attr("x-data", `{
						resize() {
							this.$el.style.height = "auto";
							this.$el.style.height = this.$el.scrollHeight + "px";
						}
					}`),
					Attr("@input", "resize()"),
					Attr("@keydown.enter", "if(!event.shiftKey) { event.preventDefault(); document.getElementById('message-form').requestSubmit(); }"),
					Attr("ondrop", "handleFileDrop(event); return false;"),
					Attr("ondragover", "event.preventDefault(); return false;"),
				),
			),
			Div(
				Class("flex flex-col gap-2"),
				// File upload button
				Label(
					Class("btn btn-circle btn-ghost cursor-pointer"),
					Title("Upload file"),
					Input(
						Type("file"),
						ID("file-input"),
						Name("files"),
						Class("hidden"),
						Attr("multiple"),
						Attr("onchange", "handleFileSelect(event)"),
					),
					Text("📎"),
				),
				// Send button
				Button(
					Type("submit"),
					ID("message-send"),
					Class("btn btn-primary btn-circle"),
					Title("Send message"),
					Text("➤"),
				),
			),
		),
		// JavaScript for file handling
		Script(
			Raw(`
			function handleFileSelect(event) {
				const files = event.target.files;
				showFilePreviews(files);
			}
			
			function handleFileDrop(event) {
				event.preventDefault();
				const files = event.dataTransfer.files;
				const fileInput = document.getElementById('file-input');
				const dataTransfer = new DataTransfer();
				for (let i = 0; i < files.length; i++) {
					dataTransfer.items.add(files[i]);
				}
				fileInput.files = dataTransfer.files;
				showFilePreviews(files);
			}
			
			function showFilePreviews(files) {
				const container = document.getElementById('file-preview-container');
				container.innerHTML = '';
				
				if (files.length === 0) {
					container.classList.add('hidden');
					return;
				}
				
				container.classList.remove('hidden');
				
				for (let i = 0; i < files.length; i++) {
					const file = files[i];
					const div = document.createElement('div');
					div.className = 'relative inline-block p-2 border border-base-300 rounded-lg bg-base-200';
					
					if (file.type.startsWith('image/')) {
						const img = document.createElement('img');
						img.src = URL.createObjectURL(file);
						img.className = 'max-w-20 max-h-20 object-cover rounded';
						div.appendChild(img);
					} else {
						const icon = document.createElement('div');
						icon.className = 'text-2xl';
						icon.textContent = '📎';
						div.appendChild(icon);
					}
					
					const name = document.createElement('div');
					name.className = 'text-xs truncate max-w-20';
					name.textContent = file.name;
					div.appendChild(name);
					
					const removeBtn = document.createElement('button');
					removeBtn.type = 'button';
					removeBtn.className = 'absolute -top-1 -right-1 btn btn-xs btn-circle btn-error';
					removeBtn.textContent = '×';
					removeBtn.onclick = function() {
						removeFile(i);
					};
					div.appendChild(removeBtn);
					
					container.appendChild(div);
				}
			}
			
			function removeFile(index) {
				const fileInput = document.getElementById('file-input');
				const dataTransfer = new DataTransfer();
				for (let i = 0; i < fileInput.files.length; i++) {
					if (i !== index) {
						dataTransfer.items.add(fileInput.files[i]);
					}
				}
				fileInput.files = dataTransfer.files;
				showFilePreviews(fileInput.files);
			}
			`),
		),
	)
}
