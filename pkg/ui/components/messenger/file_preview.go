package messenger

import (
	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

// FilePreviewContainer creates a file preview container using Alpine.js.
// This replaces innerHTML manipulations with Alpine.js reactive state.
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
				// Listen for file input changes
				const fileInput = document.getElementById('file-input');
				if (fileInput) {
					fileInput.addEventListener('change', (e) => {
						this.handleFiles(e.target.files);
					});
				}
				// Listen for drag & drop events on textarea
				const textarea = document.getElementById('message-input');
				if (textarea) {
					textarea.addEventListener('drop', (e) => {
						e.preventDefault();
						this.handleFiles(e.dataTransfer.files);
						// Update file input
						const dataTransfer = new DataTransfer();
						for (let i = 0; i < e.dataTransfer.files.length; i++) {
							dataTransfer.items.add(e.dataTransfer.files[i]);
						}
						if (fileInput) {
							fileInput.files = dataTransfer.files;
						}
					});
					textarea.addEventListener('dragover', (e) => {
						e.preventDefault();
					});
				}
			},
			handleFiles(fileList) {
				// Convert FileList to Array and add to existing files
				const newFiles = Array.from(fileList);
				this.files = [...this.files, ...newFiles];
				this.updatePreviews();
			},
			removeFile(index) {
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
				this.files = [];
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
		// File previews rendered via Alpine.js x-for (minimal JS for file handling)
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
