<template>
<el-dialog
	:title="block ? 'Edit block' : 'Add block'"
	:visible.sync="showing"
	:width="$store.getters.dialogSmallWidth"
	:close-on-click-modal="false"
	@closed="closed()"
	class="spec-edit-block-modal">

	<el-radio-group v-model="styleType">
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

	<el-radio-group v-model="refType">
		<el-radio :label="null">No media</el-radio>
		<el-radio label="subspec">Subspec</el-radio>
		<el-radio label="url">URL</el-radio>
	</el-radio-group>

	<ref-url-form
		v-if="refType === REF_TYPE_URL"
		:spec-id="specId"
		:initial-url-object="existingUrlRefItem"
		:valid.sync="refFieldsValid"
		:fields.sync="refFields"
		@open-edit-url="openEditUrl"
		/>

	<ref-subspec-form
		v-else-if="refType === REF_TYPE_SUBSPEC"
		:spec-id="specId"
		:initial-subspec="existingSubspecRefItem"
		:valid.sync="refFieldsValid"
		:fields.sync="refFields"
		/>

	<span slot="footer" class="dialog-footer">
		<el-button @click="showing = false">Cancel</el-button>
		<el-button @click="submit()" type="primary" :disabled="disableSubmit">
			{{block ? 'Save' : 'Add'}}
		</el-button>
	</span>

</el-dialog>
</template>

<script>
import $ from 'jquery';
import RefUrlForm from './ref-url-form.vue';
import RefSubspecForm from './ref-subspec-form.vue';
import {
	ajaxCreateBlock, ajaxSaveBlock,
	ajaxLoadSubspecs, ajaxCreateSubspec,
} from './ajax.js';
import {
	REF_TYPE_URL, REF_TYPE_SUBSPEC,
} from './const.js';

export default {
	components: {
		RefUrlForm,
		RefSubspecForm,
	},
	props: {
		specId: {
			type: Number,
			required: true,
		},
		subspecId: Number,
	},
	data() {
		return {
			// user inputs
			styleType: 'bullet',
			title: '',
			body: '',
			refType: null,
			refFields: null,
			refFieldsValid: false,
			// passed in
			block: null,
			parentId: null,
			insertBeforeId: null,
			callback: null,
			// state
			showing: false,
			initialRefItem: null,
		};
	},
	computed: {
		REF_TYPE_URL() {
			return REF_TYPE_URL;
		},
		REF_TYPE_SUBSPEC() {
			return REF_TYPE_SUBSPEC;
		},
		disableSubmit() {
			if (this.refType) {
				return !(this.refFieldsValid && this.refFields);
			} else {
				return !(this.title.trim() || this.body.trim());
			}
		},
		existingUrlRefItem() {
			return this.block && this.block.refType === REF_TYPE_URL && this.block.refItem || null;
		},
		existingSubspecRefItem() {
			return this.block && this.block.refType === REF_TYPE_SUBSPEC && this.block.refItem || null;
		},
	},
	watch: {
		refType(type) {
			this.refFields = null;
		},
	},
	methods: {
		showAdd(parentId, insertBeforeId, callback) {
			this.parentId = parentId;
			this.insertBeforeId = insertBeforeId;
			this.callback = callback;
			this.showing = true;
			this.focusTitleInput();
		},
		showEdit(block, callback) {
			this.block = block; // existing state
			this.styleType = block.styleType;
			this.title = block.title || '';
			this.body = block.body || '';
			this.refType = block.refType;
			this.callback = callback;
			this.showing = true;
			this.focusTitleInput();
		},
		focusTitleInput() {
			this.$nextTick(() => {
				$('input', this.$refs.titleInput.$el).focus();
			});
		},
		openEditUrl(urlObject, updated = null, deleted = null) {
			this.$emit('open-edit-url', urlObject, updatedUrlObject => {
				// Updated
				if (this.existingUrlRefItem && updatedUrlObject.id === this.existingUrlRefItem.id) {
					// Update existing ref
					this.block.refItem = updatedUrlObject;
				}
				if (updated) {
					updated(updatedUrlObject);
				}
			}, deletedId => {
				// Deleted
				if (this.existingUrlRefItem && deletedId === this.existingUrlRefItem.id) {
					// Clear existing ref
					this.block.refType = null;
					this.block.refId = null;
					this.block.refItem = null;
				}
				if (deleted) {
					deleted(deletedId);
				}
			});
		},
		submit() {
			if (this.disableSubmit) {
				return;
			}
			let sending = this.createSendingSpinner();
			let callback = this.callback; // in case modal is closed before complete
			if (this.block) {
				ajaxSaveBlock(
					this.specId,
					this.block.id,
					this.styleType,
					'plaintext', // contentType
					this.title,
					this.body,
					this.refType,
					this.refFields,
				).then(updatedBlock => {
					callback(updatedBlock);
					this.showing = false;
					sending.close();
				}).fail(() => {
					sending.close();
				});
			} else {
				ajaxCreateBlock(
					this.specId,
					this.subspecId,
					this.parentId,
					this.insertBeforeId,
					this.styleType,
					'plaintext', // contentType
					this.title,
					this.body,
					this.refType,
					this.refFields,
				).then(newBlock => {
					callback(newBlock);
					this.showing = false;
					sending.close();
				}).fail(() => {
					sending.close();
				});
			}
		},
		createSendingSpinner() {
			return this.$loading({
				lock: true,
				background: 'rgba(0, 0, 0, 0.7)',
			});
		},
		closed() {
			this.block = null;
			this.parentId = null;
			this.insertBeforeId = null;
			this.callback = null;
			this.title = '';
			this.body = '';
			this.refType = null;
			this.refFields = null;
			this.refFieldsValid = false;
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
				>.el-input {
					display: block;
					width: 100%;
				}
			}
		}
	}
}
</style>
