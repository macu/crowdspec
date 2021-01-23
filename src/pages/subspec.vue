<template>
<section class="subspec-page">

	<header v-if="subspec">

		<div v-if="onSubspecRoute && !loading" class="right">

			<template v-if="enableEditing">
				<el-button
					@click="openSubspecCommunity()"
					:type="unreadCount ? 'primary' : 'default'"
					:disabled="choosingAddPosition"
					size="mini" icon="el-icon-chat-dot-square">
					<template v-if="unreadCount">{{unreadCount}}</template>
				</el-button>
				<el-button
					v-if="enableEditing"
					@click="openManageSubspec()"
					:disabled="choosingAddPosition"
					size="mini" icon="el-icon-setting"
					/>
			</template>
			<template v-else>
				<span>
					Last modified <moment :datetime="subspec.updated" :offset="true"/>
				</span>
				<el-button
					@click="openSubspecCommunity()"
					:type="unreadCount ? 'primary' : 'default'"
					size="mini" icon="el-icon-chat-dot-square">
					<template v-if="unreadCount">{{unreadCount}}</template>
				</el-button>
			</template>

		</div>

		<h3>{{subspec.name}}</h3>

		<div v-if="subspec.desc" class="desc">{{subspec.desc}}</div>

	</header>

	<router-view
		ref="view"
		:loading="loading"
		:subspec="subspec"
		:enable-editing="enableEditing"
		@prompt-nav-spec="promptNavSpec"
		@open-community="openCommunity"
		@play-video="playVideo"
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
import {
	OWNER_TYPE_USER,
	TARGET_TYPE_SUBSPEC,
} from '../spec/const.js';
import {setWindowSubtitle} from '../utils.js';

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
			unreadCount: 0,
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
		choosingAddPosition() {
			return !!this.$store.state.movingBlockId;
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
		this.unreadCount = 0;
		setWindowSubtitle(); // clear
		next();
	},
	methods: {
		loadSubspec(specId, subspecId, loadBlocks) {
			console.debug('load subspec');
			this.loading = true;
			ajaxLoadSubspec(specId, subspecId, loadBlocks).then(subspec => {
				console.debug('subspec loaded', subspec);
				this.subspec = subspec;
				this.unreadCount = subspec.unreadCount || 0;
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
		openSubspecCommunity() {
			this.$emit('open-community', TARGET_TYPE_SUBSPEC, this.subspec.id, adjustUnreadCount => {
				this.unreadCount += adjustUnreadCount;
			});
		},
		openCommunity(targetType, targetId, onAdjustUnread) {
			this.$emit('open-community', targetType, targetId, onAdjustUnread);
		},
		playVideo(urlObject) {
			this.$emit('play-video', urlObject);
		},
		openManageSubspec() {
			this.$refs.editSubspecModal.showEdit(this.subspec, updatedSubspec => {
				this.subspec.updated = updatedSubspec.updated;
				this.subspec.name = updatedSubspec.name;
				this.subspec.desc = updatedSubspec.desc;
				this.unreadCount = updatedSubspec.unreadCount || 0;
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
@import '../_styles/_breakpoints.scss';
@import '../_styles/_colours.scss';
@import '../_styles/_spec-view.scss';
@import '../_styles/_app.scss';

.subspec-page {

	>header {
		background-color: $subspec-bg;
		color: white;
		overflow: hidden; // keep {float: right} content bounded on mobile

		padding: $page-header-vertical-padding $page-header-horiz-padding;
		@include mobile {
			padding: $page-header-vertical-padding-sm $page-header-horiz-padding-sm;
		}

		>.right {
			float: right;
			font-size: small;
			margin-left: 20px;
			text-align: right;

			>* {
				// apply for spacing between wrapped elements
				margin-bottom: 5px;
			}
			>*+* {
				margin-left: 15px;
			}

			// remove bottom margin from bottom row of buttons
			margin-bottom: -5px;

			@media screen and (max-width: $max-sm) {
				>span {
					display: block;
					margin-bottom: 10px;
				}
				>*+* {
					margin-left: 0;
				}
				>.el-button {
					margin-bottom: 5px;
				}
				>.el-button + .el-button {
					margin-left: 15px;
				}
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
