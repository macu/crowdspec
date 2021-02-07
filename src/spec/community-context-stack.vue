<template>
<div class="community-context-stack" :class="{empty}">
	<transition
		v-for="(s, i) in stack"
		:key="s.targetType + '-' + s.target.id"
		@after-leave="checkEmpty"
		name="fade" appear>
		<stack-bar
			:target="s.target"
			:target-type="s.targetType"
			@click.native="jumpStack(i)"
			/>
	</transition>
</div>
</template>

<script>
import StackBar from './community-context-stack-bar.vue';

export default {
	components: {
		StackBar,
	},
	data() {
		return {
			stack: [],
			empty: true,
		};
	},
	computed: {
		available() {
			return this.stack.length;
		},
	},
	methods: {
		pushStack(targetType, target, onAdjustUnread, onAdjustComments) {
			this.empty = false;
			this.stack.push({
				targetType,
				target,
				onAdjustUnread,
				onAdjustComments,
			});
		},
		popStack() {
			if (this.stack.length) {
				return this.jumpStack(this.stack.length - 1);
			}
			return null;
		},
		jumpStack(i) {
			let item = this.stack[i];
			this.$emit('pop-stack', item.targetType, item.target.id,
				item.onAdjustUnread, item.onAdjustComments);
			this.stack = this.stack.slice(0, i);
			return item;
		},
		clearStack() {
			this.stack = [];
		},
		checkEmpty() {
			return this.empty = !this.stack.length;
		},
	},
}
</script>

<style lang="scss">
.community-context-stack {

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
	.fade-enter {
		opacity: 0;
	}
	//.fade-leave-active {}
	//.fade-leave-to {}
}
</style>
