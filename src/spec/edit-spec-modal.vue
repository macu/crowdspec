<template>
<el-dialog
	:title="spec ? 'Manage spec' : 'Create spec'"
	:visible.sync="showing"
	:width="$store.getters.dialogSmallWidth"
	:close-on-click-modal="false"
	@closed="closed()"
	class="spec-edit-spec-modal">

	<label>
		Name
		<el-input ref="nameInput" v-model="name" clearable/>
	</label>

	<label>
		Description
		<el-input type="textarea" v-model="desc" :autosize="{minRows: 2}"/>
	</label>

	<p v-if="spec">Created {{spec.created}}</p>

	<span slot="footer" class="dialog-footer">
		<el-button @click="showing = false">Cancel</el-button>
		<el-button v-if="spec" @click="promptDeleteSpec()" type="danger">Delete</el-button>
		<el-button @click="submit()" type="primary" :disabled="disableSubmit">{{spec ? 'Save' : 'Create'}}</el-button>
	</span>

</el-dialog>
</template>

<script>
import $ from 'jquery';
import {ajaxCreateSpec, ajaxSaveSpec, ajaxDeleteSpec} from './ajax.js';

export default {
	data() {
		return {
			// user inputs
			name: '',
			desc: '',
			// passed in
			spec: null,
			callback: null,
			// state
			showing: false,
			sending: false,
		};
	},
	computed: {
		disableSubmit() {
			return !this.name.trim();
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
	methods: {
		showCreate(callback) {
			this.callback = callback;
			this.showing = true;
			this.$nextTick(() => {
				$('input', this.$refs.nameInput.$el).focus();
			});
		},
		showEdit(spec, callback) {
			this.spec = spec;
			this.name = spec.name;
			this.desc = spec.desc;
			this.callback = callback;
			this.showing = true;
			this.$nextTick(() => {
				$('input', this.$refs.nameInput.$el).focus();
			});
		},
		submit() {
			if (this.disableSubmit) {
				return;
			}
			if (this.spec) {
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
			let callback = this.callback; // in case modal is closed before complete
			ajaxCreateSpec(
				this.name,
				this.desc
			).then(newSpecId => {
				callback(newSpecId);
				this.showing = false;
				this.sending = false;
			}).fail(() => {
				this.sending = false;
			});
		},
		submitSave() {
			this.sending = true;
			let callback = this.callback; // in case modal is closed before complete
			ajaxSaveSpec(
				this.spec.id,
				this.name,
				this.desc
			).then(updatedSpec => {
				callback(updatedSpec);
				this.showing = false;
				this.sending = false;
			}).fail(() => {
				this.sending = false;
			});
		},
		promptDeleteSpec() {
			this.$confirm('Permanently delete this spec?', {
				confirmButtonText: 'Delete',
				cancelButtonText: 'Cancel',
				type: 'warning',
			}).then(() => {
				this.sending = true;
				ajaxDeleteSpec(this.spec.id).then(() => {
					this.sending = false;
					this.showing = false;
					this.$nextTick(() => {
						this.$router.push({name: 'index'});
					});
				}).fail(() => {
					this.sending = false;
				});
			}).catch(() => {
				// Cancelled
			});
		},
		closed() {
			this.spec = null;
			this.name = '';
			this.desc = '';
		},
	},
};
</script>

<style lang="scss">
.spec-edit-spec-modal {
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
