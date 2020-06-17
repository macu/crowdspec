<template>
<el-dialog
	title="Add block"
	:visible.sync="showing"
	:width="$store.getters.dialogTinyWidth"
	:close-on-click-modal="false"
	@closed="closed()"
	class="spec-add-block-modal">

	<label>
		Title
		<el-input ref="titleInput" v-model="title" clearable/>
	</label>

	<label>
		Body
		<el-input type="textarea" v-model="body" :autosize="{minRows: 2}"/>
	</label>

	<span slot="footer" class="dialog-footer" v-loading="sending">
		<el-button @click="showing = false">Cancel</el-button>
		<el-button @click="submit()" type="primary" :disabled="disableSubmit">Add</el-button>
	</span>

</el-dialog>
</template>

<script>
import $ from 'jquery';
import {ajaxCreateBlock} from './ajax.js';

export default {
	props: {
		specId: Number,
	},
	data() {
		return {
			// user inputs
			title: '',
			body: '',
			// passed in
			subspaceId: null,
			parentId: null,
			insertAt: null,
			callback: null,
			// state
			showing: false,
			sending: false,
		};
	},
	computed: {
		disableSubmit() {
			return !this.title.trim();
		},
	},
	methods: {
		show(subspaceId, parentId, insertAt, callback) {
			this.subspaceId = subspaceId;
			this.parentId = parentId;
			this.insertAt = insertAt;
			this.callback = callback;
			this.showing = true;
			this.$nextTick(() => {
				$('input', this.$refs.titleInput.$el).focus();
			});
		},
		submit() {
			if (this.disableSubmit) {
				return;
			}
			this.sending = true;
			let callback = this.callback; // in case modal is closed before complete
			ajaxCreateBlock(
				this.specId,
				this.subspaceId,
				this.parentId,
				this.insertAt,
				'text', // blockType
				null, // refId
				this.title,
				this.body
			).then(newBlock => {
				callback(newBlock);
				this.showing = false;
				this.sending = false;
			}).fail(() => {
				this.sending = false;
			});
		},
		closed() {
			this.callback = null;
			this.title = '';
			this.body = '';
		},
	},
};
</script>

<style lang="scss">
.spec-add-block-modal {
	label {
		display: block;
		input, textarea {
			display: block;
			width: 100%;
		}
	}
	label+label {
		margin-top: 10px;
	}
}
</style>
