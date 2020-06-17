<template>
<div v-if="spec" class="spec-page">

	<h2>{{spec.name}}</h2>
	<div v-if="spec.desc" class="desc">{{spec.desc}}</div>
	<p>Owner: {{spec.ownerType}} {{spec.ownerId}}</p>
	<p>Created: {{spec.created}}</p>
	<p v-if="spec.public">Public</p>
	<p v-else>Not public</p>

	<hr/>

	<spec-view :spec="spec"/>

</div>
</template>

<script>
import $ from 'jquery';
import SpecView from '../spec/view.vue';
import {ajaxLoadSpec} from '../spec/ajax.js';

export default {
	components: {
		SpecView,
	},
	data() {
		return {
			spec: null,
		};
	},
	beforeRouteEnter(to, from, next) {
		ajaxLoadSpec(to.params.specId).then(spec => {
			next(vm => {
				vm.spec = spec;
			});
		}).fail(jqXHR => {
			next({name: 'ajax-error', params: {code: jqXHR.status}, replace: true});
		});
	},
	beforeRouteUpdate(to, from, next) {
		ajaxLoadSpec(to.params.specId).then(spec => {
			this.spec = spec;
			next();
		}).fail(jqXHR => {
			next({name: 'ajax-error', params: {code: jqXHR.status}, replace: true});
		});
	},
	methods: {
	},
};
</script>

<style lang="scss">
.spec-page {
	>h2 {
		margin-top: 0;
	}
	>.desc {
		white-space: pre-wrap;
		margin-bottom: 10px;
	}
}
</style>
