<template>
<div class="spec-points">
	<ul class="spec-ul">
		<spec-li
			v-for="p in points"
			:key="p.id"
			:initial-point="p"
			@add-subpoint="addSubpoint"
			/>
		<li><el-button @click="addRootPoint()" size="mini">Add root point</el-button></li>
	</ul>
	<add-point-modal ref="addPointModal"/>
</div>
</template>

<script>
import AddPointModal from './add-point-modal.vue';
import SpecLi from './spec-li.vue';
import {ajaxCreateSubpoint} from './ajax.js';

export default {
	components: {
		AddPointModal,
		SpecLi,
	},
	props: {
		specId: Number,
		initialPoints: Array,
	},
	data() {
		return {
			points: this.initialPoints ? this.initialPoints.slice() : [],
		};
	},
	methods: {
		addRootPoint() {
			this.$refs.addPointModal.show((title, desc, closeModal) => {
				ajaxCreateSubpoint(
					this.specId, 0, this.points.length, title, desc
				).then(newPoint => {
					this.points.push(newPoint);
					closeModal();
				});
			});
		},
		addSubpoint(parentId, orderNumber, callback) {
			this.$refs.addPointModal.show((title, desc, closeModal) => {
				ajaxCreateSubpoint(
					this.specId, parentId, orderNumber, title, desc
				).then(newPoint => {
					callback(newPoint);
					closeModal();
				});
			});
		},
	},
};
</script>
