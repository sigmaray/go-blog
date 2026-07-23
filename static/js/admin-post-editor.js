(function () {
	'use strict';

	var TINYMCE_BASE = 'https://cdn.jsdelivr.net/npm/tinymce@7.7.2';
	var VISUAL_EDITOR_ID = 'post-editor-visual';

	/**
	 * Returns whether a string looks like empty TinyMCE / HTML placeholder markup.
	 * html is the HTML string to inspect.
	 */
	function isBlankHTML(html) {
		if (!html) {
			return true;
		}
		var normalized = String(html)
			.replace(/&nbsp;/gi, ' ')
			.replace(/<br\s*\/?>/gi, '')
			.replace(/<p>\s*<\/p>/gi, '')
			.replace(/\s+/g, '');
		return normalized === '';
	}

	/**
	 * Normalizes editor HTML so empty visual documents submit as an empty string.
	 * html is the HTML retrieved from the active editor mode.
	 */
	function normalizeSubmitHTML(html) {
		if (isBlankHTML(html)) {
			return '';
		}
		return html;
	}

	/**
	 * Converts HTML to Markdown using Turndown when available.
	 * html is the HTML string to convert.
	 */
	function htmlToMarkdown(html) {
		if (typeof TurndownService !== 'function') {
			return html;
		}
		var service = new TurndownService({
			headingStyle: 'atx',
			codeBlockStyle: 'fenced',
		});
		return service.turndown(html || '');
	}

	/**
	 * Converts Markdown to HTML using marked when available.
	 * markdown is the Markdown string to convert.
	 */
	function markdownToHTML(markdown) {
		if (typeof marked === 'undefined' || typeof marked.parse !== 'function') {
			return markdown;
		}
		return marked.parse(markdown || '', { async: false });
	}

	/**
	 * Reads the current HTML content from the active editing mode.
	 * root is the post-editor root element.
	 * mode is one of "visual", "markdown", or "html".
	 */
	function readHTMLFromMode(root, mode) {
		var htmlField = root.querySelector('[data-post-editor-html]');
		var markdownField = root.querySelector('[data-post-editor-markdown]');

		if (mode === 'html') {
			return htmlField ? htmlField.value : '';
		}
		if (mode === 'markdown') {
			return markdownField ? markdownToHTML(markdownField.value) : '';
		}
		if (typeof tinymce !== 'undefined') {
			var editor = tinymce.get(VISUAL_EDITOR_ID);
			if (editor) {
				return editor.getContent();
			}
		}
		return htmlField ? htmlField.value : '';
	}

	/**
	 * Writes HTML into the target editing mode's controls.
	 * root is the post-editor root element.
	 * mode is one of "visual", "markdown", or "html".
	 * html is the HTML string to apply.
	 */
	function writeHTMLToMode(root, mode, html) {
		var htmlField = root.querySelector('[data-post-editor-html]');
		var markdownField = root.querySelector('[data-post-editor-markdown]');

		if (mode === 'html' && htmlField) {
			htmlField.value = html;
			return;
		}
		if (mode === 'markdown' && markdownField) {
			markdownField.value = htmlToMarkdown(html);
			return;
		}
		if (typeof tinymce !== 'undefined') {
			var editor = tinymce.get(VISUAL_EDITOR_ID);
			if (editor) {
				editor.setContent(html || '');
			}
		}
	}

	/**
	 * Updates mode button pressed state and panel visibility for the chosen mode.
	 * root is the post-editor root element.
	 * mode is one of "visual", "markdown", or "html".
	 */
	function showMode(root, mode) {
		root.querySelectorAll('[data-post-editor-mode]').forEach(function (button) {
			var active = button.getAttribute('data-post-editor-mode') === mode;
			button.classList.toggle('is-active', active);
			button.setAttribute('aria-pressed', active ? 'true' : 'false');
		});

		root.querySelectorAll('[data-post-editor-panel]').forEach(function (panel) {
			var active = panel.getAttribute('data-post-editor-panel') === mode;
			panel.classList.toggle('post-editor__panel--hidden', !active);
			if (active) {
				panel.removeAttribute('hidden');
			} else {
				panel.setAttribute('hidden', 'hidden');
			}
		});

		root.setAttribute('data-post-editor-active-mode', mode);
	}

	/**
	 * Switches the editor to a new mode, converting content from the previous mode.
	 * root is the post-editor root element.
	 * nextMode is the mode to activate ("visual", "markdown", or "html").
	 */
	function switchMode(root, nextMode) {
		var currentMode = root.getAttribute('data-post-editor-active-mode') || 'visual';
		if (currentMode === nextMode) {
			return;
		}

		var html = readHTMLFromMode(root, currentMode);
		writeHTMLToMode(root, nextMode, html);
		showMode(root, nextMode);

		if (nextMode === 'visual' && typeof tinymce !== 'undefined') {
			var editor = tinymce.get(VISUAL_EDITOR_ID);
			if (editor) {
				editor.focus();
			}
		}
	}

	/**
	 * Copies the active mode's content into the named HTML textarea before submit.
	 * root is the post-editor root element.
	 */
	function syncHTMLField(root) {
		var mode = root.getAttribute('data-post-editor-active-mode') || 'visual';
		var htmlField = root.querySelector('[data-post-editor-html]');
		if (!htmlField) {
			return;
		}
		htmlField.value = normalizeSubmitHTML(readHTMLFromMode(root, mode));
	}

	/**
	 * Initializes TinyMCE on the visual panel textarea.
	 * root is the post-editor root element.
	 * initialHTML is the HTML loaded from the form field.
	 * Returns a Promise that resolves when TinyMCE is ready.
	 */
	function initTinyMCE(root, initialHTML) {
		var visualField = root.querySelector('#' + VISUAL_EDITOR_ID);
		if (!visualField || typeof tinymce === 'undefined') {
			return Promise.resolve();
		}

		visualField.value = initialHTML || '';

		return tinymce.init({
			selector: '#' + VISUAL_EDITOR_ID,
			license_key: 'gpl',
			base_url: TINYMCE_BASE,
			suffix: '.min',
			menubar: false,
			branding: false,
			promotion: false,
			plugins: 'lists link table autoresize',
			toolbar:
				'undo redo | styles | bold italic underline strikethrough | alignleft aligncenter alignright | bullist numlist outdent indent | link table | removeformat',
			height: 420,
			min_height: 320,
			convert_urls: false,
			setup: function (editor) {
				editor.on('init', function () {
					editor.setContent(initialHTML || '');
				});
			},
		});
	}

	/**
	 * Wires mode buttons and form submit sync for one post-editor instance.
	 * root is the element marked with data-post-editor.
	 */
	function initEditor(root) {
		var htmlField = root.querySelector('[data-post-editor-html]');
		if (!htmlField) {
			return;
		}

		var initialHTML = htmlField.value || '';
		var form = root.closest('form');

		showMode(root, 'visual');

		initTinyMCE(root, initialHTML).then(function () {
			root.setAttribute('data-post-editor-ready', 'true');
		});

		root.querySelectorAll('[data-post-editor-mode]').forEach(function (button) {
			button.addEventListener('click', function () {
				var mode = button.getAttribute('data-post-editor-mode');
				if (mode) {
					switchMode(root, mode);
				}
			});
		});

		if (form) {
			form.addEventListener('submit', function () {
				syncHTMLField(root);
			});
		}
	}

	/**
	 * Finds and initializes every post editor on the page.
	 * No-ops on pages without data-post-editor (dashboard, users, public site, etc.).
	 */
	function initAll() {
		document.querySelectorAll('[data-post-editor]').forEach(function (root) {
			initEditor(root);
		});
	}

	if (document.readyState === 'loading') {
		document.addEventListener('DOMContentLoaded', initAll);
	} else {
		initAll();
	}
})();
