<template>
<div>
	<!-- dummy element to allow inserting into DOM for simpler bind/destroy -->
	<!-- may use a completely different approach later -->
</div>
</template>

<script>
import $ from 'jquery';

export default {
	props: {
		ruleTree: Object,
		media: Array,
	},
	watch: {
		ruleTree() {
			this.updateStylesheet();
		},
		media(media) {
			if (this.style) {
				if (media) {
					this.style.setAttribute('media', media);
				} else {
					this.style.removeAttribute('media');
				}
			}
		},
	},
	mounted() {
		this.style = document.createElement('style');

		if (this.media) {
			this.style.setAttribute('media', this.media);
		}

		// WebKit hack - thanks https://davidwalsh.name/add-rules-stylesheets
		this.style.appendChild(document.createTextNode(""));

		// Add style tag to the document
		document.head.appendChild(this.style);

		this.updateStylesheet();
	},
	beforeDestroy() {
		if (this.style && this.style.parentNode) {
			this.style.parentNode.removeChild(this.style);
		}
		this.style = null;
	},
	methods: {
		updateStylesheet() {
			// TODO

			// Examine existing rules with this.style.sheet.cssRules
			// https://developer.mozilla.org/en-US/docs/Web/API/CSSStyleSheet

			// Rebuild changed rules and remove old / insert updated rules
			// https://developer.mozilla.org/en-US/docs/Web/API/CSSStyleSheet/deleteRule
			// https://developer.mozilla.org/en-US/docs/Web/API/CSSStyleSheet/insertRule
		},
	},
};
</script>
