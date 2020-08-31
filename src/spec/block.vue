<template>
<li :data-spec-block="block.id" class="spec-block" :class="classes">

	<div class="content" :class="{'mobile-adjust': showActions}">

		<div class="bg"></div>

		<div class="layover" @mouseleave="mouseLeaveLayover()">
			<template v-if="enableEditing">
				<div class="expand-control" :class="{hide: showActions}">
					<!-- only show community button to admins in collapsed mobile menu when there are unread submissions -->
					<!-- always show community button to guests, who only make community interactions -->
					<el-button v-if="unreadSubmissionsCount"
						@click="openCommunity()"
						:type="submissionsCount ? 'primary' : 'default'"
						size="mini"
						icon="el-icon-chat-dot-square">
						{{submissionsCount}}
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
						<el-button @click="cancelMoving()" type="warning" size="mini" icon="el-icon-close">Cancel move</el-button>
					</template>
					<template v-else-if="movingAnother">
						<el-button @click="cancelMoving()" type="warning" size="mini" icon="el-icon-close" circle/>
						<el-button @click="moveBeforeThis()" type="primary" size="mini" icon="el-icon-top" circle/>
						<el-button @click="moveIntoThis()" type="primary" size="mini" icon="el-icon-bottom-right" circle/>
						<el-button @click="moveAfterThis()" type="primary" size="mini" icon="el-icon-bottom" circle/>
					</template>
					<template v-else>
						<el-button @click="openCommunity()"
							:type="submissionsCount ? 'primary' : 'default'"
							size="mini"
							icon="el-icon-chat-dot-square">
							{{submissionsCount}}
						</el-button>
						<el-button @click="editBlock()" type="default" size="mini" icon="el-icon-edit" circle/>
						<el-button @click="promptDeleteBlock()" type="warning" size="mini" icon="el-icon-delete" circle/>
						<el-button @click="enterChooseAddPosition()" type="primary" size="mini" icon="el-icon-plus" circle/>
						<el-button @click="startMoving()" class="move-action" type="default" size="mini" icon="el-icon-d-caret" circle/>
						<i @click="startMoving()" class="el-icon-d-caret drag-handle"></i>
					</template>
				</div>
			</template>
			<div v-else class="visitor-actions">
				<el-button
					@click="openCommunity()"
					:type="submissionsCount ? 'primary' : 'default'"
					size="mini"
					icon="el-icon-chat-dot-square">
					{{submissionsCount}}
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
			return this.$store.state.moving === this.block.id;
		},
		movingAnother() {
			return this.$store.state.moving && !this.movingThis;
		},
		showActions() {
			return this.focusActions || this.movingThis || this.movingAnother;
		},
		unreadSubmissionsCount() {
			return 0;
		},
		submissionsCount() {
			return 0;
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
			});
		},
		raiseOpenEdit(block, callback) {
			this.$emit('open-edit', block, callback);
		},
		promptDeleteBlock() {
			this.raisePromptDeleteBlock(this.block.id, () => {
				$(this.$el).remove();
			});
		},
		raisePromptDeleteBlock(blockId, callback) {
			this.$emit('prompt-delete-block', blockId, callback);
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
		moveBeforeThis() {
			let movingId = this.$store.state.moving;
			let parentId = this.getParentId();
			let insertBeforeId = this.block.id; // Add before this
			ajaxMoveBlock(movingId, this.subspecId, parentId, insertBeforeId).then(() => {
				let $moving = $('[data-spec-block="'+movingId+'"]');
				let $sourceParentBlock = $moving.closest('.spec-block-list').closest('[data-spec-block]');
				$moving.insertBefore(this.$el);
				if ($sourceParentBlock.length) {
					$sourceParentBlock.data('vc').updateHasSubblocks();
				}
				this.$store.commit('endMoving');
			});
		},
		moveIntoThis() {
			let movingId = this.$store.state.moving;
			let parentId = this.block.id; // Add under this
			let insertBeforeId = null; // Add at end
			ajaxMoveBlock(movingId, this.subspecId, parentId, insertBeforeId).then(() => {
				let $moving = $('[data-spec-block="'+movingId+'"]');
				let $sourceParentBlock = $moving.closest('.spec-block-list').closest('[data-spec-block]');
				$moving.appendTo(this.$refs.sublist);
				if ($sourceParentBlock.length) {
					$sourceParentBlock.data('vc').updateHasSubblocks();
				}
				this.updateHasSubblocks();
				this.$store.commit('endMoving');
			});
		},
		moveAfterThis() {
			let movingId = this.$store.state.moving;
			let parentId = this.getParentId();
			let insertBeforeId = this.getFollowingBlockId();
			ajaxMoveBlock(movingId, this.subspecId, parentId, insertBeforeId).then(() => {
				let $moving = $('[data-spec-block="'+movingId+'"]');
				let $sourceParentBlock = $moving.closest('.spec-block-list').closest('[data-spec-block]');
				$moving.insertAfter(this.$el);
				if ($sourceParentBlock.length) {
					$sourceParentBlock.data('vc').updateHasSubblocks();
				}
				this.$store.commit('endMoving');
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
			this.$alert('Unimplemented');
		},
		updateHasSubblocks() {
			this.hasSubblocks = this.$refs.sublist.childElementCount > 0;
		},
	},
};
</script>

<style lang="scss">
@import '../styles/_breakpoints.scss';
@import '../styles/_colours.scss';
@import '../styles/_spec-view.scss';
@import '../styles/_app.scss';

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
