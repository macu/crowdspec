<template>
<el-dialog
	title="Block community"
	:visible.sync="showing"
	:width="$store.getters.dialogLargeWidth"
	:close-on-click-modal="false"
	@closed="closed()"
	class="spec-block-community-modal">

	<p v-if="loading">Loading...</p>

	<template v-else-if="block">

		<block-preview
			:block="block"
			@play-video="raisePlayVideo"
			/>

		<el-button @click="addComment()" type="primary">Add comment</el-button>

	</template>

	<span slot="footer" class="dialog-footer">
		<el-button @click="showing = false">Close</el-button>
	</span>

</el-dialog>
</template>

<script>
import BlockPreview from './block-preview.vue';
import {ajaxLoadBlockCommunity} from './ajax.js';

export default {
	components: {
		BlockPreview,
	},
	props: {
		specId: Number,
	},
	data() {
		return {
			showing: false,
			loading: false,
			block: null,
		};
	},
	methods: {
		show(blockId) {
			this.loading = true;
			this.showing = true;
			ajaxLoadBlockCommunity(this.specId, blockId).then(response => {
				this.loading = false;
				this.block = response.block;
			}).fail(() => {
				this.loading = false;
				this.showing = false;
			});
		},
		raisePlayVideo(urlObject) {
			this.$emit('play-video', urlObject);
		},
		closed() {
			this.block = null;
		},
	},
};
</script>

<style lang="scss">
.spec-block-community-modal {

}
</style>
