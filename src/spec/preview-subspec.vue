<template>
<div class="spec-subspec-preview community-target-preview">
	<div class="name">
		{{subspec.name}}
		<el-tag v-if="subspec.private" type="info" effect="dark">Private</el-tag>
	</div>
	<div v-if="subspec.desc" class="desc">{{subspec.desc}}</div>
</div>
</template>

<script>
export default {
	props: {
		subspec: Object,
	},
};
</script>

<style lang="scss">
@import '../_styles/_colours.scss';

.spec-subspec-preview {
	border: medium solid $subspec-bg;
	border-radius: 8px;
	background-color: $shadow-bg;
	padding: 20px;

	>.name {
		font-weight: bold;
		font-size: larger;
		>.el-tag {
			margin-left: 10px;
		}
	}

	>.desc {
		white-space: pre-wrap;
	}

	>* ~ * {
		margin-top: 10px;
	}
}
</style>
