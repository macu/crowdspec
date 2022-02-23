<template>
<div class="spec-comment-preview community-target-preview" :class="{community: showCommunity}">

	<div class="info">

		<div class="right">

			<el-checkbox
				:modelValue="userRead"
				:disabled="sendingRead"
				@change="setUserRead"
				size="small"/>

			<el-tag
				v-if="userRead"
				size="small" type="success"
				@click="setUserRead(false)">
				Read
			</el-tag>
			<el-tag
				v-else
				size="small" type="info"
				@click="setUserRead(true)">
				Unread
			</el-tag>

			<el-button
				v-if="showCommunity"
				@click="openComments()"
				:type="!!unreadCount ? 'primary' : 'default'"
				size="small"
				round>
				<i class="material-icons">forum</i>
				<span v-if="(showUnreadOnly || unreadCount) || commentsCount">
					<template v-if="showUnreadOnly || unreadCount">
						<template v-if="unreadCount">{{unreadCount}} unread</template>
					</template>
					<template v-else-if="commentsCount">{{commentsCount}}</template>
				</span>
			</el-button>

			<el-button
				v-if="userCanEdit"
				@click="edit()"
				type="default"
				size="small"
				circle>
				<i class="material-icons">edit</i>
			</el-button>
			<el-button
				v-else-if="userCanDelete"
				@click="promptDelete()"
				type="warning"
				size="small"
				circle>
				<i class="material-icons">delete</i>
			</el-button>

		</div>

		<username :username="comment.username" :highlight="comment.highlight"/>
		posted
		<moment :datetime="comment.created" :offset="true"/>

		<template v-if="updated !== comment.created">
			(updated <moment :datetime="updated" :offset="true"/>)
		</template>

	</div>

	<div v-if="sendingDelete">
		<loading-message message="Deleting..."/>
	</div>
	<div v-else-if="sendingEdit">
		<loading-message message="Saving..."/>
	</div>
	<div v-else-if="editing" class="edit-comment-area">
		<el-input ref="editCommentInput" type="textarea" v-model="editingBody" :autosize="{minRows: 2}"/>
		<div>
			<el-button @click="cancelEdit()" :disabled="disableSaveEdit" type="warning" size="small">
				Cancel
			</el-button>
			<el-button @click="saveEdit()" :disabled="disableSaveEdit" type="primary" size="small">
				Save
			</el-button>
			<el-button @click="promptDelete()" :disabled="disableSaveEdit" type="danger" size="small">
				Delete
			</el-button>
		</div>
	</div>
	<div v-else class="body">{{body}}</div>

</div>
</template>

<script>
import $ from 'jquery';
import Username from '../widgets/username.vue';
import Moment from '../widgets/moment.vue';
import LoadingMessage from '../widgets/loading.vue';
import {TARGET_TYPE_COMMENT} from './const.js';
import {ajaxMarkRead, ajaxUpdateComment, ajaxDeleteComment} from './ajax.js';

export default {
	components: {
		Username,
		Moment,
		LoadingMessage,
	},
	props: {
		specId: Number,
		comment: Object,
		showCommunity: Boolean, // show unread count
		showUnreadOnly: Boolean,
	},
	emits: ['update-unread', 'open-comments', 'deleted'],
	data() {
		return {
			userRead: this.comment.userRead || false,
			sendingRead: false,
			editing: false,
			updated: this.comment.updated,
			body: this.comment.body,
			editingBody: '',
			sendingEdit: false,
			sendingDelete: false,
		};
	},
	computed: {
		userCanEdit() {
			return this.$store.getters.currentUserId === this.comment.userId;
		},
		userCanDelete() {
			return this.userCanEdit || this.$store.getters.currentSpecOwnedByUser;
		},
		disableSaveEdit() {
			return !this.editingBody.trim() || this.sendingEdit;
		},
		unreadCount() {
			return this.comment.unreadCount || 0;
		},
		commentsCount() {
			return this.comment.commentsCount || 0;
		},
	},
	watch: {
		editing(editing) {
			if (editing) {
				this.$nextTick(() => {
					if (this.$refs.editCommentInput) {
						$('textarea', this.$refs.editCommentInput.$el).focus();
					}
				});
			}
		},
	},
	methods: {
		setUserRead(read) {
			this.sendingRead = true;
			ajaxMarkRead(this.specId, TARGET_TYPE_COMMENT, this.comment.id, read).then(() => {
				this.sendingRead = false;
				this.userRead = read;
				this.$emit('update-unread', read ? -1 : 1);
			}).fail(() => {
				this.sendingRead = false;
			});
		},
		openComments() {
			this.$emit('open-comments', this.comment.id);
		},
		edit() {
			if (!this.editingBody.trim()) {
				this.editingBody = this.body;
			}
			this.editing = true;
		},
		cancelEdit() {
			this.editing = false;
		},
		saveEdit() {
			if (this.disableSaveEdit) {
				return;
			}
			this.sendingEdit = true;
			ajaxUpdateComment(this.specId, this.comment.id, this.editingBody).then(response => {
				this.sendingEdit = false;
				this.editing = false;
				this.updated = response.updated;
				this.body = response.body;
				this.editingBody = '';
			}).fail(() => {
				this.sendingEdit = false;
			});
		},
		promptDelete() {
			this.$confirm('Permanently delete this comment?', {
				confirmButtonText: 'Delete',
				cancelButtonText: 'Cancel',
				type: 'warning',
			}).then(() => {
				this.sendingDelete = true;
				ajaxDeleteComment(this.specId, this.comment.id).then(() => {
					this.sendingDelete = false;
					this.$emit('deleted', this.comment.id);
				}).fail(() => {
					this.sendingDelete = false;
				});
			}).catch(function() {
				// Cancelled
			});
		},
		isUnread() {
			return !this.userRead;
		},
	},
};
</script>

<style lang="scss">
@import '../_styles/_colours.scss';

.spec-comment-preview {
	border: medium solid $comment-bg;
	border-radius: 8px;
	background-color: $shadow-bg;
	padding: 20px;

	&.community {
		border: thin solid $comment-bg;
		background-color: unset;
	}

	>.info {
		>.right {
			float: right;
			white-space: nowrap;
			margin-left: 10px;
			margin-bottom: 10px;
			display: inline-flex;
			flex-direction: row;
			flex-wrap: nowrap;
			align-items: center;
			>* {
				margin: 0;
			}
			* {
				white-space: nowrap;
			}
			>.el-tag {
				margin-left: 5px;
				cursor: pointer; // click to toggle read
			}
			>.el-button {
				margin-left: 10px;
			}
			.el-checkbox__input.is-checked .el-checkbox__inner {
				color: white;
				background-color: $e-success;
				border-color: $e-success;
			}
		}
	}

	>.body {
		white-space: pre-wrap;
	}

	>* ~ * {
		margin-top: 20px;
	}

	>.edit-comment-area {
		>.el-textarea {
			margin-bottom: 10px;
			max-height: 60vh;
			overflow-y: auto;
		}
	}
}
</style>
