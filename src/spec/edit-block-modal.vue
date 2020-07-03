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
		<el-radio label="url">URL</el-radio>
	</el-radio-group>

	<div v-if="refType === 'url'" class="ref-url-area">
		<el-input v-model="url">
			<template slot="prepend">URL</template>
		</el-input>
		<template v-if="existingUrlRefItem">
			<ref-url v-if="existingUrlRefItem.url === url" :item="existingUrlRefItem"/>
			<el-checkbox v-if="showRefreshUrl" v-model="refreshUrl">
				Refresh URL preview
			</el-checkbox>
			<p v-else-if="refreshUrl">URL preview will be updated.</p>
		</template>
	</div>

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
import RefUrl from './ref-url.vue';
import {ajaxCreateBlock, ajaxSaveBlock} from './ajax.js';
import {isValidURL} from '../utils.js';

export default {
	components: {
		RefUrl,
	},
	props: {
		specId: Number,
	},
	data() {
		return {
			// user inputs
			styleType: 'bullet',
			title: '',
			body: '',
			refType: null,
			url: '',
			refreshUrl: false,
			// passed in
			block: null,
			subspaceId: null,
			parentId: null,
			insertBeforeId: null,
			callback: null,
			// state
			showing: false,
		};
	},
	computed: {
		disableSubmit() {
			if (this.refType) {
				switch (this.refType) {
					case 'url':
						return !isValidURL(this.url);
						break;
					default:
						// Unrecognized; don't allow submit
						return true;
				}
			} else {
				return !(this.title.trim() || this.body.trim());
			}
		},
		existingUrlRefItem() {
			return this.block && this.block.refType === 'url' && this.block.refItem;
		},
		showRefreshUrl() {
			// Show the refresh option if not changing the URL; otherwise refresh is automatic
			return this.existingUrlRefItem && this.existingUrlRefItem.url === this.url;
		},
	},
	watch: {
		url(url) {
			if (this.existingUrlRefItem) {
				// Set refresh to whether the URL has been changed
				this.refreshUrl = this.existingUrlRefItem.url !== this.url;
			}
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
			this.refType = block.refType;
			this.callback = callback;

			// Extract refItem data
			if (block.refType === 'url' && block.refItem) {
				this.url = block.refItem.url;
			}

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
			let sending = this.createSendingSpinner();
			let callback = this.callback; // in case modal is closed before complete
			let refFields = null;
			if (this.refType === 'url' && isValidURL(this.url)) {
				refFields = {
					refType: 'url',
					refUrl: this.url,
				};
			}
			ajaxCreateBlock(
				this.specId,
				this.subspaceId,
				this.parentId,
				this.insertBeforeId,
				this.styleType,
				'plaintext', // contentType
				this.title,
				this.body,
				refFields,
			).then(newBlock => {
				callback(newBlock);
				this.showing = false;
				sending.close();
			}).fail(() => {
				sending.close();
			});
		},
		submitSave() {
			let sending = this.createSendingSpinner();
			let callback = this.callback; // in case modal is closed before complete
			let refFields = null;
			if (this.refType === 'url') {
				refFields = {
					refType: 'url',
					refId: this.existingUrlRefItem ? this.existingUrlRefItem.id : null,
					refUrl: this.url,
					refRefreshUrl: this.refreshUrl,
				};
			}
			ajaxSaveBlock(
				this.specId,
				this.block.id,
				this.styleType,
				'plaintext', // contentType
				this.title,
				this.body,
				refFields,
			).then(updatedBlock => {
				callback(updatedBlock);
				this.showing = false;
				sending.close();
			}).fail(() => {
				sending.close();
			});
		},
		createSendingSpinner() {
			return this.$loading({
				lock: true,
				background: 'rgba(0, 0, 0, 0.7)',
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
			this.refType = null;
			this.url = '';
			this.refreshUrl = false;
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
			>.ref-url-area {
				margin-top: 20px;
				>* + * {
					margin-top: 10px;
				}
			}
		}
	}
}
</style>
