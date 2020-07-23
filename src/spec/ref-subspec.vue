<template>
<div class="ref-subspec">
	<div>
		<div class="title">
			<span><i class="el-icon-s-unfold"></i> Subspec</span>
			<router-link :to="{name: 'subspec', params: {specId: item.specId, subspecId: item.id}}">
				{{item.name.trim()}}
			</router-link>
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
.ref-subspec {
	border: thin solid lightgray;
	border-radius: 5px;
	padding: 10px;
	display: flex; // fixes overlap with block layover
	>div {
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
