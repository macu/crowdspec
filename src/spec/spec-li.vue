<template>
<li class="spec-li">
	<div class="title">{{title}}</div>
	<div v-if="desc" class="desc">{{desc}}</div>
	<ul class="subpoints">
		<template v-if="subpoints.length">
			<spec-li
				v-for="p in subpoints"
				:key="p.id"
				:initial-point="p"
				@add-subpoint="raiseAddSubpoint"
				/>
			</template>
		<li><el-button @click="addSubpoint()" size="mini">Add subpoint</el-button></li>
	</ul>
</li>
</template>

<script>
export default {
	name: 'spec-li',
	props: {
		initialPoint: Object,
	},
	data() {
		return {
			id: this.initialPoint.id,
			title: this.initialPoint.title,
			desc: this.initialPoint.desc,
			subpoints: this.initialPoint.points || [],
		};
	},
	methods: {
		addSubpoint() {
			this.raiseAddSubpoint(this.id, this.subpoints.length, newSubpoint => {
				this.subpoints.push(newSubpoint);
			});
		},
		raiseAddSubpoint(parentId, orderNumber, callback) {
			this.$emit('add-subpoint', parentId, orderNumber, callback);
		},
	},
};
</script>

<style lang="scss">
.spec-li {
	margin-bottom: 5px;
	&:last-child {
		margin-bottom: 0;
	}
	>.title {
		font-weight: bold;
		margin-bottom: 5px;
	}
	>.desc {
		white-space: pre-wrap;
		margin-bottom: 5px;
	}
	>:last-child {
		margin-bottom: 0;
	}
}
</style>
