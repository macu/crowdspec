<template>
<section v-if="subspec" class="subspec-page">

	<header>

		<div class="right">

			<el-button
				v-if="enableEditing"
				@click="openManageSubspec()"
				size="mini" icon="el-icon-setting"/>

			<span v-else>
				Last modified <moment :datetime="subspec.updated" :offset="true"/>
			</span>

		</div>

		<h3>{{subspec.name}}</h3>

		<div v-if="subspec.desc" class="desc">{{subspec.desc}}</div>

	</header>

	<router-view
		:subspec="subspec"
		:enable-editing="enableEditing"
		@prompt-nav-spec="promptNavSpec()"
		/>

	<edit-subspec-modal
	 	v-if="enableEditing"
		ref="editSubspecModal"
		:spec-id="subspec.specId"
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
import {setWindowSubtitle} from '../utils.js';

export default {
	components: {
		Moment,
		SpecView,
		EditSubspecModal,
	},
	props: {
		enableEditing: Boolean,
	},
	data() {
		return {
			subspec: null,
		};
	},
	beforeRouteEnter(to, from, next) {
		console.debug('beforeRouteEnter subspec', to);
		ajaxLoadSubspec(to.params.specId, to.params.subspecId, to.name === 'subspec').then(subspec => {
			next(vm => {
				vm.setSubspec(subspec);
			});
		}).fail(jqXHR => {
			next({
				name: 'ajax-error',
				params: {code: jqXHR.status},
				query: {url: encodeURIComponent(to.fullPath)},
				replace: true,
			});
		});
	},
	beforeRouteUpdate(to, from, next) {
		console.debug('beforeRouteUpdate subspec', to);
		// Reload spec even if same across navigation as view must be rebuilt using latest state
		ajaxLoadSubspec(to.params.specId, to.params.subspecId, to.name === 'subspec').then(subspec => {
			this.setSubspec(subspec);
			next();
			this.$nextTick(this.restoreScroll);
		}).fail(jqXHR => {
			next({
				name: 'ajax-error',
				params: {code: jqXHR.status},
				query: {url: encodeURIComponent(to.fullPath)},
				replace: true,
			});
		});
	},
	beforeRouteLeave(to, from, next) {
		console.debug('beforeRouteLeave subspec');
		this.subspec = null;
		setWindowSubtitle(); // clear
		next();
	},
	methods: {
		setSubspec(subspec) {
			console.debug('setSubspec');
			this.subspec = subspec;
			setWindowSubtitle(subspec.name);
			// vue-router scrollBehavior is applied before spec-view has a chance to populate,
			// so restore the scroll position again after fully rendering.
			this.$nextTick(this.restoreScroll);
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
			let position = this.$store.state.savedScrollPosition;
			if (position) {
				console.debug('restoreScroll subspec');
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
