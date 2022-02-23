<template>
<el-dialog
	:title="subspec ? 'Manage subspec' : 'Create subspec'"
	v-model="showing"
	:width="$store.getters.dialogSmallWidth"
	:close-on-click-modal="false"
	@closed="closed()"
	custom-class="spec-edit-subspec-modal">

	<p v-if="subspec">
		Created <strong><moment :datetime="subspec.created"/></strong>;
		last modified <strong><moment :datetime="subspec.updated" :offset="true"/></strong>
	</p>

	<label>
		<div>Name</div>
		<el-input ref="nameInput" v-model="name" :maxlength="nameMaxLength" clearable/>
	</label>

	<label>
		<div>Description</div>
		<el-input type="textarea" v-model="desc" :autosize="{minRows: 2}"/>
	</label>

	<template #footer>
		<span class="dialog-footer">
			<el-button @click="showing = false">Cancel</el-button>
			<el-button v-if="subspec" @click="promptDeleteSubspec()" type="danger">Delete</el-button>
			<el-button @click="submit()" type="primary" :disabled="disableSubmit">{{subspec ? 'Save' : 'Create'}}</el-button>
		</span>
	</template>

</el-dialog>
</template>

<script>
import $ from 'jquery';
import Moment from '../widgets/moment.vue';
import {ajaxCreateSubspec, ajaxSaveSubspec, ajaxDeleteSubspec} from './ajax.js';

export default {
	components: {
		Moment,
	},
	props: {
		specId: Number,
	},
	data() {
		return {
			// user inputs
			name: '',
			desc: '',
			// passed in
			subspec: null,
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
		nameMaxLength() {
			return window.const.specNameMaxLength;
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
			this.callback = callback;
			this.showing = true;
			this.$nextTick(() => {
				$('input', this.$refs.nameInput.$el).focus();
			});
		},
		showEdit(subspec, callback) {
			this.subspec = subspec;
			this.name = subspec.name;
			this.desc = subspec.desc;
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
			if (this.subspec) {
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
			ajaxCreateSubspec(
				this.specId,
				this.name,
				this.desc
			).then(newSubspec => {
				callback(newSubspec.id);
				this.showing = false;
				this.sending = false;
			}).fail(() => {
				this.sending = false;
			});
		},
		submitSave() {
			this.sending = true;
			let callback = this.callback; // in case modal is closed before complete
			ajaxSaveSubspec(
				this.subspec.id,
				this.name,
				this.desc
			).then(updatedSubspec => {
				callback(updatedSubspec);
				this.showing = false;
				this.sending = false;
			}).fail(() => {
				this.sending = false;
			});
		},
		promptDeleteSubspec() {
			this.$confirm('Permanently delete this subspec?', {
				confirmButtonText: 'Delete',
				cancelButtonText: 'Cancel',
				type: 'warning',
			}).then(() => {
				this.sending = true;
				ajaxDeleteSubspec(this.subspec.id).then(() => {
					this.sending = false;
					this.showing = false;
					this.$nextTick(() => {
						this.$router.push({name: 'spec', params: {specId: this.specId}});
					});
				}).fail(() => {
					this.sending = false;
				});
			}).catch(() => {
				// Cancelled
			});
		},
		closed() {
			this.subspec = null;
			this.name = '';
			this.desc = '';
		},
	},
};
</script>

<style lang="scss">
.spec-edit-subspec-modal.el-dialog {
	>.el-dialog__body {
		>p {
			margin-top: 0;
		}
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
</style>
