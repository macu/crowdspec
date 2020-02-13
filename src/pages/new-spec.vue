<template>
<div class="new-spec-page">
	<label>
		Name:
		<el-input ref="nameInput" v-model="newSpecName" clearable/>
	</label>
	<label>
		Description:
		<el-input type="textarea" v-model="newSpecDesc" :autosize="{minRows: 2}"/>
	</label>
	<el-button @click="cancel()">Cancel</el-button>
	<el-button @click="create()" :disabled="disableCreate" type="primary">Create</el-button>
</div>
</template>

<script>
import $ from 'jquery';
import {ajaxCreateSpec} from '../spec/ajax.js';

export default {
	data() {
		return {
			newSpecName: '',
			newSpecDesc: '',
		};
	},
	computed: {
		disableCreate() {
			return !this.newSpecName.trim();
		},
	},
	beforeRouteEnter (to, from, next) {
		next(vm => vm.nextTickFocusNameInput());
	},
	beforeRouteUpdate(to, from, next) {
		this.newSpecName = '';
		this.newSpecDesc = '';
		this.nextTickFocusNameInput();
		next();
	},
	methods: {
		nextTickFocusNameInput() {
			this.$nextTick(() => {
				$('input', this.$refs.nameInput.$el).focus();
			});
		},
		cancel() {
			this.$router.push({name: 'index'});
		},
		create() {
			if (this.disableCreate) {
				return;
			}
			ajaxCreateSpec(this.newSpecName, this.newSpecDesc).then(specId => {
				this.$router.push({name: 'spec', params: {specId}});
			});
		},
	},
};
</script>

<style lang="scss">
.new-spec-page {
	>label {
		display: block;
		margin-bottom: 20px;
		>input {
			display: block;
			width: 100%;
		}
	}
}
</style>
