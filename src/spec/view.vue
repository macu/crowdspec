<template>
<div class="spec-view">

	<ul>
		<spec-block
			v-for="b in blocks"
			:key="b.id"
			:block="b"
			@prompt-add-subblock="promptAddSubblock"
			/>
	</ul>

	<button @click="promptAddBlock()">Add block</button>

	<add-block-modal ref="addBlockModal" :spec-id="spec.id"/>

</div>
</template>

<script>
import SpecBlock from './block.vue';
import AddBlockModal from './add-block-modal.vue';
import {END_INDEX} from './const.js';

export default {
	components: {
		SpecBlock,
		AddBlockModal,
	},
	props: {
		spec: Object,
	},
	data() {
		return {
			blocks: this.spec.blocks ? this.spec.blocks.slice() : [],
		};
	},
	methods: {
		promptAddBlock() {
			this.$refs.addBlockModal.show(null, null, END_INDEX, newBlock => {
				this.blocks.push(newBlock);
			});
		},
		promptAddSubblock(subspaceId, parentId, insertAt, callback) {
			this.$refs.addBlockModal.show(subspaceId, parentId, insertAt, callback);
		},
	},
};
</script>

<style lang="scss">
.spec-view {
}
</style>
