<template>
<section v-if="spec" class="spec-page">

	<header>

		<div class="right">

			<span v-if="currentUserOwns">You own this</span>
			<span v-else>
				Owned by
				<template v-if="spec.username">{{spec.username}}</template>
				<template v-else>{{spec.ownerType}} {{spec.ownerId}}</template>
			</span>

			<template v-if="enableEditing">
				<span v-if="spec.public">
					Public
				</span>
				<span v-else>
					<el-tooltip content="Unpublished" placement="left">
						<i class="el-icon-lock"></i>
					</el-tooltip>
				</span>
				<el-button @click="openManageSpec()" size="mini" icon="el-icon-setting"/>
			</template>

			<span v-else>
				Last modified <moment :datetime="spec.updated" :offset="true"/>
			</span>

		</div>

		<h2>{{spec.name}}</h2>

		<div v-if="spec.desc" class="desc">{{spec.desc}}</div>

	</header>

	<spec-view
		:key="spec.id"
		:spec="spec"
		:enable-editing="enableEditing"
		/>

	<edit-spec-modal
		v-if="enableEditing"
		ref="editSpecModal"
		/>

</section>
</template>

<script>
import $ from 'jquery';
import Moment from '../widgets/moment.vue';
import SpecView from '../spec/view.vue';
import EditSpecModal from '../spec/edit-spec-modal.vue';
import {ajaxLoadSpec} from '../spec/ajax.js';
import {OWNER_TYPE_USER} from '../spec/const.js';
import {setWindowSubtitle, idsEq} from '../utils.js';

export default {
	components: {
		Moment,
		SpecView,
		EditSpecModal,
	},
	data() {
		return {
			spec: null,
		};
	},
	computed: {
		currentUserOwns() {
			return this.spec.ownerType === OWNER_TYPE_USER &&
				this.$store.getters.userID === this.spec.ownerId;
		},
		enableEditing() {
			// Currently users may edit only their own specs
			return this.currentUserOwns;
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
			setWindowSubtitle(spec.name);
			// vue-router scrollBehavior is applied before spec-view has a chance to populate,
			// so restore the scroll position again after fully rendering.
			this.$nextTick(this.restoreScroll);
		},
		openManageSpec() {
			this.$refs.editSpecModal.showEdit(this.spec, updatedSpec => {
				// Update properties provided back by ajaxSaveSpec
				this.spec.updated = updatedSpec.updated;
				this.spec.name = updatedSpec.name;
				this.spec.desc = updatedSpec.desc;
				this.spec.public = updatedSpec.public;
				setWindowSubtitle(updatedSpec.name);
			});
		},
		restoreScroll() {
			let position = this.$store.state.savedScrollPosition;
			if (position) {
				// Restore scroll position from history
				$(window).scrollTop(position.y).scrollLeft(position.x);
			} else if (
				idsEq(this.$store.state.currentSpecId, this.spec.id) &&
				!!this.$store.state.currentSpecScrollTop
			) {
				// Restore last saved scroll position on spec page
				$(window).scrollTop(this.$store.state.currentSpecScrollTop);
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

		padding: $page-header-vertical-padding $page-header-horiz-padding;
		@include mobile {
			padding: $page-header-vertical-padding-sm $page-header-horiz-padding-sm;
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
			margin-right: #{-$content-area-padding-sm};
		}
	}
}
</style>
