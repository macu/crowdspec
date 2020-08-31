<template>
<section v-if="subspec" class="subspec-page">

	<header>

		<div class="spec" @click="gotoSpec()">
			<h2>
				{{subspec.specName}}
			</h2>
		</div>

		<div class="subspec">

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

		</div>

	</header>

	<spec-view
		:key="subspec.id"
		:subspec="subspec"
		:enable-editing="enableEditing"
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
	data() {
		return {
			subspec: null,
		};
	},
	computed: {
		currentUserOwns() {
			return this.subspec.ownerType === OWNER_TYPE_USER &&
				this.$store.getters.userID === this.subspec.ownerId;
		},
		enableEditing() {
			// Currently users may edit only their own specs
			return this.currentUserOwns;
		},
	},
	beforeRouteEnter(to, from, next) {
		ajaxLoadSubspec(to.params.specId, to.params.subspecId).then(subspec => {
			next(vm => {
				vm.setSubspec(subspec);
			});
		}).fail(jqXHR => {
			next({name: 'ajax-error', params: {code: jqXHR.status}, replace: true});
		});
	},
	beforeRouteUpdate(to, from, next) {
		// Reload spec even if same across navigation as view must be rebuilt using latest state
		ajaxLoadSubspec(to.params.specId, to.params.subspecId).then(subspec => {
			this.setSubspec(subspec);
			next();
			this.$nextTick(this.restoreScroll);
		}).fail(jqXHR => {
			next({name: 'ajax-error', params: {code: jqXHR.status}, replace: true});
		});
	},
	beforeRouteLeave(to, from, next) {
		setWindowSubtitle(); // clear
		next();
	},
	methods: {
		setSubspec(subspec) {
			this.subspec = subspec;
			setWindowSubtitle(subspec.name);
			// vue-router scrollBehavior is applied before spec-view has a chance to populate,
			// so restore the scroll position again after fully rendering.
			this.$nextTick(this.restoreScroll);
		},
		gotoSpec() {
			this.$router.push({name: 'spec', params: {specId: this.subspec.specId}});
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
		margin-bottom: 1cm;

		>.spec {
			background-color: $spec-bg;
			color: white;
			cursor: pointer;

			padding: $page-header-vertical-padding $page-header-horiz-padding;
			@include mobile {
				padding: $page-header-vertical-padding-sm $page-header-horiz-padding-sm;
			}

			>h2 {
				margin: 0;
				padding: 0;
			}
		} // .spec

		>.subspec {
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
		} // .subspec
	} // header

	@include mobile {
		>.spec-view {
			margin-left: #{-$content-area-padding-sm};
			margin-right: #{-$content-area-padding-sm};
		}
	}
}
</style>
