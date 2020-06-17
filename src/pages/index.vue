<template>
<div class="index-page">

	<el-button @click="gotoNewSpec()">New spec</el-button>

	<div class="user-specs">
		<h2>Your specs</h2>
		<p v-if="loading">Loading...</p>
		<ul v-else-if="userSpecs && userSpecs.length">
			<li v-for="s in userSpecs" :key="s.id">
				<router-link :to="{name: 'spec', params: {specId: s.id}}">{{s.name}}</router-link>
			</li>
		</ul>
		<p v-else>You do not have any specs.</p>
	</div>

</div>
</template>

<script>
import $ from 'jquery';
import {alertError} from '../utils.js';

export default {
	data() {
		return {
			userSpecs: [],
			loading: true,
		};
	},
	mounted() {
		this.reloadSpecs();
	},
	beforeRouteUpdate(to, from, next) {
		this.reloadSpecs();
		next();
	},
	methods: {
		reloadSpecs() {
			this.loading = true;
			$.get('/ajax/user-specs').then(specs => {
				this.userSpecs = specs;
				this.loading = false;
			}).fail(jqXHR => {
				this.loading = false;
				alertError(jqXHR);
			});
		},
		gotoNewSpec() {
			this.$router.push({name: 'new-spec'});
		},
	},
};
</script>

<style lang="scss">
.index-page {
	.user-specs, .public-specs {
		margin-top: 20px;
	}
}
</style>
