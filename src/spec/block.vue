<template>
<li :class="classes">

	<div class="content">

		<div v-if="block.title" class="title">{{block.title}}</div>

		<!-- TODO ref item here -->

		<div v-if="block.body" class="body">{{block.body}}</div>

	</div>

	<ul v-if="subblocks.length" class="spec-list">
		<spec-block
			v-for="b in subblocks"
			:key="b.id"
			:block="b"
			@prompt-add-subblock="raisePromptAddSubblock"
			/>
	</ul>

	<button @click="promptAddSubblock()">Add subblock</button>

</li>
</template>

<script>
import {END_INDEX} from './const.js';

export default {
	name: 'spec-block',
	props: {
		block: Object,
	},
	data() {
		return {
			subblocks: this.block.subblocks ? this.block.subblocks.slice() : [],
		};
	},
	computed: {
		classes() {
			return ['spec-block', this.block.styleType];
		},
	},
	methods: {
		promptAddSubblock() {
			this.raisePromptAddSubblock(null, this.block.id, END_INDEX, newBlock => {
				this.subblocks.push(newBlock);
			});
		},
		raisePromptAddSubblock(subspaceId, parentId, insertAt, callback) {
			this.$emit('prompt-add-subblock', subspaceId, parentId, insertAt, callback);
		},
	},
};
</script>

<style lang="scss">
.spec-block {
	padding-top: 10px;
	padding-bottom: 10px;

	&:not(:first-child) {
		border-top: thin solid #eee;
	}

	>.content {
		>.title {
			font-weight: bold;
		}
		>.body {
			white-space: pre-wrap;
		}
		>.title+.body {
			margin-top: 10px;
		}
	}

	>ul {
		margin-top: 10px;
	}

	>button {
		margin-top: 10px;
	}
}
</style>
