<template>
<section v-if="spec" class="spec-page">

	<header>
		<div class="right">
			<span v-if="currentUserOwns">You own this</span>
			<span v-else>Owned by {{spec.ownerType}} {{spec.ownerId}}</span>

			<span v-if="spec.public">
				Public
			</span>
			<span v-else>
				<el-tooltip content="Unpublished" placement="left">
					<i class="el-icon-lock"></i>
				</el-tooltip>
			</span>

			<el-button @click="openManageSpec()" size="mini" icon="el-icon-setting"/>
		</div>
		<h2>{{name}}</h2>
		<div v-if="desc" class="desc">{{desc}}</div>
	</header>

	<spec-view :key="spec.id" :spec="spec"/>

	<edit-spec-modal ref="editSpecModal"/>

</section>
</template>

<script>
import $ from 'jquery';
import SpecView from '../spec/view.vue';
import EditSpecModal from '../spec/edit-spec-modal.vue';
import {ajaxLoadSpec} from '../spec/ajax.js';
import {setWindowSubtitle} from '../utils.js';

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
	beforeRouteLeave(to, from, next) {
		setWindowSubtitle(); // clear
		next();
	},
	methods: {
		setSpec(spec) {
			this.spec = spec;
			this.name = spec.name;
			this.desc = spec.desc;
			setWindowSubtitle(spec.name);
			// vue-router scrollBehavior is applied before spec-view has a chance to populate,
			// so restore the scroll position again after fully rendering.
			this.$nextTick(this.restoreScroll);
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
				setWindowSubtitle(updatedSpec.name);
			});
		},
		restoreScroll() {
			let position = this.$store.state.savedScrollPosition;
			if (position) {
				$(window).scrollTop(position.y).scrollLeft(position.x);
			}
		},
	},
};
</script>

<style lang="scss">
@import '../styles/_breakpoints.scss';
@import '../styles/_colours.scss';
@import '../styles/_spec-view.scss';
@import '../styles/_app.scss';

.spec-page {

	>header {
		margin-bottom: 1cm;

		background-color: $spec-bg;
		color: white;

		padding: 0.5cm $page-header-horiz-padding;
		@include mobile {
			padding: 15px $page-header-horiz-padding-sm 15px;
		}

		>.right {
			float: right;
			font-size: small;
			margin-left: 20px;

			>*+* {
				margin-left: 15px;
			}
		}

		>h2 {
			margin: 0;
		}

		>.desc {
			white-space: pre-wrap;
			margin-top: 10px;
			color: white;
		}
	} // header

	@include mobile {
		>.spec-view {
			margin-left: #{-$content-area-padding-sm};
			margin-right: calc(-1 * (#{$content-area-padding-sm} - #{$spec-block-bg-padding-horiz}));
		}
	}
}
</style>
