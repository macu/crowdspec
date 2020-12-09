<template>
<div class="ajax-error-page content-page">

	<p v-if="parseInt($route.params.code, 10) === 0">
		Could not connect to server
	</p>
	<p v-else>
		Request failed with error code {{$route.params.code}}
	</p>

	<el-button v-if="url" @click="retry()">Retry</el-button>

</div>
</template>

<script>
export default {
	computed: {
		url() {
			return decodeURIComponent(this.$route.query.url);
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
