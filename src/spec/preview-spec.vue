<template>
<div class="spec-preview community-target-preview">
	<div class="name">
		{{spec.name}}
		<el-tag v-if="!spec.public" type="info" effect="dark">Private</el-tag>
	</div>
	<div v-if="spec.desc" class="desc">{{spec.desc}}</div>
</div>
</template>

<script>
export default {
	props: {
		spec: Object,
	},
};
</script>

<style lang="scss">
@import '../_styles/_colours.scss';

.spec-preview {
	border: medium solid $spec-bg;
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
