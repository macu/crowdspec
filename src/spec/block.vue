<template>
<li :class="classes">
	<div v-if="block.title" class="title">{{block.title}}</div>
	<div v-if="block.body" class="body">{{block.body}}</div>
	<ul v-if="subblocks.length">
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
			return ['spec-block', this.block.type];
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

	>.title {
		font-weight: bold;
	}

	>.body {
		white-space: pre-wrap;
	}

	>ul {
		margin-top: 10px;
	}

	>button {
		margin-top: 10px;
	}
}
</style>
