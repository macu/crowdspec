<template>
<section class="community-review-page">

	<header>
		<h2>Community review</h2>
	</header>

	<div class="content-page">

		<p v-if="loading">
			<loading-message message="Loading..."/>
		</p>

		<template v-else-if="$store.getters.loggedIn">

			<section>

				<template v-if="specs.length">
					<h3>Your specs</h3>
					<div v-for="s in specs" :key="s.id" class="spec review">
						<div class="flex-row">
							<div class="flex-row nowrap fill">
								<div class="expand">
									<template v-if="s.hasSubspecs">
										<el-button v-if="expandedSpecs[s.id]"
											@click="collapseSpec(s.id)"
											:type="s.hasUnreadSubspec ? 'primary' : 'default'"
											circle>
											<i class="material-icons">expand_less</i>
										</el-button>
										<el-button v-else
											@click="expandSpec(s.id)"
											:type="s.hasUnreadSubspec ? 'primary' : 'default'"
											circle>
											<i class="material-icons">expand_more</i>
										</el-button>
									</template>
									<!-- claim same space with hidden button -->
									<el-button v-else
										type="default"
										circle
										style="visibility:hidden;">
										<i class="material-icons">filter_center_focus</i>
									</el-button>
								</div>
								<div class="name fill">
									<router-link :to="{name: 'spec', params: {specId: s.id}}">
										{{s.name}}
									</router-link>
								</div>
							</div>
							<div>
								<el-button @click="openSpecCommunity(s.id)"
									:type="s.unread ? 'primary': 'default'">
									<i class="material-icons">forum</i>
									<span>
										{{s.unread}} unread
										({{s.total}} total)
									</span>
								</el-button>
							</div>
							<div>
								<el-button @click="gotoSpec(s.id)"
									:type="s.blockUnread ? 'primary' : 'default'">
									<i class="material-icons">folder</i>
									<span>
										{{s.blockUnread}} unread
										({{s.blockTotal}} total) on blocks
									</span>
								</el-button>
							</div>
						</div>
						<div v-if="expandedSpecs[s.id]" class="subspecs">
							<template v-if="subspecsBySpecId[s.id]">
								<div v-for="ss in subspecsBySpecId[s.id]" :key="ss.id" class="subspec review">
									<div class="flex-row">
										<div class="name fill">
											<router-link :to="{name: 'subspec', params: {specId: s.id, subspecId: ss.id}}">
												{{ss.name}}
											</router-link>
										</div>
										<div>
											<el-button @click="openSubspecCommunity(s.id, ss.id)"
												:type="ss.unread ? 'primary' : 'default'">
												<i class="material-icons">forum</i>
												<span>
													{{ss.unread}} unread
													({{ss.total}} total)
												</span>
											</el-button>
										</div>
										<div>
											<el-button @click="gotoSubspec(s.id, ss.id)"
												:type="ss.blockUnread ? 'primary' : 'default'">
												<i class="material-icons">folder</i>
												<span>
													{{ss.blockUnread}} unread
													({{ss.blockTotal}} total) on blocks
												</span>
											</el-button>
										</div>
									</div>
								</div>
							</template>
							<p v-else>
								<loading-message/>
							</p>
						</div>
					</div>
				</template>
				<p v-else>No specs</p>

			</section>

			<section>

				<h3>Your comments</h3>

				<div class="flex-row wrap-reverse comment-filters">
					<div class="fill nowraptext count-message">
						<template v-if="showUnreadOnly">
							<template v-if="totalComments === 1">
								1 comment with unread replies
							</template>
							<template v-else-if="totalComments > 0">
								{{totalComments}} comments with unread replies
							</template>
							<em v-else>No comments with unread replies</em>
						</template>
						<template v-else>
							<template v-if="totalComments === 1">
								1 comment
							</template>
							<template v-else-if="totalComments > 0">
								{{totalComments}} comment
							</template>
							<em v-else>No comments</em>
						</template>
					</div>
					<el-checkbox v-model="showUnreadOnly" @change="reloadComments()">
						Show only comments with unread replies
					</el-checkbox>
				</div>

				<template v-if="comments.length">
					<div v-for="c in comments" :key="c.id" class="comment review">
						<div class="flex-row">
							<div class="fill">
								<username :username="$store.getters.username"/>
								<template v-if="c.updated !== c.created">
									updated <moment :datetime="c.updated" :offset="true"/>
								</template>
								<template v-else>
									posted <moment :datetime="c.created" :offset="true"/>
								</template>
							</div>
							<div>
								<el-button @click="openCommentCommunity(c.specId, c.id)"
									:type="c.unread ? 'primary' : 'default'"
									size="small">
									<i class="material-icons">forum</i>
									<span>
										{{c.unread}} unread
										({{c.total}} total)
									</span>
								</el-button>
							</div>
						</div>
						<div class="body" v-text="c.body"/>
					</div>

					<p v-if="loadingCommentsPage">
						<loading-message message="Loading more comments..."/>
					</p>
					<el-button v-else-if="hasMoreComments" @click="loadMoreComments()">Load more</el-button>
				</template>
				<p v-else-if="loadingCommentsPage">
					<loading-message message="Loading..."/>
				</p>

			</section>

		</template>

		<el-alert v-else type="info" show-icon :closable="false">On this page a logged in user can view community spaces associated with their specs and comments.</el-alert>

	</div>

	<template v-if="$store.getters.loggedIn">

		<community-modal
			ref="communityModal"
			:spec-id="specId"
			@play-video="playVideo"
			@comment-updated="commentUpdated"
			@comment-deleted="commentDeleted"
			/>

		<play-video-modal
			ref="playVideoModal"
			/>

	</template>

</section>
</template>

<script>
import $ from 'jquery';
import Moment from '../widgets/moment.vue';
import CommunityModal from '../spec/community-modal.vue';
import PlayVideoModal from '../widgets/play-video-modal.vue';
import Username from '../widgets/username.vue';
import LoadingMessage from '../widgets/loading.vue';
import {alertError, idsEq} from '../utils.js';
import {
	TARGET_TYPE_SPEC,
	TARGET_TYPE_SUBSPEC,
	TARGET_TYPE_COMMENT,
} from '../spec/const';

function fetchCommunity(args) {
	return $.get('/ajax/community-review', args);
}

export default {
	components: {
		Moment,
		CommunityModal,
		PlayVideoModal,
		Username,
		LoadingMessage,
	},
	data() {
		return {
			loading: true,
			showUnreadOnly: false,
			specs: [],
			expandedSpecs: {},
			subspecsBySpecId: {},
			comments: [],
			totalComments: 0,
			hasMoreComments: false,
			loadingCommentsPage: false,
			specId: null, // used as param to community-modal
		};
	},
	computed: {
		commentsPageArgs() {
			let args = {
				unreadOnly: this.showUnreadOnly,
			};
			if (this.comments.length) {
				// Get updated time of last comment.
				// Page frame is based on updated rather than offset
				// because new comments may be updated while browsing.
				args.updatedBefore = this.comments[this.comments.length - 1].updated;
			}
			return args;
		},
	},
	beforeRouteEnter(to, from, next) {
		console.debug('beforeRouteEnter community-review', to);
		next(vm => {
			vm.initPage();
		});
	},
	beforeRouteUpdate(to, from, next) {
		console.debug('beforeRouteUpdate community-review', to);
		this.initPage();
		next();
	},
	beforeRouteLeave(to, from, next) {
		console.debug('beforeRouteLeave community-review');
		this.specs = [];
		this.expandedSpecs = {};
		this.subspecsBySpecId = {};
		this.comments = [];
		this.totalComments = 0;
		this.hasMoreComments = false;
		this.specId = null;
		next();
	},
	methods: {
		initPage() {
			this.showUnreadOnly = this.$store.getters.userSettings.community.unreadOnly;
			this.loading = true;
			fetchCommunity(
				$.extend({request: 'all'}, this.commentsPageArgs)
			).then(payload => {
				this.specs = payload.specs;
				this.comments = payload.comments;
				this.totalComments = payload.totalComments;
				this.hasMoreComments = payload.hasMoreComments;
				this.loading = false;
			}).fail(jqXHR => {
				if (jqXHR.status === 403) {
					// Show not logged in message
					this.loading = false;
					return;
				}
				this.$router.replace({
					name: 'ajax-error',
					params: {code: jqXHR.status},
					query: {url: encodeURIComponent(this.$route.fullPath)},
				});
			});
		},
		expandSpec(specId) {
			var prevVal = this.expandedSpecs[specId];
			this.expandedSpecs[specId] = true;
			if (typeof prevVal === 'undefined') {
				fetchCommunity({
					request: 'subspecs',
					specId,
				}).then(payload => {
					this.subspecsBySpecId[specId] = payload.subspecs;
				}).fail(jqXHR => {
					this.expandedSpecs[specId] = false;
					alertError(jqXHR);
				});
			}
		},
		collapseSpec(specId) {
			this.expandedSpecs[specId] = false;
		},
		reloadComments() {
			this.loadingCommentsPage = true;
			this.comments = [];
			fetchCommunity(
				$.extend({request: 'comments'}, this.commentsPageArgs)
			).then(response => {
				this.loadingCommentsPage = false;
				this.comments = response.comments;
				this.totalComments = response.totalComments;
				this.hasMoreComments = response.hasMoreComments;
			}).fail(jqXHR => {
				this.loadingCommentsPage = false;
				alertError(jqXHR);
			});
		},
		loadMoreComments() {
			this.loadingCommentsPage = true;
			fetchCommunity(
				$.extend({request: 'comments'}, this.commentsPageArgs)
			).then(response => {
				this.loadingCommentsPage = false;
				this.comments = this.comments.concat(response.comments);
				// Don't update totalComments - continue to show
				// the number of comments counted at the beginning of paging
				this.hasMoreComments = response.hasMoreComments;
			}).fail(jqXHR => {
				this.loadingCommentsPage = false;
				alertError(jqXHR);
			});
		},
		openSpecCommunity(specId) {
			this.specId = specId; // set prop for modal
			this.$nextTick(() => { // allow prop to apply
				this.$refs.communityModal.openCommunityReview(TARGET_TYPE_SPEC, specId,
					adjustUnread => {
						for (let i = 0; i < this.specs.length; i++) {
							if (idsEq(specId, this.specs[i].id)) {
								this.specs[i].unread = this.specs[i].unread + adjustUnread;
								break;
							}
						}
					}, adjustTotal => {
						for (let i = 0; i < this.specs.length; i++) {
							if (idsEq(specId, this.specs[i].id)) {
								this.specs[i].total = this.specs[i].total + adjustTotal;
								break;
							}
						}
					}
				);
			});
		},
		gotoSpec(specId) {
			this.$router.push({
				name: 'spec',
				params: {specId},
			});
		},
		openSubspecCommunity(specId, subspecId) {
			this.specId = specId; // set prop for modal
			this.$nextTick(() => { // allow prop to apply
				this.$refs.communityModal.openCommunityReview(TARGET_TYPE_SUBSPEC, subspecId,
					adjustUnread => {
						let subspecs = this.subspecsBySpecId[specId];
						for (let i = 0; i < subspecs.length; i++) {
							if (idsEq(subspecId, subspecs[i].id)) {
								subspecs[i].unread = subspecs[i].unread + adjustUnread;
								break;
							}
						}
						// Update spec.hasUnreadSubspec
						let hasUnreadSubspec = false;
						if (adjustUnread > 0) {
							hasUnreadSubspec = true;
						} else {
							for (let i = 0; i < subspecs.length; i++) {
								if (subspecs[i].unread > 0 || subspecs[i].blockUnread > 0) {
									hasUnreadSubspec = true;
									break;
								}
							}
						}
						for (let i = 0; i < this.specs.length; i++) {
							if (idsEq(specId, this.specs[i].id)) {
								this.specs[i].hasUnreadSubspec = hasUnreadSubspec;
								break;
							}
						}
					}, adjustTotal => {
						let subspecs = this.subspecsBySpecId[specId];
						for (let i = 0; i < subspecs.length; i++) {
							if (idsEq(subspecId, subspecs[i].id)) {
								subspecs[i].total = subspecs[i].total + adjustTotal;
								break;
							}
						}
					}
				);
			});
		},
		gotoSubspec(specId, subspecId) {
			this.$router.push({
				name: 'subspec',
				params: {specId, subspecId},
			});
		},
		openCommentCommunity(specId, commentId) {
			this.specId = specId; // set prop for modal
			this.$nextTick(() => { // allow prop to apply
				this.$refs.communityModal.openCommunityReview(TARGET_TYPE_COMMENT, commentId,
					adjustUnread => {
						for (let i = 0; i < this.comments.length; i++) {
							if (idsEq(commentId, this.comments[i].id)) {
								this.comments[i].unread = this.comments[i].unread + adjustUnread;
								break;
							}
						}
					}, adjustTotal => {
						for (let i = 0; i < this.comments.length; i++) {
							if (idsEq(commentId, this.comments[i].id)) {
								this.comments[i].total = this.comments[i].total + adjustTotal;
								break;
							}
						}
					}
				);
			});
		},
		playVideo(urlObject) {
			this.$refs.playVideoModal.show(urlObject);
		},
		commentUpdated(comment) {
			for (let i = 0; i < this.comments.length; i++) {
				if (this.comments[i].id === comment.id) {
					// Apply changes
					this.comments[i] = $.extend(true, this.comments[i], comment);
					break;
				}
			}
		},
		commentDeleted(commentId) {
			for (let i = 0; i < this.comments.length; i++) {
				if (this.comments[i].id === commentId) {
					// Remove from list
					this.comments.splice(i, 1);
					this.totalComments--;
					break;
				}
			}
		},
	},
};
</script>

<style lang="scss">
@import '../_styles/_breakpoints.scss';
@import '../_styles/_colours.scss';
@import '../_styles/_app.scss';

.community-review-page {

	>header {
		background-color: $spec-bg;
		color: white;

		padding: $page-header-vertical-padding $page-header-horiz-padding;
		@include mobile {
			padding: $page-header-vertical-padding-sm $page-header-horiz-padding-sm;
		}

		>h2 {
			margin: 0;
		}
	} // header

	section {
		margin-bottom: 60px;
		>h3 {
			margin: 0 0 40px 0;
			padding: 20px;
			background-color: $section-highlight;
		}
		>div.flex-row.comment-filters {
			padding: 0 10px;
			>* {
				margin-top: 0;
				margin-bottom: 10px;
			}
			>div.count-message {
				>em {
					color: gray;
					font-style: italic;
				}
			}
		}
	}

	.body {
		white-space: pre-wrap;
	}

	.review {
		margin: 10px 0;
	}

	.spec.review {
		>.flex-row {
			background-color: rgba($spec-bg, .2);
			>.flex-row {
				>.name {
					margin-left: 15px;
					background-color: white;
					padding: 5px 15px;
					min-width: 145px; // prevent squash
				}
			}
		}
		>.subspecs {
			padding-left: 33px;
			>.subspec.review {
				>.flex-row {
					background-color: rgba($subspec-bg, .2);
					>.name {
						background-color: white;
						padding: 5px 15px;
						min-width: 145px; // prevent squash
					}
				}
			}
		}
	}

	.comment.review {
		border: thin solid $comment-bg;
		border-radius: 8px;
		background-color: $shadow-bg;
		margin: 20px 0;
		padding: 5px;
		.body {
			padding: 0 10px 10px;
			white-space: pre-wrap;
		}
	}
}
</style>
