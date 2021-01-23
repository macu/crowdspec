<template>
<li :data-spec-block="block.id" class="spec-block" :class="classes" @click.stop.prevent="clearSelection()">

	<div class="content" :class="{'mobile-adjust': mobileAdjust}">

		<div class="bg"></div>

		<div class="layover" @mouseleave="mouseLeaveLayover()">
			<template v-if="enableEditing">
				<div class="expand-control" :class="{hide: showActions}">
					<!-- only show community button to admins in collapsed mobile menu when there are unread submissions -->
					<!-- always show community button to guests, who only make community interactions -->
					<el-button v-if="unreadCount"
						@click="openCommunity()"
						type="primary"
						size="mini"
						icon="el-icon-chat-dot-square">
						{{unreadCount}}
					</el-button>
					<el-button @click="focusActions = true" type="default" size="mini" icon="el-icon-more" circle/>
				</div>
				<div class="actions" :class="{show: showActions}">
					<template v-if="choosingAddPosition">
						<el-button @click="cancelChooseAddPosition()" type="warning" size="mini" icon="el-icon-close" circle/>
						<el-button @click="addBeforeThis()" type="primary" size="mini" icon="el-icon-top" circle/>
						<el-button @click="addIntoThis()" type="primary" size="mini" icon="el-icon-bottom-right" circle/>
						<el-button @click="addAfterThis()" type="primary" size="mini" icon="el-icon-bottom" circle/>
					</template>
					<template v-else-if="movingThis">
						<el-button @click="promptNavSpec()" size="mini" icon="el-icon-folder-add">Change context</el-button>
						<el-button @click="cancelMoving()" type="warning" size="mini" icon="el-icon-close">Cancel move</el-button>
					</template>
					<template v-else-if="movingAnother">
						<el-button @click="cancelMoving()" type="warning" size="mini" icon="el-icon-close" circle/>
						<el-button @click="moveBeforeThis()" type="success" size="mini" icon="el-icon-top" circle/>
						<el-button @click="moveIntoThis()" type="success" size="mini" icon="el-icon-bottom-right" circle/>
						<el-button @click="moveAfterThis()" type="success" size="mini" icon="el-icon-bottom" circle/>
					</template>
					<template v-else>
						<el-button @click="openCommunity()"
							:type="unreadCount ? 'primary' : 'default'"
							size="mini"
							icon="el-icon-chat-dot-square">
							<template v-if="unreadCount">{{unreadCount}}</template>
						</el-button>
						<el-button @click="editBlock()" type="default" size="mini" icon="el-icon-edit" circle/>
						<el-button v-if="showDeleteButton" @click="promptDeleteBlock()" type="warning" size="mini" icon="el-icon-delete" circle/>
						<el-button @click="enterChooseAddPosition()" type="primary" size="mini" icon="el-icon-plus" circle/>
						<el-button @click="startMoving()" class="move-action" type="default" size="mini" icon="el-icon-d-caret" circle/>
						<i @click="startMoving()" class="el-icon-d-caret drag-handle"></i>
					</template>
				</div>
			</template>
			<div v-else class="visitor-actions">
				<el-button
					@click="openCommunity()"
					:type="unreadCount ? 'primary' : 'default'"
					size="mini"
					icon="el-icon-chat-dot-square">
					<template v-if="unreadCount">{{unreadCount}}</template>
				</el-button>
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

		<div v-if="hasBody" class="body">{{body}}</div>

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
import {ajaxMoveBlock} from './ajax.js';
import {REF_TYPE_URL, REF_TYPE_SUBSPEC} from './const.js';

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
			refItem: this.block.refItem,
			unreadCount: this.block.unreadCount || 0,
			subblocks: this.block.subblocks ? this.block.subblocks.slice() : [],

			// Dynamic
			choosingAddPosition: false,
			focusActions: false,
			hasSubblocks: !!(this.block.subblocks && this.block.subblocks.length),
		};
	},
	computed: {
		REF_TYPE_URL() {
			return REF_TYPE_URL;
		},
		REF_TYPE_SUBSPEC() {
			return REF_TYPE_SUBSPEC;
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
			return !!(this.body && this.body.trim());
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
		movingThis() {
			return this.$store.state.movingBlockId === this.block.id;
		},
		movingAnother() {
			return this.$store.state.movingBlockId && !this.movingThis;
		},
		showActions() {
			return this.focusActions || this.movingThis || this.movingAnother;
		},
		mobileAdjust() {
			// whether to add {clear: both} to ref item area
			// (show layover above rather than to right of ref item)
			return this.showActions || !!this.unreadCount;
		},
	},
	mounted() {
		this.eventBus.$on('url-updated', this.urlUpdated);
		this.eventBus.$on('url-deleted', this.urlDeleted);
	},
	beforeDestroy() {
		this.eventBus.$off('url-updated', this.urlUpdated);
		this.eventBus.$off('url-deleted', this.urlDeleted);
	},
	methods: {
		getBlockId() {
			return this.block.id;
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
			this.raiseOpenEdit({
				id: this.block.id,
				created: this.block.created,
				updated: this.updated,
				styleType: this.styleType,
				contentType: this.contentType,
				refType: this.refType,
				refId: this.refId,
				refItem: this.refItem,
				title: this.title,
				body: this.body,
			}, updatedBlock => {
				this.updated = updatedBlock.updated;
				this.styleType = updatedBlock.styleType;
				this.contentType = updatedBlock.contentType;
				this.refType = updatedBlock.refType;
				this.refId = updatedBlock.refId;
				this.refItem = updatedBlock.refItem;
				this.title = updatedBlock.title;
				this.body = updatedBlock.body;
				this.unreadCount = updatedBlock.unreadCount || 0;
			});
		},
		raiseOpenEdit(block, callback) {
			this.$emit('open-edit', block, callback);
		},
		promptDeleteBlock() {
			this.$emit('prompt-delete', this.block.id);
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
			this.raisePromptAddSubblock(parentId, insertBeforeId);
		},
		addIntoThis() {
			let parentId = this.block.id;
			let insertBeforeId = null;
			this.raisePromptAddSubblock(parentId, insertBeforeId);
		},
		addAfterThis() {
			let parentId = this.getParentId();
			let insertBeforeId = this.getFollowingBlockId();
			this.raisePromptAddSubblock(parentId, insertBeforeId);
		},
		raisePromptAddSubblock(parentId, insertBeforeId) {
			this.$emit('prompt-add-subblock', parentId, insertBeforeId);
		},
		startMoving() {
			this.$emit('start-moving', this.block.id);
			// Mouseover state is lost without triggering mouseleave
			this.focusActions = false;
		},
		cancelMoving() {
			this.$emit('end-moving', this.block.id);
		},
		promptNavSpec() {
			this.$emit('prompt-nav-spec');
		},
		moveBeforeThis() {
			let movingId = this.$store.state.movingBlockId;
			let parentId = this.getParentId();
			let insertBeforeId = this.block.id; // Add before this
			ajaxMoveBlock(movingId, this.subspecId, parentId, insertBeforeId).then((block = null) => {
				if (block) {
					// moved here from another context
					let $block = this.$parent.insertBlock(block, false, true);
					$block.insertBefore(this.$el);
				} else {
					// moved within current context
					let $moving = $('[data-spec-block="'+movingId+'"]');
					let $sourceParentBlock = $moving.closest('.spec-block-list').closest('[data-spec-block]');
					$moving.insertBefore(this.$el);
					if ($sourceParentBlock.length) {
						$sourceParentBlock.data('vc').updateHasSubblocks();
					}
					// this block's parent already has subblocks so no need to update
				}
				this.$store.commit('endMovingBlock');
			});
		},
		moveIntoThis() {
			let movingId = this.$store.state.movingBlockId;
			let parentId = this.block.id; // Add under this
			let insertBeforeId = null; // Add at end
			ajaxMoveBlock(movingId, this.subspecId, parentId, insertBeforeId).then((block = null) => {
				if (block) {
					let $block = this.$parent.insertBlock(block, false, true);
					$block.appendTo(this.$refs.sublist);
				} else {
					let $moving = $('[data-spec-block="'+movingId+'"]');
					let $sourceParentBlock = $moving.closest('.spec-block-list').closest('[data-spec-block]');
					$moving.appendTo(this.$refs.sublist);
					if ($sourceParentBlock.length) {
						$sourceParentBlock.data('vc').updateHasSubblocks();
					}
				}
				this.updateHasSubblocks();
				this.$store.commit('endMovingBlock');
			});
		},
		moveAfterThis() {
			let movingId = this.$store.state.movingBlockId;
			let parentId = this.getParentId();
			let insertBeforeId = this.getFollowingBlockId();
			ajaxMoveBlock(movingId, this.subspecId, parentId, insertBeforeId).then((block = null) => {
				if (block) {
					let $block = this.$parent.insertBlock(block, false, true);
					$block.insertAfter(this.$el);
				} else {
					let $moving = $('[data-spec-block="'+movingId+'"]');
					let $sourceParentBlock = $moving.closest('.spec-block-list').closest('[data-spec-block]');
					$moving.insertAfter(this.$el);
					if ($sourceParentBlock.length) {
						$sourceParentBlock.data('vc').updateHasSubblocks();
					}
				}
				this.$store.commit('endMovingBlock');
			});
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
			this.$emit('play-video', urlObject);
		},
		openCommunity() {
			this.$emit('open-community', this.block.id, adjustUnreadCount => {
				this.unreadCount += adjustUnreadCount;
			});
		},
		updateHasSubblocks() {
			this.hasSubblocks = this.$refs.sublist.childElementCount > 0;
		},
		clearSelection() {
			// @click.stop.prevent should prevent selections on tap on mobile;
			// text selection should still be possible through long press;
			// clear selection if any
			let s = window.getSelection(); // IE9+
			if (s) {
				this.$nextTick(() => {
					s.removeAllRanges();
				});
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

				>.el-button {
					padding: 3px;
					font-size: 12px;
				}

				>.el-button+.el-button {
					margin-left: 5px;
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

		>.title {
			font-weight: bold;
		}

		>.body {
			white-space: pre-wrap;
		}

		>.el-alert {
			width: unset;
		}

		>.layover + * ~ * {
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
		}
	}

	>button {
		margin-top: 10px;
	}
}
</style>
