<template>
<div class="spec-page">
	<h2>{{spec.name}}</h2>
	<div v-if="spec.desc" class="desc">{{spec.desc}}</div>
	<p>Owner: {{spec.ownerName}}</p>
	<p>Created: {{spec.created}}</p>
	<p v-if="spec.public">Public</p>
	<p v-else>Not public</p>
	<p v-if="spec.userIsAdmin">You are admin</p>
	<p v-if="spec.userIsContributor">You are a contributor</p>
</div>
</template>

<script>
import $ from 'jquery';
import {alertError} from '../utils.js';

function ajaxLoadSpec(specId) {
	return $.get('/ajax/spec', {specId}).fail(alertError);
}

export default {
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
