<template>
<el-dialog
	:title="urlObject ? 'Manage link' : 'Create link'"
	:visible.sync="showing"
	:width="$store.getters.dialogSmallWidth"
	:close-on-click-modal="false"
	@closed="closed()"
	class="spec-edit-url-modal">

	<p v-if="urlObject">
		Created <strong><moment :datetime="urlObject.created"/></strong>;
		last modified <strong><moment :datetime="urlObject.updated" :offset="true"/></strong>
	</p>

	<label>
		URL
		<el-input ref="urlInput" v-model="url" :maxlength="urlMaxLength"/>
	</label>

	<template v-if="urlObject">
		<div v-if="url === urlObject.url" class="preview">
			<img v-if="urlObject.imageData" :src="urlObject.imageData"/>
			<div>
				<h3 v-if="urlObject.title">{{urlObject.title.trim()}}</h3>
				<div v-if="urlObject.desc">{{urlObject.desc.trim()}}</div>
				<p v-if="autoRefresh">URL preview will be updated</p>
				<el-checkbox v-else v-model="refresh">Refresh URL preview</el-checkbox>
			</div>
		</div>
		<p v-else>Link will be updated with new URL</p>
	</template>

	<span slot="footer" class="dialog-footer">
		<el-button @click="showing = false">Cancel</el-button>
		<el-button v-if="urlObject" @click="promptDelete()" type="danger">Delete</el-button>
		<el-button @click="submit()" type="primary" :disabled="disableSubmit">{{urlObject ? 'Save' : 'Create'}}</el-button>
	</span>

</el-dialog>
</template>

<script>
import $ from 'jquery';
import Moment from '../widgets/moment.vue';
import {ajaxCreateUrl, ajaxRefreshUrl, ajaxDeleteUrl} from './ajax.js';
import {isValidURL} from '../utils.js';

export default {
	components: {
		Moment,
	},
	props: {
		specId: {
			type: Number,
			required: true,
		},
	},
	data() {
		return {
			// user inputs
			url: '',
			refresh: false,
			// passed in
			urlObject: null,
			createdCallback: null,
			updatedCallback: null,
			deletedCallback: null,
			// state
			showing: false,
			sending: false,
		};
	},
	computed: {
		disableSubmit() {
			window.isValidURL = isValidURL;
			return !isValidURL(this.url) || (this.urlObject && !(this.refresh || this.autoRefresh));
		},
		urlMaxLength() {
			return window.const.urlMaxLength;
		},
		autoRefresh() {
			return this.urlObject && this.url !== this.urlObject.url;
		},
	},
	watch: {
		sending(sending) {
			if (sending) {
				this.sendingSpinner = this.$loading({
					lock: true,
					background: 'rgba(0, 0, 0, 0.7)',
				});
			} else if (this.sendingSpinner) {
				this.sendingSpinner.close();
				this.sendingSpinner = null;
			}
		},
	},
	beforeDestroy() {
		if (this.sendingSpinner) {
			this.sendingSpinner.close();
			this.sendingSpinner = null;
		}
	},
	methods: {
		showCreate(callback) {
			this.createdCallback = callback;
			this.showing = true;
			this.$nextTick(() => {
				$('input', this.$refs.urlInput.$el).focus();
			});
		},
		showEdit(urlObject, updated, deleted) {
			this.urlObject = urlObject;
			this.url = urlObject.url;
			this.refresh = true;
			this.updatedCallback = updated;
			this.deletedCallback = deleted;
			this.showing = true;
			this.$nextTick(() => {
				$('input', this.$refs.urlInput.$el).focus();
			});
		},
		submit() {
			if (this.disableSubmit) {
				return;
			}
			if (this.urlObject) {
				this.submitSave();
			} else {
				this.submitCreate();
			}
		},
		submitCreate() {
			if (this.disableSubmit) {
				return;
			}
			this.sending = true;
			let callback = this.createdCallback; // in case modal is closed before complete
			ajaxCreateUrl(
				this.specId,
				this.url
			).then(newUrl => {
				callback(newUrl);
				this.sending = false;
				this.showing = false;
			}).fail(() => {
				this.sending = false;
			})
		},
		submitSave() {
			if (this.disableSubmit) {
				return;
			}
			this.sending = true;
			let callback = this.updatedCallback; // in case modal is closed before complete
			ajaxRefreshUrl(
				this.urlObject.id,
				this.url
			).then(updatedUrlItem => {
				callback(updatedUrlItem);
				this.sending = false;
				this.showing = false;
			}).fail(() => {
				this.sending = false;
			});
		},
		promptDelete() {
			this.$confirm('Permanently delete this link? All references to this link within this spec will be cleared.', {
				confirmButtonText: 'Delete',
				cancelButtonText: 'Cancel',
				type: 'warning',
			}).then(() => {
				this.sending = true;
				let callback = this.deletedCallback; // in case modal is closed before complete
				ajaxDeleteUrl(this.urlObject.id).then(() => {
					if (callback) {
						callback(this.urlObject.id);
					}
					this.sending = false;
					this.showing = false;
				}).fail(() => {
					this.sending = false;
				});
			}).catch(() => {
				// Cancelled
			});
		},
		closed() {
			this.url = '';
			this.refresh = false;
			this.urlObject = null;
			this.createdCallback = null;
			this.updatedCallback = null;
			this.deletedCallback = null;
		},
	},
};
</script>

<style lang="scss">
.spec-edit-url-modal {
	>.el-dialog {
		>.el-dialog__body {
			>p {
				margin-top: 0;
			}
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
			>.preview {
				display: flex;
				flex-direction: row;
				>img {
					max-width: 20%;
					margin-right: 20px;
				}
				>div {
					flex: 1;
					>* {
						margin: 10px 0;
					}
					>h3 {
						margin: 0 0 10px;
					}
				}
			}
		}
	}
}
</style>
