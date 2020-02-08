<template>
<div class="new-spec-page">
	<label>
		Name:
		<el-input v-model="newSpecName" clearable/>
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
import {alertError} from '../utils.js';

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
	beforeRouteUpdate(to, from, next) {
		this.newSpecName = '';
		this.newSpecDesc = '';
		next();
	},
	methods: {
		cancel() {
			this.$router.push({name: 'index'});
		},
		create() {
			if (this.disableCreate) {
				return;
			}
			$.post('/ajax/create-spec', {
				name: this.newSpecName,
				desc: this.newSpecDesc,
			}).then(specId => {
				this.$router.push({name: 'spec', params: {specId}});
			}).fail(alertError);
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
