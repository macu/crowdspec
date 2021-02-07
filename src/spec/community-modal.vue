<template>
<el-dialog
	:title="modalTitle"
	:visible.sync="showing"
	:width="$store.getters.dialogSmallWidth"
	:close-on-click-modal="false"
	@closed="closed()"
	class="spec-block-community-modal">

	<context-stack
		ref="contextStack"
		@pop-stack="popStack"
		/>

	<p v-if="loading"><i class="el-icon-loading"/> Loading...</p>
	<p v-else-if="error">{{error}}</p>

	<template v-else-if="target">

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
			/>


		<div class="flex-row">
			<div class="fill" v-text="formattedCommentsCount"/>
			<el-checkbox v-model="unreadOnly" @change="reloadCommunity()">
				Show only unread comments
			</el-checkbox>
		</div>

		<div class="new-comment-area">
			<p v-if="sendingComment"><i class="el-icon-loading"/> Posting comment...</p>
			<template v-else-if="addingComment">
				<el-input
					v-if="addingComment"
					ref="newCommentInput"
					type="textarea"
					v-model="newCommentBody"
					:autosize="{minRows: 2}"
					placeholder="Type comment here"/>
				<div>
					<el-button @click="cancelAddComment()" type="warning" size="small">
						Cancel
					</el-button>
					<el-button @click="addComment()" :disabled="disablePostComment" type="primary" size="small">
						Post
					</el-button>
				</div>
			</template>
			<el-button v-else @click="addComment()" type="primary">
				Add comment
			</el-button>
		</div>

		<p v-if="reloadingComments"><i class="el-icon-loading"/> Reloading...</p>
		<template v-else-if="comments.length">

			<preview-comment
				v-for="c in comments"
				ref="comments"
				:key="c.id"
				:spec-id="specId"
				:comment="c"
				:show-community="true"
				@open-comments="openComments"
				@update-unread="adjustUnread"
				@deleted="commentDeleted"
				/>

			<p v-if="loadingPage"><i class="el-icon-loading"/> Loading more comments...</p>
			<el-button v-else-if="hasMoreComments" @click="loadMoreComments()">
				Load more
			</el-button>

		</template>

	</template>

	<span slot="footer" class="dialog-footer">
		<el-button @click="showing = false">Close</el-button>
	</span>

</el-dialog>
</template>

<script>
import $ from 'jquery';
import ContextStack from './community-context-stack.vue';
import PreviewSpec from './preview-spec.vue';
import PreviewSubspec from './preview-subspec.vue';
import PreviewBlock from './preview-block.vue';
import PreviewComment from './preview-comment.vue';
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
		PreviewSpec,
		PreviewSubspec,
		PreviewBlock,
		PreviewComment,
	},
	props: {
		specId: Number,
	},
	data() {
		return {
			showing: false,
			loading: false, // loading community space
			error: null,
			targetType: null,
			target: null,
			comments: [],
			commentsCount: 0,
			hasMoreComments: false,
			loadingPage: false, // loading more comments
			addingComment: false, // during new comment composition
			newCommentBody: '',
			sendingComment: false, // during POST
			onAdjustUnread: null,
			unreadOnly: false,
			reloadingComments: false,
		};
	},
	computed: {
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
		disablePostComment() {
			return !this.newCommentBody.trim();
		},
		formattedCommentsCount() {
			return this.commentsCount +
				(this.targetType === TARGET_TYPE_COMMENT ? ' sub' : '') +
				' comment' + ((this.commentsCount !== 1) ? 's' : '');
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
		openCommunity(targetType, targetId, onAdjustUnread = null) {
			this.onAdjustUnread = onAdjustUnread;
			this.unreadOnly = this.$store.getters.userSettings.community.unreadOnly;
			this.loadCommunity(targetType, targetId);
			this.showing = true;
		},
		loadCommunity(targetType, targetId) {
			this.loading = true;
			this.error = null;
			ajaxLoadCommunity(this.specId, targetType, targetId,
				this.unreadOnly,
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
				this.comments = response.comments;
				this.commentsCount = response.commentsCount;
				this.hasMoreComments = response.commentsCount > response.comments.length;
				this.newCommentBody = '';
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
				this.commentsCount++; // one more known comment
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
				this.targetType, this.target, this.onAdjustUnread);
			this.onAdjustUnread = null; // no longer applies
			this.loadCommunity('comment', commentId);
		},
		adjustUnread(direction) {
			if (this.onAdjustUnread) {
				this.onAdjustUnread(direction);
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
			} else {
				for (var i = 0; i < this.comments.length; i++) {
					if (idsEq(this.comments[i].id, commentId)) {
						let unread = !this.$refs.comments[i].userRead;
						this.comments.splice(i, 1); // Remove from array
						this.commentsCount--; // one less known comment
						if (unread) {
							// if comment was unread, call onAdjustUnread
							this.adjustUnread(-1);
						}
						break;
					}
				}
			}
		},
		popStack(targetType, targetId, onAdjustUnread) {
			this.onAdjustUnread = onAdjustUnread; // restore for top-level context
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
			this.$refs.contextStack.clearStack();
		},
	},
};
</script>

<style lang="scss">
.spec-block-community-modal {
	>.el-dialog {
		>.el-dialog__body {
			>*:not(:last-child) {
				margin-bottom: 20px;
			}
			>.community-context-stack {
				margin-bottom: 25px;
				&.empty {
					display: none;
				}
			}
			>.new-comment-area {
				margin: 45px 0 40px;
				>.el-textarea {
					margin-bottom: 10px;
					max-height: 60vh;
					overflow-y: auto;
				}
			}
			.flex-row {
				display: flex;
				flex-direction: row;
				flex-wrap: wrap;
				align-items: center;
				>* {
					margin-bottom: 5px;
				}
				>*+* {
					margin-left: 20px;
				}
				>.fill {
					flex: 1;
				}
			}
		}
	}
}
</style>
