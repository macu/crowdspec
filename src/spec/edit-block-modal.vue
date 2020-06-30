<template>
<el-dialog
	:title="block ? 'Edit block' : 'Add block'"
	:visible.sync="showing"
	:width="$store.getters.dialogSmallWidth"
	:close-on-click-modal="false"
	@closed="closed()"
	class="spec-edit-block-modal">

	<el-radio-group v-model="styleType" >
		<el-radio label="bullet">Bullet point</el-radio>
		<el-radio label="numbered">Numbered point</el-radio>
		<el-radio label="none">Indented block</el-radio>
	</el-radio-group>

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
		<el-button @click="submit()" type="primary" :disabled="disableSubmit">{{block ? 'Save' : 'Add'}}</el-button>
	</span>

</el-dialog>
</template>

<script>
import $ from 'jquery';
import {ajaxCreateBlock, ajaxSaveBlock} from './ajax.js';

export default {
	props: {
		specId: Number,
	},
	data() {
		return {
			// user inputs
			styleType: 'bullet',
			title: '',
			body: '',
			// passed in
			block: null,
			subspaceId: null,
			parentId: null,
			insertBeforeId: null,
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
		showAdd(subspaceId, parentId, insertBeforeId, callback) {
			this.subspaceId = subspaceId;
			this.parentId = parentId;
			this.insertBeforeId = insertBeforeId;
			this.callback = callback;
			this.showing = true;
			this.$nextTick(() => {
				$('input', this.$refs.titleInput.$el).focus();
			});
		},
		showEdit(block, callback) {
			this.block = block;
			this.styleType = block.styleType;
			this.title = block.title;
			this.body = block.body;
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
			if (this.block) {
				this.submitSave();
			} else {
				this.submitAdd();
			}
		},
		submitAdd() {
			this.sending = true;
			let callback = this.callback; // in case modal is closed before complete
			ajaxCreateBlock(
				this.specId,
				this.subspaceId,
				this.parentId,
				this.insertBeforeId,
				this.styleType,
				'plaintext', // contentType
				null, // refType
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
		submitSave() {
			this.sending = true;
			let callback = this.callback; // in case modal is closed before complete
			ajaxSaveBlock(
				this.specId,
				this.block.id,
				this.styleType,
				'plaintext', // contentType
				null, // refType
				null, // refId
				this.title,
				this.body
			).then(updatedBlock => {
				callback(updatedBlock);
				this.showing = false;
				this.sending = false;
			}).fail(() => {
				this.sending = false;
			});
		},
		closed() {
			this.block = null;
			this.subspaceId = null;
			this.parentId = null;
			this.insertBeforeId = null;
			this.callback = null;
			this.title = '';
			this.body = '';
		},
	},
};
</script>

<style lang="scss">
.spec-edit-block-modal {
	>.el-dialog {
		>.el-dialog__body {
			>*+* {
				margin-top: 20px;
			}
			>label {
				display: block;
				input, textarea {
					display: block;
					width: 100%;
				}
			}
		}
	}
}
</style>
