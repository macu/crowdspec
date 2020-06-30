<template>
<section v-if="spec" class="spec-page">
	<header>
		<h2>
			<div class="right">
				<span v-if="currentUserOwns">You own this</span>
				<span v-else>Owned by {{spec.ownerType}} {{spec.ownerId}}</span>

				<span v-if="spec.public">Public</span>
				<span v-else>Not public</span>

				<el-button @click="openManageSpec()" size="mini" icon="el-icon-setting"/>
			</div>
			{{name}}
		</h2>
		<div v-if="desc" class="desc">{{desc}}</div>
	</header>

	<spec-view :spec="spec"/>

	<edit-spec-modal ref="editSpecModal"/>

</section>
</template>

<script>
import $ from 'jquery';
import SpecView from '../spec/view.vue';
import EditSpecModal from '../spec/edit-spec-modal.vue';
import {ajaxLoadSpec} from '../spec/ajax.js';

export default {
	components: {
		SpecView,
		EditSpecModal,
	},
	data() {
		return {
			spec: null,
			name: '',
			desc: '',
		};
	},
	computed: {
		currentUserOwns() {
			return this.spec.ownerType === 'user' &&
				this.$store.getters.userID === this.spec.ownerId;
		},
	},
	beforeRouteEnter(to, from, next) {
		ajaxLoadSpec(to.params.specId).then(spec => {
			next(vm => {
				vm.setSpec(spec);
			});
		}).fail(jqXHR => {
			next({name: 'ajax-error', params: {code: jqXHR.status}, replace: true});
		});
	},
	beforeRouteUpdate(to, from, next) {
		ajaxLoadSpec(to.params.specId).then(spec => {
			this.setSpec(spec);
			next();
		}).fail(jqXHR => {
			next({name: 'ajax-error', params: {code: jqXHR.status}, replace: true});
		});
	},
	methods: {
		setSpec(spec) {
			this.spec = spec;
			this.name = spec.name;
			this.desc = spec.desc;
		},
		openManageSpec() {
			this.$refs.editSpecModal.showEdit({
				id: this.spec.id,
				name: this.name,
				desc: this.desc,
				created: this.spec.created,
			}, updatedSpec => {
				this.name = updatedSpec.name;
				this.desc = updatedSpec.desc;
			});
		},
	},
};
</script>

<style lang="scss">
.spec-page {
	>header {
		margin-top: -1cm;
		margin-left: -1cm;
		margin-right: -1cm;
		margin-bottom: 1cm;
		padding: 0.5cm 1cm;
		background-color: darkblue;
		color: white;
		>h2 {
			margin: 0;
			>.right {
				float: right;
				font-size: small;
				margin-left: 20px;
				>*+* {
					margin-left: 10px;
				}
			}
		}
		>.desc {
			white-space: pre-wrap;
			margin-top: 10px;
		}
	}
}
</style>
