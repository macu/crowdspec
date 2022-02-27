<template>
<div class="policy-disclaimer">
	<el-alert v-if="expanded" type="info" show-icon :closable="false">
		<p>No illegal or inappropriate content is allowed on this platform.</p>
		<p>All specs and community spaces, public and private, may be manually reviewed for legality and content appropriateness.</p>
		<p>Your specs and comments may be deleted and your account may be removed for posting illegal or unsuitable content.</p>
	</el-alert>
	<el-button v-else size="small" type="info" @click.prevent="expanded = true">
		View policy disclaimer
	</el-button>
</div>
</template>

<script>
export default {
	data() {
		return {
			expanded: false,
		};
	},
};
</script>

<style lang="scss">
.policy-disclaimer {
	>.el-alert {
		p {
			margin: 0 0 5px;
			&:last-child {
				margin: 0;
			}
		}
	}
	>.el-button {
		margin: 0;
	}
}
</style>
