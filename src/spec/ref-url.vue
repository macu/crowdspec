<template>
<div class="ref-url">
	<img v-if="item.imageData" :src="item.imageData"/>
	<div>
		<div class="title">
			<i class="el-icon-link"></i>
			<a :href="item.url">{{(item.title && item.title.trim()) || item.url}}</a>
		</div>
		<div v-if="item.desc" class="desc">{{item.desc.trim()}}</div>
	</div>
</div>
</template>

<script>
export default {
	props: {
		item: Object,
	},
};
</script>

<style lang="scss">
.ref-url {
	border: thin solid lightgray;
	border-radius: 5px;
	padding: 10px;
	display: flex;
	align-items: flex-start;
	flex-direction: row;
	>img {
		width: 50px;
		margin-right: 10px;
	}
	>div {
		flex: 1;
		>.title {
			font-weight: bold;
		}
		>.desc {
			margin-top: 5px;
		}
	}
}
</style>
