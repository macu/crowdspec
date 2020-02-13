<template>
<el-dialog
	title="Add point"
	:visible.sync="showing"
	:width="$store.getters.dialogTinyWidth"
	@closed="closed()"
	class="spec-add-point-modal">

	<label>
		Title
		<el-input ref="titleInput" v-model="title" clearable/>
	</label>

	<label>
		Description
		<el-input type="textarea" v-model="desc" :autosize="{minRows: 2}"/>
	</label>

	<span slot="footer" class="dialog-footer">
		<el-button @click="showing = false">Cancel</el-button>
		<el-button @click="submit()" type="primary" :disabled="disableSubmit">Add</el-button>
	</span>

</el-dialog>
</template>

<script>
import $ from 'jquery';

export default {
	data() {
		return {
			showing: false,
			callback: null,
			title: '',
			desc: '',
		};
	},
	computed: {
		disableSubmit() {
			return !(this.title.trim() || this.desc.trim());
		},
	},
	methods: {
		show(callback) {
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
			this.callback(this.title, this.desc, () => {
				this.showing = false;
			});
		},
		closed() {
			this.callback = null;
			this.title = '';
			this.desc = '';
		},
	},
};
</script>

<style lang="scss">
.spec-add-point-modal {
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
