<template>
<div class="ajax-error-page content-page">

	<p v-if="statusCode === 0">Could not connect to server</p>
	<p v-else-if="statusCode === 503">CrowdSpec is currently offline for database upgrades</p>
	<p v-else-if="specForbidden">You do not have access to the requested spec.</p>
	<p v-else-if="subspecForbidden">You do not have access to the requested subspec.</p>
	<p v-else>Request failed with error code {{statusCode}}</p>

	<el-button v-if="url" @click="retry()">Retry</el-button>

</div>
</template>

<script>
export default {
	computed: {
		url() {
			return decodeURIComponent(this.$route.query.url);
		},
		errorType() {
			return this.$route.query.e || false;
		},
		statusCode() {
			return parseInt(this.$route.params.code, 10) || 0;
		},
		specForbidden() {
			return this.statusCode === 403 && this.errorType === 'spec';
		},
		subspecForbidden() {
			return this.statusCode === 403 && this.errorType === 'subspec';
		},
	},
	methods: {
		retry() {
			this.$router.replace(this.url);
		},
	},
};
</script>

<style lang="scss">
.ajax-error-page {
	>.el-button {
		margin-top: 40px;
	}
}
</style>
