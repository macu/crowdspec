<template>
<div class="ref-url">
	<img v-if="item.imageData" :src="item.imageData"/>
	<div>
		<div v-if="showEdit" class="right">
			<el-button @click="$emit('edit')" size="mini" icon="el-icon-edit" circle/>
		</div>
		<div class="title">
			<span><i class="el-icon-link"></i> URL</span>
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
		showEdit: Boolean,
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
		>.right {
			float: right;
			margin-left: 10px;
			>.el-button {
				padding: 3px;
				font-size: 12px;
			}
		}
		>.title {
			font-weight: bold;
			>span {
				font-weight: normal;
				color: gray;
			}
		}
		>.desc {
			margin-top: 5px;
		}
	}
}
</style>
