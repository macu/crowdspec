<template>
<div class="scroll-context-stack">
	<div ref="stack" class="stack">
	</div>
	<slot></slot>
</div>
</template>

<!--
Stack floats fixed at top of viewport.
When scroll past start of scroll-context, context bar fades in at bottom of stack.
-->

<script>
import $ from 'jquery';
import {isBefore} from '../utils.js';

export let contextStack = null;

export default {
	mounted() {
		// Only one instance is supported
		if (!contextStack) {
			contextStack = this;
		} else {
			console.error('multiple scroll-context-stack instances');
		}
	},
	beforeDestroy() {
		if (contextStack === this) {
			contextStack = null;
		}
	},
	methods: {
		addContextBar(vm) {
			vm.$placeholder = $('<span class="context-bar-placeholder"/>');
			vm.$contextBar = $(vm.$refs.contextBar);
			vm.$contextBar.data('vm', vm);
			vm.$contextBar.replaceWith(vm.$placeholder);
			$(this.$refs.stack).append(vm.$contextBar);
			this.sortElements();
		},
		removeContextBar(vm) {
			vm.$placeholder.replaceWith(vm.$contextBar);
			vm.$placeholder = null;
			vm.$contextBar = null;
		},
		sortElements() {
			let $stack = $(this.$refs.stack);
			// sort context-bar by order of appearance of placeholders in DOM
			let sorted = $stack.find('>*').toArray().sort(function (e1, e2) {
				if (e1 === e2) {
					return 0;
				}
				let place1 = $(e1).data('vm').$placeholder;
				let place2 = $(e2).data('vm').$placeholder;
				return isBefore(place1, place2) ? -1 : 1;
			});
			// Move into reordered positions
			$(sorted).appendTo($stack);
		},
	},
};
</script>

<style lang="scss">
.scroll-context-stack {
	>.stack {
		position: fixed;
		top: 0;
		left: 0;
		width: 100%;
		>.context-bar {
			display: none;

		}
	}
}
</style>
