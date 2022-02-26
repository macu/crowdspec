<template>
<el-dialog
	:title="modalTitle"
	v-model="showing"
	:width="$store.getters.dialogSmallWidth"
	:close-on-click-modal="false"
	@closed="closed()"
	custom-class="spec-block-community-modal">

	<context-stack
		ref="contextStack"
		@pop-stack="popStack"
		/>

	<p v-if="loading">
		<loading-message message="Loading..."/>
	</p>
	<p v-else-if="error">{{error}}</p>

	<template v-else-if="target">

		<div v-if="targetType !== TARGET_TYPE_COMMENT" class="updated">
			Last modified <strong><moment :datetime="targetUpdated" :offset="true"/></strong>
		</div>

		<preview-spec
			v-if="targetType === TARGET_TYPE_SPEC"
			:spec="target"
			/>
		<preview-subspec
			v-else-if="targetType === TARGET_TYPE_SUBSPEC"
			:subspec="target"
			/>
		<preview-block
			v-else-if="targetType === TARGET_TYPE_BLOCK"
			:block="target"
			@play-video="raisePlayVideo"
			/>
		<preview-comment
			v-else-if="targetType === TARGET_TYPE_COMMENT"
			ref="previewComment"
			:spec-id="specId"
			:comment="target"
			@deleted="commentDeleted"
			@update-unread="contextUpdateUnread"
			/>

		<div v-if="enableWrite" class="new-comment-area">
			<p v-if="sendingComment">
				<loading-message message="Posting comment..."/>
			</p>
			<template v-else-if="addingComment">
				<el-input
					v-if="addingComment"
					ref="newCommentInput"
					type="textarea"
					v-model="newCommentBody"
					:autosize="{minRows: 2}"
					placeholder="Type comment here"/>
				<div>
					<el-button @click="cancelAddComment()" type="warning">
						Cancel
					</el-button>
					<el-button @click="addComment()" :disabled="disablePostComment" type="primary">
						Post
					</el-button>
				</div>
			</template>
			<el-button v-else @click="addComment()" type="primary">
				Comment on this {{targetType}}
			</el-button>
		</div>

		<div class="controls-area flex-row wrap-reverse">
			<div class="fill nowraptext" v-text="formattedCommentsCount"/>
			<el-checkbox v-if="loggedIn" v-model="unreadOnly" @change="reloadCommunity()">
				Show only unread comments
			</el-checkbox>
		</div>

		<p v-if="reloadingComments">
			<loading-message message="Reloading..."/>
		</p>
		<template v-else-if="comments.length">

			<preview-comment
				v-for="c in comments"
				ref="comments"
				:key="c.id"
				:spec-id="specId"
				:comment="c"
				:show-community="true"
				:show-unread-only="unreadOnly"
				@open-comments="openComments"
				@update-unread="adjustUnread"
				@deleted="commentDeleted"
				/>

			<p v-if="loadingPage">
				<loading-message message="Loading more comments..."/>
			</p>
			<el-button v-else-if="hasMoreComments" @click="loadMoreComments()">
				Load more
			</el-button>

		</template>

	</template>

	<template #footer>
		<span class="dialog-footer">
			<el-button @click="showing = false">Close</el-button>
		</span>
	</template>

</el-dialog>
</template>

<script>
import $ from 'jquery';
import ContextStack from './community-context-stack.vue';
import Moment from '../widgets/moment.vue';
import PreviewSpec from './preview-spec.vue';
import PreviewSubspec from './preview-subspec.vue';
import PreviewBlock from './preview-block.vue';
import PreviewComment from './preview-comment.vue';
import LoadingMessage from '../widgets/loading.vue';
import {
	idsEq,
} from '../utils.js';
import {
	ajaxLoadCommunity,
	ajaxLoadCommentsPage,
	ajaxAddComment,
} from './ajax.js';
import {
	TARGET_TYPE_SPEC,
	TARGET_TYPE_SUBSPEC,
	TARGET_TYPE_BLOCK,
	TARGET_TYPE_COMMENT,
} from './const.js';

export default {
	components: {
		ContextStack,
		Moment,
		PreviewSpec,
		PreviewSubspec,
		PreviewBlock,
		PreviewComment,
		LoadingMessage,
	},
	props: {
		specId: Number,
		enableWrite: Boolean,
	},
	emits: ['play-video'],
	data() {
		return {
			showing: false,
			loading: false, // loading community space
			error: null,
			targetType: null,
			target: null,
			comments: [],
			unreadCount: 0,
			commentsCount: 0,
			hasMoreComments: false,
			loadingPage: false, // loading more comments
			addingComment: false, // during new comment composition
			newCommentBody: '',
			sendingComment: false, // during POST
			onAdjustUnread: null,
			onAdjustComments: null,
			reloadingComments: false,
		};
	},
	computed: {
		loggedIn() {
			return this.$store.getters.loggedIn;
		},
		TARGET_TYPE_SPEC() {
			return TARGET_TYPE_SPEC;
		},
		TARGET_TYPE_SUBSPEC() {
			return TARGET_TYPE_SUBSPEC;
		},
		TARGET_TYPE_BLOCK() {
			return TARGET_TYPE_BLOCK;
		},
		TARGET_TYPE_COMMENT() {
			return TARGET_TYPE_COMMENT;
		},
		modalTitle() {
			switch (this.targetType) {
				case TARGET_TYPE_SPEC:
					return 'Spec community';
				case TARGET_TYPE_SUBSPEC:
					return 'Subspec community';
				case TARGET_TYPE_BLOCK:
					return 'Block community';
				case TARGET_TYPE_COMMENT:
					return 'Comment community';
			}
			return 'Community space';
		},
		targetUpdated() {
			if (this.target && this.targetType !== TARGET_TYPE_COMMENT) {
				return this.target.updated;
			}
			return null;
		},
		disablePostComment() {
			return !this.newCommentBody.trim();
		},
		formattedCommentsCount() {
			let count = this.unreadOnly ? this.unreadCount : this.commentsCount;
			return count + (this.unreadOnly ? ' unread' : '') +
				(this.targetType === TARGET_TYPE_COMMENT ? ' sub' : '') +
				' comment' + ((count !== 1) ? 's' : '');
		},
		unreadOnly() {
			return this.$store.getters.loggedIn &&
				this.$store.getters.userSettings.community.unreadOnly;
		},
	},
	watch: {
		addingComment(adding) {
			if (adding) {
				this.$nextTick(() => {
					if (this.$refs.newCommentInput) {
						$('textarea', this.$refs.newCommentInput.$el).focus().select();
					}
				});
			}
		},
	},
	methods: {
		openCommunity(targetType, targetId, onAdjustUnread = null, onAdjustComments = null) {
			this.onAdjustUnread = onAdjustUnread;
			this.onAdjustComments = onAdjustComments;
			this.loadCommunity(targetType, targetId);
			this.showing = true;
		},
		openCommunityReview(targetType, targetId, onAdjustUnread = null, onAdjustComments = null) {
			this.onAdjustUnread = onAdjustUnread;
			this.onAdjustComments = onAdjustComments;
			this.loadCommunity(targetType, targetId, true);
			this.showing = true;
		},
		loadCommunity(targetType, targetId, loadStack = false) {
			this.loading = true;
			this.error = null;
			ajaxLoadCommunity(this.specId, targetType, targetId,
				this.unreadOnly, loadStack,
			).then(response => {
				this.loading = false;
				this.targetType = targetType;
				switch (targetType) {
					case TARGET_TYPE_SPEC:
						this.target = response.spec;
						break;
					case TARGET_TYPE_SUBSPEC:
						this.target = response.subspec;
						break;
					case TARGET_TYPE_BLOCK:
						this.target = response.block;
						break;
					case TARGET_TYPE_COMMENT:
						this.target = response.comment;
						break;
				}
				if (loadStack && response.stack) {
					this.$refs.contextStack.replaceStack(response.stack);
				}
				this.comments = response.comments;
				this.unreadCount = response.unreadCount;
				this.commentsCount = response.commentsCount;
				this.hasMoreComments = response.commentsCount > response.comments.length;
				this.newCommentBody = '';

				// Restore handlers from stack history
				let cachedItem = this.$refs.contextStack.retrieveCachedHandlers(targetType, targetId);
				if (cachedItem) {
					this.onAdjustUnread = cachedItem.onAdjustUnread;
					this.onAdjustComments = cachedItem.onAdjustComments;
				}
			}).fail(() => {
				this.loading = false;
				this.error = 'Failed to load community.';
				if (this.$refs.contextStack.checkEmpty()) {
					this.showing = false;
				}
			});
		},
		reloadCommunity() {
			this.reloadingComments = true;
			let updatedBefore = null; // start at beginning
			ajaxLoadCommentsPage(this.specId, this.targetType, this.target.id,
				updatedBefore, this.unreadOnly,
			).then(response => {
				this.reloadingComments = false;
				this.comments = response.comments;
				this.unreadCount = response.unreadCount;
				this.commentsCount = response.commentsCount;
				this.hasMoreComments = response.hasMore;
			}).fail(() => {
				this.reloadingComments = false;
			});
		},
		loadMoreComments() {
			this.loadingPage = true;
			// Get updated time of last comment.
			// Page frame is based on updated rather than offset
			// because new comments may be added while paging.
			let updatedBefore = this.comments[this.comments.length - 1].updated;
			ajaxLoadCommentsPage(this.specId, this.targetType, this.target.id,
				updatedBefore, this.unreadOnly,
			).then(response => {
				this.loadingPage = false;
				this.comments = this.comments.concat(response.comments);
				// Don't update commentsCount - continue to show
				// the number of comments counted at the beginning of paging
				this.hasMoreComments = response.hasMore;
			}).fail(() => {
				this.loadingPage = false;
			});
		},
		addComment() {
			if (!this.addingComment) {
				this.addingComment = true;
				return;
			}
			if (this.disablePostComment) {
				return;
			}
			this.sendingComment = true;
			ajaxAddComment(this.specId, this.targetType, this.target.id,
				this.newCommentBody,
			).then(newComment => {
				this.sendingComment = false;
				this.comments.unshift(newComment); // most recent appear first
				this.adjustComments(1); // one more known comment
				this.addingComment = false;
				this.newCommentBody = '';
				// Do not adjust unread count on parent item -
				// new comments marked read by author by default
			}).fail(() => {
				this.sendingComment = false;
			});
		},
		cancelAddComment() {
			this.addingComment = false;
			// Preserve newCommentBody in case user wishes to edit
		},
		openComments(commentId) {
			// add to stack
			this.$refs.contextStack.pushStack(
				this.targetType, this.target,
				this.onAdjustUnread, this.onAdjustComments,
			);
			this.onAdjustUnread = null; // no longer applies
			this.onAdjustComments = null;
			this.loadCommunity('comment', commentId);
		},
		adjustUnread(direction) {
			this.unreadCount += direction;
			if (this.onAdjustUnread) {
				this.onAdjustUnread(direction);
			}
		},
		adjustComments(direction) {
			this.commentsCount += direction;
			if (this.onAdjustComments) {
				this.onAdjustComments(direction);
			}
		},
		commentDeleted(commentId) {
			if (this.targetType === TARGET_TYPE_COMMENT && this.target.id === commentId) {
				let unread = this.$refs.previewComment.isUnread();
				// Back up stack - only way here is via sub comments
				let parentContext = this.$refs.contextStack.popStack(); // back up stack
				if (unread && parentContext.onAdjustUnread) {
					// Adjust unread count if parent context has onAdjustUnread,
					// as when deleting an immediate comment of a spec, subspec, or block
					// from within the comment's community view
					parentContext.onAdjustUnread(-1);
				}
				if (parentContext.onAdjustComments) {
					parentContext.onAdjustComments(-1);
				}
			} else {
				for (var i = 0; i < this.comments.length; i++) {
					if (idsEq(this.comments[i].id, commentId)) {
						let unread = !this.$refs.comments[i].userRead;
						this.comments.splice(i, 1); // Remove from array
						if (unread) {
							// if comment was unread, call onAdjustUnread
							this.adjustUnread(-1);
						}
						this.adjustComments(-1);
						break;
					}
				}
			}
		},
		contextUpdateUnread(read) {
			// Called when toggling read on comment representing current community context
			let parentContext = this.$refs.contextStack.getParentContext();
			if (parentContext && parentContext.onAdjustUnread) {
				parentContext.onAdjustUnread(read);
			}
		},
		popStack(targetType, targetId, onAdjustUnread, onAdjustComments) {
			// Cache current view in case of interaction handlers
			this.$refs.contextStack.cacheItemHandlers({
				targetType: this.targetType,
				target: this.target,
				onAdjustUnread: this.onAdjustUnread,
				onAdjustComments: this.onAdjustComments,
			});

			this.onAdjustUnread = onAdjustUnread; // restore for top-level context
			this.onAdjustComments = onAdjustComments; // restore for top-level context
			this.loadCommunity(targetType, targetId);
		},
		raisePlayVideo(urlObject) {
			this.$emit('play-video', urlObject);
		},
		closed() {
			this.targetType = null;
			this.target = null;
			this.error = null;
			this.addingComment = false;
			this.newCommentBody = '';
			this.onAdjustUnread = null;
			this.onAdjustComments = null;
			this.$refs.contextStack.clearStack();
		},
	},
};
</script>

<style lang="scss">
.spec-block-community-modal.el-dialog {
	>.el-dialog__body {
		>*:not(.updated, :last-child) {
			margin-bottom: 20px;
		}
		>.community-context-stack {
			margin-bottom: 25px;
			&.empty {
				display: none;
			}
		}
		>.updated {
			text-align: right;
			font-size: smaller;
			margin-bottom: 5px;
		}
		>.new-comment-area {
			margin: 45px 0;
			>.el-textarea {
				margin-bottom: 10px;
				max-height: 60vh;
				overflow-y: auto;
			}
		}
		>.controls-area {
			// Adjust for padding in .flex-row;
			// result is 20px spacing
			margin-bottom: 10px;
		}
	}
}
</style>
