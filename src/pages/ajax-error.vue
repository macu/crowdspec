<template>
<div class="ajax-error-page content-page">

	<p v-if="statusCode === 0">Could not connect to server</p>
	<p v-else-if="statusCode === 503">CrowdSpec is currently offline for database upgrades</p>
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
		statusCode() {
			return parseInt(this.$route.params.code, 10) || 0;
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
