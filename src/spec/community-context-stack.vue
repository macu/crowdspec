<template>
<div class="community-context-stack" :class="{empty}">
	<transition
		v-for="(s, i) in stack"
		:key="s.type + '-' + s.element.id"
		name="fade" appear>
		<stack-bar
			:element="s.element"
			:element-type="s.type"
			@click="jumpStack(i)"
			/>
	</transition>
</div>
</template>

<script>
import StackBar from './community-context-stack-bar.vue';
import {idsEq} from '../utils.js';

export default {
	emits: ['pop-stack'],
	components: {
		StackBar,
	},
	data() {
		return {
			stack: [],
			empty: true,
			cachedContext: [],
		};
	},
	computed: {
		available() {
			return this.stack.length;
		},
	},
	methods: {
		replaceStack(stack) {
			this.stack = stack;
			this.cachedContext = [];
			this.checkEmpty();
		},
		pushStack(type, element, onAdjustUnread, onAdjustComments) {
			this.empty = false;
			this.stack.push({
				type,
				element,
				onAdjustUnread,
				onAdjustComments,
			});
		},
		popStack() {
			// called when comment is deleted
			if (this.stack.length) {
				let items = this.stack.splice(this.stack.length - 1, 1); // remove last item
				let item = items[0];
				this.$emit('pop-stack', item.type, item.element.id,
					item.onAdjustUnread, item.onAdjustComments);
				this.cacheItemHandlers(item);
				this.checkEmpty();
				return item;
			}
			return null;
		},
		jumpStack(i) {
			let items = this.stack.splice(i, this.stack.length - i); // remove items
			let item = items[0];
			this.$emit('pop-stack', item.type, item.element.id,
				item.onAdjustUnread, item.onAdjustComments);
			for (var i = 0; i < items.length; i++) {
				this.cacheItemHandlers(items[i]);
			}
			this.checkEmpty();
			return item;
		},
		cacheItemHandlers(item) {
			// Delete old cache entry
			for (var i = 0; i < this.cachedContext.length; i++) {
				let c = this.cachedContext[i];
				if (
					c.type === item.type &&
					idsEq(c.element.id, item.element.id)
				) {
					this.cachedContext.splice(i, 1); // remove from array
					break;
				}
			}
			if (item.onAdjustUnread || item.onAdjustComments) {
				// Cached popped item to preserve handlers
				this.cachedContext.push(item);
			}
		},
		retrieveCachedHandlers(type, elementId) {
			// Attempt to restore handlers from cached context
			for (var i = 0; i < this.cachedContext.length; i++) {
				let c = this.cachedContext[i];
				if (
					c.type === type &&
					idsEq(c.element.id, elementId)
				) {
					let items = this.cachedContext.splice(i, 1);
					let cachedItem = items[0];
					return cachedItem;
				}
			}
			return null;
		},
		getParentContext() {
			if (this.stack.length) {
				return this.stack[this.stack.length - 1];
			}
			return null;
		},
		clearStack() {
			this.stack = [];
			this.cachedContext = [];
			this.empty = true;
		},
		checkEmpty() {
			return this.empty = !this.stack.length;
		},
	},
}
</script>

<style lang="scss">
.community-context-stack {

	&.empty {
		display: none;
	}

	.community-context-stack-bar {
		&:first-child {
			margin-top: 0;
		}
		&:last-child {
			margin-bottom: 0;
		}
	}

	.fade-enter-active {
		transition: opacity .5s;
	}
	.fade-enter-from {
		opacity: 0;
	}
	//.fade-leave-active {}
	//.fade-leave-to {}
}
</style>
