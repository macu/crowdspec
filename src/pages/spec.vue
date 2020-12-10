<template>
<section class="spec-page">

	<header v-if="spec">

		<div v-if="onSpecRoute && !loading" class="right">

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

		<h2 @click="gotoSpec()">{{spec.name}}</h2>

		<div v-if="spec.desc" class="desc">{{spec.desc}}</div>

	</header>

	<router-view
		ref="view"
		:loading="loading"
		:spec="spec"
		:enable-editing="enableEditing"
		@prompt-nav-spec="promptNavSpec()"
		/>

	<edit-spec-modal
		v-if="enableEditing"
		ref="editSpecModal"
		/>

	<nav-spec-modal
		ref="navSpecModal"
		:spec-id="specId"
		:subspec-id="subspecId"
		/>

</section>
</template>

<script>
import $ from 'jquery';
import Moment from '../widgets/moment.vue';
import EditSpecModal from '../spec/edit-spec-modal.vue';
import NavSpecModal from '../spec/nav-spec-modal.vue';
import {ajaxLoadSpec} from '../spec/ajax.js';
import {OWNER_TYPE_USER} from '../spec/const.js';
import {setWindowSubtitle, idsEq} from '../utils.js';

export default {
	components: {
		Moment,
		EditSpecModal,
		NavSpecModal,
	},
	data() {
		return {
			loading: true,
			spec: null,
		};
	},
	computed: {
		specId() {
			return parseInt(this.$route.params.specId, 10);
		},
		subspecId() {
			return parseInt(this.$route.params.subspecId, 10) || null;
		},
		onSpecRoute() {
			return this.$route.name === 'spec';
		},
		currentUserOwns() {
			return this.spec &&
				this.spec.ownerType === OWNER_TYPE_USER &&
				this.$store.getters.userID === this.spec.ownerId;
		},
		enableEditing() {
			// Currently users may edit only their own specs
			return this.currentUserOwns;
		},
	},
	beforeRouteEnter(to, from, next) {
		console.debug('beforeRouteEnter spec', to);
		next(vm => {
			vm.loadSpec(to.params.specId, to.name === 'spec');
		});
	},
	beforeRouteUpdate(to, from, next) {
		console.debug('beforeRouteUpdate spec', to);
		this.loadSpec(to.params.specId, to.name === 'spec');
		next();
	},
	beforeRouteLeave(to, from, next) {
		console.debug('beforeRouteLeave spec');
		this.spec = null;
		setWindowSubtitle(); // clear
		next();
	},
	methods: {
		loadSpec(specId, loadBlocks) {
			console.debug('load spec');
			this.loading = true;
			ajaxLoadSpec(specId, loadBlocks).then(spec => {
				console.debug('spec loaded', spec);
				this.spec = spec;
				setWindowSubtitle(spec.name);
				this.loading = false;
				this.$refs.view.$once('rendered', this.restoreScroll);
			}).fail(jqXHR => {
				this.$router.replace({
					name: 'ajax-error',
					params: {code: jqXHR.status},
					query: {url: encodeURIComponent(this.$route.fullPath)},
				});
			});
		},
		gotoSpec() {
			if (this.$route.name !== 'spec') {
				this.$router.push({
					name: 'spec',
					params: {specId: this.specId},
				});
			}
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
		promptNavSpec() {
			this.$refs.navSpecModal.show();
		},
		restoreScroll(position) {
			if (this.onSpecRoute) {
				let position = this.$store.state.savedScrollPosition;
				if (position) {
					console.debug('restoreScroll spec');
					// Restore scroll position from history
					$(window).scrollTop(position.y).scrollLeft(position.x);
				} else if (
					idsEq(this.$store.state.currentSpecId, this.spec.id) &&
					!!this.$store.state.currentSpecScrollTop
				) {
					console.debug('restoreScroll spec from currentSpecScrollTop');
					// Restore last saved scroll position on spec page when returning through forward nav
					$(window).scrollTop(this.$store.state.currentSpecScrollTop);
				}
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
			cursor: pointer;
		}

		>.desc {
			white-space: pre-wrap;
			margin-top: 10px;
			color: white;
		}
	} // header
}
</style>
