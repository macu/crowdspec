<template>
<li :data-spec-block="block.id" class="spec-block" :class="classes" @touchstart="touchstart" @selectstart="selectstart">

	<div class="content" :class="{'clear-ref': clearRef, 'mobile-adjust': mobileAdjust}">

		<div class="bg"></div>

		<div class="layover" @mouseleave="mouseLeaveLayover()">
			<template v-if="enableEditingDynamic">
				<div class="expand-control" :class="{hide: showActions}">
					<!-- only show community button to admins in collapsed mobile menu when there are unread submissions -->
					<!-- always show community button to guests, who only make community interactions -->
					<el-button v-if="showUnreadCount"
						@click="openCommunity()"
						type="primary"
						size="small"
						round>
						<i class="material-icons">forum</i>
						<span>{{unreadCount}} unread</span>
					</el-button>
					<el-button @click="focusActions = true" type="default" size="small" circle>
						<i class="material-icons">more_horiz</i>
					</el-button>
				</div>
				<div class="actions" :class="{show: showActions}">
					<template v-if="choosingAddPosition">
						<el-button @click="cancelChooseAddPosition()" type="warning" size="small" circle>
							<i class="material-icons">close</i>
						</el-button>
						<el-button @click="addBeforeThis()" type="primary" size="small" circle>
							<i class="material-icons">arrow_upward</i>
						</el-button>
						<el-button @click="addIntoThis()" type="primary" size="small" circle>
							<i class="material-icons">south_east</i>
						</el-button>
						<el-button @click="addAfterThis()" type="primary" size="small" circle>
							<i class="material-icons">arrow_downward</i>
						</el-button>
					</template>
					<template v-else-if="movingThis">
						<el-button @click="cancelMoving()" type="warning" size="small">
							<i class="material-icons">close</i>
							<span>
								<template v-if="$store.getters.mobileViewport">Cancel</template>
								<template v-else>Cancel move</template>
							</span>
						</el-button>
						<el-button @click="promptNavSpec()" size="small">
							<i class="material-icons">input</i>
							<span>
								<template v-if="$store.getters.mobileViewport">Context</template>
								<template v-else>Change context</template>
							</span>
						</el-button>
						<el-checkbox key="removeFromMoving" :data-moving-block-id="block.id" :modelValue="true" @click="removeThisFromMovingBlocks()"/>
					</template>
					<template v-else-if="movingOtherBlocks">
						<el-button @click="cancelMoving()" type="warning" size="small" circle>
							<i class="material-icons">close</i>
						</el-button>
						<el-button @click="moveBeforeThis()" type="success" size="small" circle>
							<i class="material-icons">arrow_upward</i>
						</el-button>
						<el-button @click="moveIntoThis()" type="success" size="small" circle>
							<i class="material-icons">south_east</i>
						</el-button>
						<el-button @click="moveAfterThis()" type="success" size="small" circle>
							<i class="material-icons">arrow_downward</i>
						</el-button>
						<el-checkbox key="addToMoving" v-if="showAddToMoving" :modelValue="false" @click="addThisToMovingBlocks()"/>
					</template>
					<template v-else>
						<el-button @click="openCommunity()"
							:type="showUnreadCount ? 'primary' : 'default'"
							size="small"
							round>
							<i class="material-icons">forum</i>
							<span v-if="showUnreadCount">{{unreadCount}} unread</span>
							<span v-else-if="commentsCount">{{commentsCount}}</span>
						</el-button>
						<el-button @click="editBlock()" type="default" size="small" circle>
							<i class="material-icons">edit</i>
						</el-button>
						<el-button v-if="showDeleteButton" @click="promptDeleteBlock()" type="warning" size="small" circle>
							<i class="material-icons">delete</i>
						</el-button>
						<el-button @click="enterChooseAddPosition()" type="primary" size="small" circle>
							<i class="material-icons">add</i>
						</el-button>
						<el-button @click="startMoving()" class="move-action" type="default" size="small" circle>
							<i class="material-icons">drag_handle</i>
						</el-button>
						<i @click="startMoving()" class="material-icons drag-handle">drag_handle</i>
					</template>
				</div>
			</template>
			<div v-else class="visitor-actions">
				<el-button
					@click="openCommunity()"
					:type="showUnreadCount ? 'primary' : 'default'"
					size="small"
					round>
					<i class="material-icons">forum</i>
					<span v-if="showUnreadCount">{{unreadCount}} unread</span>
					<span v-else-if="commentsCount">{{commentsCount}}</span>
				</el-button>
			</div>
		</div>
		<div v-if="currentlyMovingBlocks" class="layover parent-moving-layover">
			<div>
				<el-checkbox :modelValue="true" disabled/>
			</div>
		</div>

		<div v-if="hasTitle" class="title">{{title}}</div>

		<template v-if="hasRefItem">
			<ref-url v-if="refType === REF_TYPE_URL" :item="refItem" class="ref-item" @play="raisePlayVideo(refItem)"/>
			<ref-subspec v-else-if="refType === REF_TYPE_SUBSPEC" :item="refItem" class="ref-item"/>
		</template>

		<el-alert v-else-if="refItemMissing"
			title="Content unavailable"
			:closable="false"
			type="warning"/>

		<template v-if="hasBody">
			<div v-if="contentType === CONTENT_TYPE_PLAIN"
				class="body plain" v-text="body"></div>
			<div v-else-if="contentType === CONTENT_TYPE_MARKDOWN"
				ref="renderedHtml"
				class="body markdown" v-html="renderedHtml"></div>
		</template>

	</div>

	<ul ref="sublist" class="spec-block-list" :class="{'moving': movingThis}">
		<!-- managed programatically -->
	</ul>

</li>
</template>

<script>
import $ from 'jquery';
import RefUrl from './ref-url.vue';
import RefSubspec from './ref-subspec.vue';
import {
	CONTENT_TYPE_PLAIN, CONTENT_TYPE_MARKDOWN,
	REF_TYPE_URL, REF_TYPE_SUBSPEC,
} from './const.js';
import {idsEq} from '../utils.js';
import {
	SCRIPT_HLJS,
	loadScript,
} from '../widgets/script-loader.js';

const TOUCH_DELAY_CLEAR_SELECTION = 200;

export default {
	components: {
		RefUrl,
		RefSubspec,
	},
	props: {
		block: Object,
		subspecId: Number,
		eventBus: Object,
		enableEditing: Boolean,
		justAdded: Boolean,
	},
	data() {
		return {
			// Copy dynamic values
			updated: this.block.updated,
			styleType: this.block.styleType,
			contentType: this.block.contentType,
			refType: this.block.refType,
			refId: this.block.refId,
			title: this.block.title,
			body: this.block.body,
			renderedHtml: this.block.html,
			refItem: this.block.refItem,
			unreadCount: this.block.unreadCount || 0,
			commentsCount: this.block.commentsCount || 0,
			subblocks: this.block.subblocks ? this.block.subblocks.slice() : [],

			// Dynamic
			choosingAddPosition: false,
			focusActions: false,
			hasSubblocks: !!(this.block.subblocks && this.block.subblocks.length),
		};
	},
	computed: {
		CONTENT_TYPE_PLAIN() {
			return CONTENT_TYPE_PLAIN;
		},
		CONTENT_TYPE_MARKDOWN() {
			return CONTENT_TYPE_MARKDOWN;
		},
		REF_TYPE_URL() {
			return REF_TYPE_URL;
		},
		REF_TYPE_SUBSPEC() {
			return REF_TYPE_SUBSPEC;
		},
		enableEditingDynamic() {
			return this.$store.getters.loggedIn && this.enableEditing;
		},
		showDeleteButton() {
			switch (this.$store.getters.userSettings.blockEditing.deleteButton) {
				case 'modal':
					return false;
				case 'recent':
					return this.justAdded;
				default:
					return true;
			}
		},
		hasTitle() {
			return !!(this.title && this.title.trim());
		},
		hasBody() {
			return !!(
				(this.body && this.body.trim()) ||
				(this.renderedHtml && this.renderedHtml.trim())
			);
		},
		hasRefItem() {
			return !!(this.refType && this.refItem);
		},
		refItemMissing() {
			return !!this.refType && !this.refItem;
		},
		classes() {
			return {
				[this.styleType]: true,
				'title-only': this.hasTitle && !this.hasBody && !this.hasRefItem && !this.hasSubblocks,
				'ref-item-only': this.hasRefItem && !this.hasTitle && !this.hasBody && !this.hasSubblocks,
			};
		},
		currentlyMovingBlocks() {
			return this.$store.getters.currentlyMovingBlocks;
		},
		movingThis() {
			return this.$store.getters.currentlyMovingBlock(this.block.id);
		},
		movingOtherBlocks() {
			return this.currentlyMovingBlocks && !this.movingThis;
		},
		showAddToMoving() {
			return this.movingOtherBlocks &&
				idsEq(this.$store.state.movingBlocksSourceSubspecId, this.subspecId);
		},
		showActions() {
			return this.focusActions || this.currentlyMovingBlocks;
		},
		showUnreadOnly() {
			return this.$store.getters.loggedIn &&
				this.$store.getters.userSettings.community.unreadOnly;
		},
		showUnreadCount() {
			return this.$store.getters.loggedIn &&
				!!this.unreadCount;
		},
		clearRef() {
			// whether to add {clear: both} to ref item area
			// clear because "N unread" appears in actions
			return !!this.unreadCount;
		},
		mobileAdjust() {
			// whether to add {clear: both} to ref item area on mobile
			// (show layover above rather than to right of ref item)
			return this.showActions ||
				!!this.unreadCount ||
				(!this.showUnreadOnly && !!this.commentsCount);
		},
	},
	watch: {
		renderedHtml() {
			this.addCodeHighlighting();
		},
	},
	mounted() {
		this.eventBus.on('url-updated', this.urlUpdated);
		this.eventBus.on('url-deleted', this.urlDeleted);
		this.addCodeHighlighting();
	},
	beforeDestroy() {
		this.eventBus.off('url-updated', this.urlUpdated);
		this.eventBus.off('url-deleted', this.urlDeleted);
	},
	methods: {
		getBlockId() {
			return this.block.id;
		},
		getStyleType() {
			return this.styleType;
		},
		addCodeHighlighting() {
			if (!(this.contentType === CONTENT_TYPE_MARKDOWN && this.renderedHtml)) {
				return;
			}
			this.$nextTick(() => {
				let $codeblocks = $('pre>code[class*="language-"]', this.$refs.renderedHtml);
				if ($codeblocks.length) {
					loadScript(SCRIPT_HLJS).then(hljs => {
						$codeblocks.each((i, e) => {
							hljs.highlightBlock(e);
						});
					});
				}
			});
		},
		getParentId() {
			let $parent = $(this.$el).parent().closest('[data-spec-block]');
			return $parent.length ? $parent.attr('data-spec-block') : null;
		},
		getFollowingBlockId() {
			let $nextBlock = $(this.$el).next('[data-spec-block]');
			return $nextBlock.length ? $nextBlock.attr('data-spec-block') : null;
		},
		editBlock() {
			this.raiseOpenEdit(this.block.id, updatedBlock => {
				this.updated = updatedBlock.updated;
				this.styleType = updatedBlock.styleType;
				this.contentType = updatedBlock.contentType;
				this.refType = updatedBlock.refType;
				this.refId = updatedBlock.refId;
				this.refItem = updatedBlock.refItem;
				this.title = updatedBlock.title;
				this.body = updatedBlock.body;
				this.renderedHtml = updatedBlock.html;
				this.unreadCount = updatedBlock.unreadCount || 0;
				this.commentsCount = updatedBlock.commentsCount || 0;
			});
		},
		raiseOpenEdit(blockId, callback) {
			this.eventBus.emit('open-edit', {blockId, callback});
		},
		promptDeleteBlock() {
			this.eventBus.emit('prompt-delete', this.block.id);
		},
		enterChooseAddPosition() {
			this.choosingAddPosition = true;
		},
		cancelChooseAddPosition() {
			this.choosingAddPosition = false;
		},
		addBeforeThis() {
			let parentId = this.getParentId();
			let insertBeforeId = this.block.id;
			// use style type of this block for default
			this.raisePromptAddSubblock(parentId, insertBeforeId, this.styleType);
		},
		addIntoThis() {
			let parentId = this.block.id;
			let insertBeforeId = null;
			// allow edit-block-modal to determine initial style type
			this.raisePromptAddSubblock(parentId, insertBeforeId, true);
		},
		addAfterThis() {
			let parentId = this.getParentId();
			let insertBeforeId = this.getFollowingBlockId();
			// use style type of this block for default
			this.raisePromptAddSubblock(parentId, insertBeforeId, this.styleType);
		},
		raisePromptAddSubblock(parentId, insertBeforeId, defaultStyleType) {
			this.eventBus.emit('prompt-add-subblock', {parentId, insertBeforeId, defaultStyleType});
		},
		startMoving() {
			this.eventBus.emit('start-moving', this.block.id);
			// Mouseover state is lost without triggering mouseleave
			this.focusActions = false;
		},
		addThisToMovingBlocks() {
			this.eventBus.emit('add-to-moving', this.block.id);
		},
		removeThisFromMovingBlocks() {
			this.eventBus.emit('remove-from-moving', this.block.id);
		},
		promptNavSpec() {
			this.eventBus.emit('prompt-nav-spec');
		},
		cancelMoving() {
			this.eventBus.emit('cancel-moving', this.block.id);
		},
		moveBeforeThis() {
			this.eventBus.emit('move-before', this.block.id);
		},
		moveIntoThis() {
			this.eventBus.emit('move-into', this.block.id);
		},
		moveAfterThis() {
			this.eventBus.emit('move-after', this.block.id);
		},
		mouseLeaveLayover() {
			this.choosingAddPosition = false;
			this.focusActions = false;
		},
		urlUpdated(updatedURLObject) {
			if (this.refType === REF_TYPE_URL && updatedURLObject.id === this.refId) {
				this.refItem = updatedURLObject;
			}
		},
		urlDeleted(refId) {
			if (this.refType === REF_TYPE_URL && refId === this.refId) {
				this.refItem = null;
			}
		},
		raisePlayVideo(urlObject) {
			this.eventBus.emit('play-video', urlObject);
		},
		openCommunity() {
			this.eventBus.emit('open-community', {
				blockId: this.block.id,
				onAdjustUnread: adjustUnreadCount => {
					this.unreadCount += adjustUnreadCount;
				},
				onAdjustComments: adjustCommentsCount => {
					this.commentsCount += adjustCommentsCount;
				},
			});
		},
		updateHasSubblocks() {
			this.hasSubblocks = this.$refs.sublist.childElementCount > 0;
		},
		touchstart(e) {
			this.lastTouchStartTime = Date.now();
		},
		selectstart(e) {
			if (!this.lastTouchStartTime) {
				return;
			}
			if ((Date.now() - this.lastTouchStartTime) < TOUCH_DELAY_CLEAR_SELECTION) {
				// disable select text on click block;
				// text selection should still be possible through long press
				e.preventDefault();
				e.stopPropagation();
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

.spec-block {
	padding-top: $spec-block-margin;
	padding-bottom: $spec-block-margin;

	&:not(:first-child) {
		border-top: thin solid #eee;
	}

	// Remove dividers between simple subsequent blocks of the same type
	&.title-only + .title-only,
	&.ref-item-only + .ref-item-only {
		// Account for $spec-block-margin applied twice as padding
		margin-top: #{(-2 * $spec-block-margin) + $spec-block-title-only-spacing};
		border-top: none !important;
	}

	>.content {
		position: relative;
		min-height: 20px;

		>.bg {
			display: none;
			z-index: -1;
			background-color: $shadow-bg;

			position: absolute;

			left: #{- $spec-block-shadow-offset-left};
			right: #{- $spec-block-shadow-offset};
			top: #{- $spec-block-shadow-offset};
			bottom: #{- $spec-block-shadow-offset};
		}

		&:hover {
			>.bg {
				display: block;
			}
		}

		&.clear-ref {
			>.layover {
				margin-bottom: 5px;
			}

			>.ref-item {
				clear: both;
			}
		}

		@include mobile {
			&.mobile-adjust {

				>.layover {
					margin-bottom: 5px;
				}

				>.ref-item {
					clear: both;
				}
			}
		}

		>.layover {
			float: right;
			margin-left: 10px;
			// user-select: none; // Don't include in text selection

			>div {
				display: inline-flex;
				flex-direction: row;
				flex-wrap: nowrap;
				align-items: center;

				&.expand-control {
					display: none;

					@include mobile {
						display: block;

						&.hide {
							display: none;
						}
					}
				}

				&.actions {

					>.move-action {
						// Move button hidden by default
						display: none;
					}

					@include mobile {
						display: none;

						&.show {
							display: block;
						}

						>.move-action {
							// Show move button in small viewport
							display: inline-block;
						}

						>.drag-handle {
							// Don't show drag handle in small viewport
							display: none;
						}
					}
				} // &.actions

				>.el-button+.el-button {
					margin-left: 5px;
				}

				>.el-checkbox {
					display: inline-block;
					margin-left: 10px; // distance from adjacent button
					height: unset;
				}

				>.drag-handle {
					display: inline-block;
					padding: 3px;
					font-size: 12px;
					border: 1px solid transparent;
					margin-left: 5px;
					vertical-align: middle;
					cursor: ns-resize;
				}
			}
		} // .layover

		>.parent-moving-layover {
			display: none;
		}

		>.title {
			font-weight: bold;
		}

		>.body {
			&.plain {
				white-space: pre-wrap;
			}
			// &.markdown {
			// }
		}

		>.el-alert {
			width: unset;
		}

		// add vertical spacing between elements of block content,
		// skipping layovers such as .parent-moving-layover
		>.layover ~ *:not(.layover) + * {
			margin-top: 10px;
		}
	} // .content

	// &:hover {
	// 	>.content {
	// 		>.bg {
	// 			display: block;
	// 		}
	// 	}
	// }

	>ul.spec-block-list {
		margin-top: $spec-block-margin;

		&.moving {
			.layover {
				// Hide controls on subblocks of block being moved
				display: none;
			}
			.parent-moving-layover {
				// show this layover when a parent is selected for move
				display: block;
			}
		}
	}

	>button {
		margin-top: 10px;
	}
}
</style>
