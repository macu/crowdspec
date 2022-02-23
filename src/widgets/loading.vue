<template>
<span class="loading-message">
	<i class="material-icons spin">loop</i>
	<span v-if="message">{{message}}</span>
</span>
</template>

<script>
export default {
	props: {
		message: String,
	},
};
</script>

<style lang="scss">
.loading-message {
	>i {
		display: inline-block;
	}
	>span {
		display: inline-block;
		margin-left: 1ex;
	}
}
</style>
