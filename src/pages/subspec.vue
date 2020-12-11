<template>
<section class="subspec-page">

	<header v-if="subspec">

		<div v-if="onSubspecRoute && !loading" class="right">

			<el-button
				v-if="enableEditing"
				@click="openManageSubspec()"
				size="mini" icon="el-icon-setting"/>

			<span v-else>
				Last modified <moment :datetime="lastModifiedMoment" :offset="true"/>
			</span>

		</div>

		<h3>{{subspec.name}}</h3>

		<div v-if="subspec.desc" class="desc">{{subspec.desc}}</div>

	</header>

	<router-view
		ref="view"
		:loading="loading"
		:subspec="subspec"
		:enable-editing="enableEditing"
		@prompt-nav-spec="promptNavSpec()"
		/>

	<edit-subspec-modal
	 	v-if="enableEditing"
		ref="editSubspecModal"
		:spec-id="specId"
		/>

</section>
</template>

<script>
import $ from 'jquery';
import Moment from '../widgets/moment.vue';
import SpecView from '../spec/view.vue';
import EditSubspecModal from '../spec/edit-subspec-modal.vue';
import {ajaxLoadSubspec} from '../spec/ajax.js';
import {OWNER_TYPE_USER} from '../spec/const.js';
import {setWindowSubtitle, momentIsAfter, greatestMoment} from '../utils.js';

export default {
	components: {
		Moment,
		SpecView,
		EditSubspecModal,
	},
	inheritAttrs: false,
	props: {
		enableEditing: Boolean,
	},
	data() {
		return {
			loading: true,
			subspec: null,
		};
	},
	computed: {
		specId() {
			return parseInt(this.$route.params.specId, 10);
		},
		subspecId() {
			return parseInt(this.$route.params.subspecId, 10);
		},
		onSubspecRoute() {
			return this.$route.name === 'subspec';
		},
		lastModifiedMoment() {
			if (this.subspec) {
				return greatestMoment(this.subspec.updated, this.subspec.blocksUpdated);
			}
			return null;
		},
	},
	beforeRouteEnter(to, from, next) {
		console.debug('beforeRouteEnter subspec', to);
		next(vm => {
			vm.loadSubspec(to.params.specId, to.params.subspecId, to.name === 'subspec');
		});
	},
	beforeRouteUpdate(to, from, next) {
		console.debug('beforeRouteUpdate subspec', to);
		this.loadSubspec(to.params.specId, to.params.subspecId, to.name === 'subspec');
		next();
	},
	beforeRouteLeave(to, from, next) {
		console.debug('beforeRouteLeave subspec');
		this.subspec = null;
		setWindowSubtitle(); // clear
		next();
	},
	methods: {
		loadSubspec(specId, subspecId, loadBlocks) {
			console.debug('load subspec');
			this.loading = true;
			let cached = this.$store.getters.getCachedFullSubspec(subspecId);
			ajaxLoadSubspec(specId, subspecId, loadBlocks, cached).then(subspec => {
				console.debug('subspec loaded', subspec);
				if (
					loadBlocks && ( // blocks requested
						!cached || // no cached
						subspec.blocks || // blocks were returned
						momentIsAfter(subspec.updated, cached.updated) || // header updated since
						momentIsAfter(subspec.blocksUpdated, cached.blocksUpdated) // blocks updated since
					)
				) {
					this.subspec = subspec;
					this.$store.commit('cacheLatestFullSubspec', subspec);
				} else if (loadBlocks && cached) {
					// Latest available; returned spec doesn't include blocks
					this.subspec = cached;
				} else {
					this.subspec = subspec;
				}
				setWindowSubtitle(subspec.name);
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
		promptNavSpec() {
			this.$emit('prompt-nav-spec');
		},
		openManageSubspec() {
			this.$refs.editSubspecModal.showEdit(this.subspec, updatedSubspec => {
				this.subspec.updated = updatedSubspec.updated;
				this.subspec.name = updatedSubspec.name;
				this.subspec.desc = updatedSubspec.desc;
				setWindowSubtitle(updatedSubspec.name);
			});
		},
		restoreScroll() {
			let savedPosition = this.$store.state.savedScrollPosition;
			if (savedPosition && this.onSubspecRoute) {
				console.debug('restoreScroll subspec');
				$(window).scrollTop(savedPosition.y).scrollLeft(savedPosition.x);
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

.subspec-page {

	>header {
		background-color: $subspec-bg;
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
				margin-left: 10px;
			}
		}

		>h3 {
			margin: 0;
			padding: 0;
		}

		>.desc {
			white-space: pre-wrap;
			margin-top: 10px;
		}
	} // header
}
</style>
