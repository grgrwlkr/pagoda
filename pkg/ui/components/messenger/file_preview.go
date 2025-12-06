package messenger

import (
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// FilePreviewContainer creates a file preview container using Alpine.js.
// This uses Alpine.js for File API (File API requires JS).
// The component shows file previews and allows removing files before upload.
func FilePreviewContainer() Node {
	return Div(
		ID("file-preview-container"),
		Class("mb-2 flex flex-wrap gap-2"),
		// Alpine.js component for managing file previews
		Attr("x-data", `{
			files: [],
			previews: [],
			init() {
				// Listen for file events from other components
				this.$el.addEventListener('file-input-change', (e) => {
					if (e.detail.files) {
						this.handleFiles(e.detail.files);
					}
				});
				this.$el.addEventListener('file-drop', (e) => {
					if (e.detail.files) {
						this.handleFiles(e.detail.files);
						// Update file input
						const fileInput = document.getElementById('file-input');
						if (fileInput) {
							const dataTransfer = new DataTransfer();
							for (let i = 0; i < e.detail.files.length; i++) {
								dataTransfer.items.add(e.detail.files[i]);
							}
							fileInput.files = dataTransfer.files;
						}
					}
				});
				this.$el.addEventListener('clear-files', () => {
					this.clearFiles();
				});
			},
			handleFiles(fileList) {
				// Convert FileList to Array and add to existing files
				const newFiles = Array.from(fileList);
				this.files = [...this.files, ...newFiles];
				this.updatePreviews();
			},
			removeFile(index) {
				// Revoke object URL if it's an image
				if (this.previews[index] && this.previews[index].url) {
					URL.revokeObjectURL(this.previews[index].url);
				}
				this.files.splice(index, 1);
				// Update file input
				const fileInput = document.getElementById('file-input');
				if (fileInput) {
					const dataTransfer = new DataTransfer();
					for (let i = 0; i < this.files.length; i++) {
						dataTransfer.items.add(this.files[i]);
					}
					fileInput.files = dataTransfer.files;
				}
				this.updatePreviews();
			},
			clearFiles() {
				// Revoke all object URLs
				this.previews.forEach(preview => {
					if (preview.url) {
						URL.revokeObjectURL(preview.url);
					}
				});
				this.files = [];
				// Clear file input
				const fileInput = document.getElementById('file-input');
				if (fileInput) {
					fileInput.value = '';
				}
				this.updatePreviews();
			},
			updatePreviews() {
				// Generate preview data for each file
				this.previews = this.files.map((file, index) => {
					const preview = {
						index: index,
						name: file.name,
						type: file.type,
						url: file.type.startsWith('image/') ? URL.createObjectURL(file) : null
					};
					return preview;
				});
			}
		}`),
		// Show/hide based on files array
		Attr("x-show", "files.length > 0"),
		// Hidden by default
		Style("display: none;"),
		// File previews rendered via Alpine.js x-for (File API requires JS)
		// Note: Alpine.js x-for requires template element, so we use Raw HTML
		Raw(`
		<template x-for="(preview, index) in previews" :key="index">
			<div class="relative inline-block p-2 border border-base-300 rounded-lg bg-base-200">
				<template x-if="preview.url">
					<img :src="preview.url" class="max-w-20 max-h-20 object-cover rounded" />
				</template>
				<template x-if="!preview.url">
					<div class="text-2xl">📎</div>
				</template>
				<div class="text-xs truncate max-w-20" x-text="preview.name"></div>
				<button type="button" @click="removeFile(index)" class="absolute -top-1 -right-1 btn btn-xs btn-circle btn-error">×</button>
			</div>
		</template>
		`),
	)
}
