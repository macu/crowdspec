<template>
<section class="spec-page">

	<header v-if="spec">

		<div v-if="!loading" class="right">

			<span v-if="currentUserOwns">You own this</span>
			<span v-else>
				Owned by
				<username v-if="spec.username" :username="spec.username" :highlight="spec.highlight"/>
				<template v-else>{{spec.ownerType}} {{spec.ownerId}}</template>
			</span>

			<template v-if="onSpecRoute">
				<template v-if="enableEditing">
					<span v-if="spec.public">
						Public
					</span>
					<span v-else>
						<el-tooltip content="Unpublished" placement="left">
							<i class="el-icon-lock"></i>
						</el-tooltip>
					</span>
					<el-button
						@click="openSpecCommunity()"
						:type="unreadCount ? 'primary' : 'default'"
						:disabled="choosingAddPosition"
						size="mini" icon="el-icon-chat-dot-square">
						<template v-if="unreadCount">{{unreadCount}}</template>
					</el-button>
					<el-button
						@click="openManageSpec()"
						:disabled="choosingAddPosition"
						size="mini" icon="el-icon-setting"
						/>
				</template>
				<template v-else>
					<span>
						Last modified <moment :datetime="spec.updated" :offset="true"/>
					</span>
					<el-button
						@click="openSpecCommunity()"
						:type="unreadCount ? 'primary' : 'default'"
						size="mini" icon="el-icon-chat-dot-square">
						<template v-if="unreadCount">{{unreadCount}}</template>
					</el-button>
				</template>
			</template>

			<el-button @click="promptNavSpec()" size="mini" icon="el-icon-folder"/>

		</div>

		<h2 @click="gotoSpec()">{{spec.name}}</h2>

		<div v-if="spec.desc" class="desc">{{spec.desc}}</div>

	</header>

	<router-view
		ref="view"
		:loading="loading"
		:spec="spec"
		:enable-editing="enableEditing"
		@prompt-nav-spec="promptNavSpec"
		@open-community="openCommunity"
		@play-video="playVideo"
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

	<community-modal
		ref="communityModal"
		:spec-id="specId"
		@play-video="playVideo"
		/>

	<play-video-modal
		ref="playVideoModal"
		/>

</section>
</template>

<script>
import $ from 'jquery';
import Username from '../widgets/username.vue';
import Moment from '../widgets/moment.vue';
import EditSpecModal from '../spec/edit-spec-modal.vue';
import NavSpecModal from '../spec/nav-spec-modal.vue';
import CommunityModal from '../spec/community-modal.vue';
import PlayVideoModal from '../widgets/play-video-modal.vue';
import {ajaxLoadSpec} from '../spec/ajax.js';
import {OWNER_TYPE_USER, TARGET_TYPE_SPEC} from '../spec/const.js';
import {setWindowSubtitle, idsEq} from '../utils.js';

export default {
	components: {
		Username,
		Moment,
		EditSpecModal,
		NavSpecModal,
		CommunityModal,
		PlayVideoModal,
	},
	data() {
		return {
			loading: true,
			spec: null,
			unreadCount: 0,
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
				this.$store.getters.currentUserId === this.spec.ownerId;
		},
		enableEditing() {
			// Currently users may edit only their own specs
			return this.currentUserOwns;
		},
		choosingAddPosition() {
			return !!this.$store.state.movingBlockId;
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
		this.unreadCount = 0;
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
				this.unreadCount = spec.unreadCount || 0;
				setWindowSubtitle(spec.name);
				this.loading = false;
				this.$refs.view.$once('rendered', this.restoreScroll);
				this.$store.commit('saveCurrentSpec', spec);
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
				this.unreadCount = updatedSpec.unreadCount || 0;
				setWindowSubtitle(updatedSpec.name);
			});
		},
		openSpecCommunity() {
			this.$refs.communityModal.openCommunity(TARGET_TYPE_SPEC, this.spec.id, adjustUnreadCount => {
				this.unreadCount += adjustUnreadCount;
			});
		},
		openCommunity(targetType, targetId, onAdjustUnread) {
			this.$refs.communityModal.openCommunity(targetType, targetId, onAdjustUnread);
		},
		playVideo(urlObject) {
			this.$refs.playVideoModal.show(urlObject);
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
					idsEq(this.$store.getters.currentSpecId, this.spec.id) &&
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
@import '../_styles/_breakpoints.scss';
@import '../_styles/_colours.scss';
@import '../_styles/_spec-view.scss';
@import '../_styles/_app.scss';

.spec-page {

	>header {
		background-color: $spec-bg;
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

			.username {
				display: inline-block;
				margin-left: $icon-spacing;
			}

			@include mobile {
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
