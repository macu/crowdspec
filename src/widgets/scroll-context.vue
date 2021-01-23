<template>
<div class="scroll-context">
	<div ref="contextBar" class="context-bar">
		<span class="bar-label">{{barLabel}}</span>
		<span class="bar-content">{{barContent}}</span>
	</div>
	<slot></slot>
</div>
</template>

<script>
import $ from 'jquery';
import viewport from '../viewport.js';
import {contextStack} from './scroll-context-stack.vue';

export default {
	props: {
		type: String,
		content: Object,
	},
	computed: {
		barLabel() {
			if (this.type === 'block') {
				if (this.content.refItem) {
					if (this.content.refType === 'url') {
						return 'URL (Block)';
					}
					return this.content.refType.charAt(0).toUpperCase() +
						this.content.refType.slice(1);
				}
				return 'Block';
			}
			return this.type.charAt(0).toUpperCase() +
				this.type.slice(1);
		},
		barContent() {
			switch (this.type) {
				case 'spec':
				case 'subspec':
					return this.content.name;
				case 'block':
					let content;
					if (content = (this.content.title || '').trim()) {
						return content;
					}
					if (content = (this.content.body || '').trim()) {
						return content;
					}
					if (this.content.refItem) {
						switch (this.content.refType) {
							case 'subspec':
								return this.content.refItem.name;
							case 'video':
							case 'url':
								return this.content.refItem.title || this.content.refItem.url;
						}
					}
					return '';
				case 'comment':
					return this.content.body;
			}
		},
		inViewport() {
			return viewport.testElementWithinViewport(this.$el, viewport.dims);
		},
	},
	mounted() {


	},
	beforeDestroy() {
		// TODO remove context bar
		// TODO remove context stack if empty
	},
};
</script>

<style lang="scss">
.context-bar-wrapper {
	position: relative;
	>.context-bar {
		position: fixed;
	}
}
</style>
