<template>
<div v-if="spec || subspec" class="spec-view">

	<ul ref="list" class="spec-block-list" :class="{dragging}">
		<!-- managed programatically -->
	</ul>

	<template v-if="enableEditing">

		<el-button
			v-if="choosingAddPosition"
			@click="moveBlocksToBottom()"
			size="small"
			type="success"
			icon="el-icon-top">Move block here</el-button>
		<el-button
			v-else
			@click="promptAddBlock()"
			size="small"
			type="primary">Add block</el-button>

		<ul ref="mirrorList" class="mirror-list">
			<!-- holds mirror element when dragging block -->
		</ul>

		<edit-block-modal
			ref="editBlockModal"
			:spec-id="specId"
			:subspec-id="subspecId"
			@open-edit-url="openEditUrl"
			@play-video="playVideo"
			@prompt-delete="promptDeleteBlock"
			/>

		<edit-url-modal
			ref="editUrlModal"
			:spec-id="specId"
			/>

	</template>

</div>
</template>

<script>
import $ from 'jquery';
import Vue from 'vue';
import SpecBlock from './block.vue';
import EditBlockModal from './edit-block-modal.vue';
import EditUrlModal from './edit-url-modal.vue';
import {ajaxDeleteBlock, ajaxMoveBlocks} from './ajax.js';
import {TARGET_TYPE_BLOCK} from './const.js';
import store from '../store.js';
import router from '../router.js';
import {idsEq, startAutoscroll} from '../utils.js';
import {SCRIPT_DRAGULA, loadScript} from '../widgets/script-loader.js';

const SpecBlockClass = Vue.extend(SpecBlock);

const CONTENT_MIN_OFFSET_TOP = 100; // distance top of content should be within window from top
const CONTENT_MIN_OFFSET_BOTTOM = 200; // distance top of content should be within window from bottom

/*
transitRelativeScroll
transit - carry across a transition "in transit"
apply same relative condition from present to next state
*/

export default {
	components: {
		EditBlockModal,
		EditUrlModal,
	},
	props: {
		spec: Object,
		subspec: Object,
		enableEditing: Boolean,
	},
	data() {
		return {
			autoscroller: null,
			eventBus: new Vue(),
		};
	},
	computed: {
		dragging() {
			return this.$store.state.dragging;
		},
		specId() {
			return this.spec ? this.spec.id
				: (this.subspec ? this.subspec.specId : null);
		},
		subspecId() {
			return this.subspec ? this.subspec.id : null;
		},
		choosingAddPosition() {
			return this.$store.getters.currentlyMovingBlocks;
		},
		dragAndDropEnabled() { // drag-and-drop blocks
			return this.enableEditing && !this.$store.getters.mobileViewport;
		},
	},
	watch: {
		dragAndDropEnabled(enabled) {
			if (enabled) {
				this.mountDragAndDrop();
			} else if (this.drake) {
				this.drake.destroy();
				this.drake = null;
			}
		},
		dragging(moving) {
			if (moving) {
				this.autoscroller = startAutoscroll();
			} else if (this.autoscroller) {
				this.autoscroller.stop();
				this.autoscroller = null;
			}
		},
	},
	mounted() {
		if (!(this.spec || this.subspec)) {
			throw 'spec or subspec param required';
		} else if (this.spec && this.subspec) {
			throw 'both spec and subspec provided; exactly one required';
		}

		let blocks;
		if (this.spec) {
			blocks = this.spec.blocks;
		} else if (this.subspec) {
			blocks = this.subspec.blocks;
		} else {
			return;
		}

		if (blocks) {
			for (var i = 0; i < blocks.length; i++) {
				this.insertBlock(blocks[i]);
			}
		}

		this.$nextTick(() => {
			this.$emit('rendered');
		});

		this.mountDragAndDrop();
	},
	beforeDestroy() {
		if (this.drake) {
			this.drake.destroy();
			this.drake = null;
		}

		if (this.autoscroller) {
			this.autoscroller.stop();
			this.autoscroller = null;
		}

		// If moving blocks carries to another context,
		// ensure only parent IDs appear in the correct order.
		if (
			this.$store.getters.currentlyMovingBlocks &&
			this.$store.state.movingBlocksSourceSubspecId === this.subspecId
		) {
			this.$store.commit('setMovingBlocks', {
				subspecId: this.subspecId,
				blockIds: this.getFinalMovingBlockIds(),
			});
		}

		// Destroy bus
		this.eventBus.$destroy();
		this.eventBus = null;

		// Clean up all independent block component VMs
		$('[data-spec-block]', this.$refs.list).each((i, e) => {
			$(e).data('vc').$destroy();
		});
	},
	methods: {
		transitRelativeScrollFirst() {
			// Retain scroll position relative to first visible block
			let windowTop = $(window).scrollTop();
			$('[data-spec-block]', this.$refs.list).each((i, e) => {
				let $e = $(e);
				let offset = $e.offset(); // offset before DOM update
				if (offset.top > windowTop) {
					let diff = offset.top - windowTop;
					this.$nextTick(() => {
						// Restore relative scroll position after interface transition
						$(window).scrollTop($e.offset().top - diff);
					});
					return false; // exit loop
				}
			});
		},
		transitRelativeScroll(blockId) {
			// Retain scroll position relative to specified block
			let windowTop = $(window).scrollTop();
			let $block = $('[data-spec-block="' + blockId + '"]', this.$refs.list);
			let offset = $block.offset(); // offset before DOM update
			let diff = offset.top - windowTop;
			this.$nextTick(() => {
				// Restore relative scroll position after interface transition
				$(window).scrollTop($block.offset().top - diff);
			});
		},
		insertBlock(block, append = true, justAdded = false) {
			let vc = new SpecBlockClass({
				parent: this, // allows Vue devtools to detect instances
				store,
				router,
				propsData: {
					block,
					subspecId: this.subspecId,
					eventBus: this.eventBus,
					enableEditing: this.enableEditing,
					justAdded,
				},
			}).$mount();

			vc.$on('open-community', this.openBlockCommunity);
			vc.$on('open-edit', this.openEditBlock);
			vc.$on('prompt-add-subblock', this.promptAddSubblock);
			vc.$on('prompt-delete', this.promptDeleteBlock);
			vc.$on('start-moving', this.startMovingBlocks);
			vc.$on('add-to-moving', this.addToMovingBlocks);
			vc.$on('remove-from-moving', this.removeFromMovingBlocks);
			vc.$on('prompt-nav-spec', this.promptNavSpec);
			vc.$on('move-before', this.moveBeforeBlock);
			vc.$on('move-into', this.moveIntoBlock);
			vc.$on('move-after', this.moveAfterBlock);
			vc.$on('cancel-moving', this.cancelMovingBlocks);
			vc.$on('play-video', this.playVideo);

			let $vc = $(vc.$el).data('vc', vc);

			if (block.subblocks) {
				for (var i = 0; i < block.subblocks.length; i++) {
					let subblock = block.subblocks[i];
					let $subblock = this.insertBlock(subblock, false);
					$('>ul.spec-block-list', $vc).append($subblock);
				}
			}

			if (append) {
				$vc.appendTo(this.$refs.list);
			}

			return $vc;
		},
		mountDragAndDrop() {
			if (!this.dragAndDropEnabled) {
				return;
			}
			if (this.drake) {
				return;
			}
			loadScript(SCRIPT_DRAGULA).then(Dragula => {
				console.debug('create drake');
				this.drake = Dragula({
					isContainer(el) {
						return $(el).is('.spec-block-list');
					},
					accepts(el, target, source, sibling) {
						// Don't allow dropping in the transit node
						return !$(target).closest('.gu-transit').length;
					},
					moves(el, source, handle, sibling) {
						return $(handle).is('.drag-handle');
					},
					// revertOnSpill: true,
					mirrorContainer: this.$refs.mirrorList,
				}).on('drag', (el, source) => {
					this.$store.commit('startDragging');
					this.transitRelativeScroll($(el).attr('data-spec-block'));
				}).on('dragend', (el) => {
					this.$store.commit('endDragging');
					this.transitRelativeScroll($(el).attr('data-spec-block'));
				}).on('drop', (el, target, source, sibling) => {
					let blockId = $(el).data('vc').getBlockId();
					let $sourceParentBlock = $(source).closest('[data-spec-block]');
					if ($sourceParentBlock.length) {
						$sourceParentBlock.data('vc').updateHasSubblocks();
					}
					let parentId = null;
					let $parentBlock = $(target).closest('[data-spec-block]');
					if ($parentBlock.length) {
						let parentVc = $parentBlock.data('vc');
						parentId = parentVc.getBlockId();
						parentVc.updateHasSubblocks();
					}
					let insertBeforeId = sibling ? $(sibling).data('vc').getBlockId() : null;
					// TODO Revert on error?
					ajaxMoveBlocks([blockId], this.subspecId, parentId, insertBeforeId);
				});
			});
		},
		promptAddBlock() {
			this.$refs.editBlockModal.showAddBlock(null, null, true, newBlock => {
				let $block = this.insertBlock(newBlock, true, true);
				this.panToBlock($block);
			});
		},
		openBlockCommunity(blockId, onAdjustUnread, onAdjustComments) {
			this.$emit('open-community', TARGET_TYPE_BLOCK, blockId, onAdjustUnread, onAdjustComments);
		},
		openEditBlock(blockId, callback) {
			this.$refs.editBlockModal.showEditBlock(blockId, callback);
		},
		promptAddSubblock(parentId, insertBeforeId, defaultStyleType) {
			this.$refs.editBlockModal.showAddBlock(parentId, insertBeforeId, defaultStyleType, newBlock => {
				let $block = this.insertBlock(newBlock, false, true);
				// Add to sublist
				if (insertBeforeId) {
					$block.insertBefore('[data-spec-block="'+insertBeforeId+'"]');
				} else if (parentId) {
					let $parentBlock = $('[data-spec-block="'+parentId+'"]');
					$block.appendTo($parentBlock.find('>ul.spec-block-list'));
					$parentBlock.data('vc').updateHasSubblocks();
				} else {
					$block.appendTo(this.$refs.list);
				}
				this.panToBlock($block);
			});
		},
		promptDeleteBlock(blockId, callback) {
			this.$confirm('Delete block and subblocks?', 'Confirm', {
				confirmButtonText: 'Delete',
				cancelButtonText: 'Cancel',
				type: 'warning',
			}).then(() => {
				ajaxDeleteBlock(blockId).then(() => {
					$('[data-spec-block="' + blockId + '"]').remove();
					if (callback) {
						callback();
					}
				});
			}).catch(() => { /* Cancelled */ });
		},
		startMovingBlocks(blockId) {
			this.$store.commit('setMovingBlocks', {
				subspecId: this.subspecId,
				blockIds: [blockId],
			});
			this.transitRelativeScroll(blockId);
		},
		addToMovingBlocks(blockId) {
			this.$store.commit('setMovingBlocks', {
				subspecId: this.subspecId,
				blockIds: this.$store.state.movingBlockIds.concat(blockId),
			});
		},
		removeFromMovingBlocks(blockId) {
			this.$store.commit('setMovingBlocks', {
				subspecId: this.subspecId,
				blockIds: this.$store.state.movingBlockIds.filter(v => !idsEq(v, blockId)),
			});
		},
		moveBeforeBlock(blockId) {
			let $target = $('[data-spec-block="'+blockId+'"]', this.$refs.list);
			let $parent = $target.parent().closest('[data-spec-block]');
			let parentId = $parent.length ? $parent.attr('data-spec-block') : null;
			let insertBeforeId = blockId;

			this.moveBlocks(parentId, insertBeforeId, $block => {
				$block.insertBefore($target);
			});

			// this block's parent already has subblocks so no need to update
		},
		moveIntoBlock(blockId) {
			let $target = $('[data-spec-block="'+blockId+'"]', this.$refs.list);
			let parentId = $target.attr('data-spec-block'); // Add under this
			let insertBeforeId = null; // Add at end

			let $targetSublist = $target.find('>ul.spec-block-list');
			this.moveBlocks(parentId, insertBeforeId, $block => {
				$block.appendTo($targetSublist);
			});

			$target.data('vc').updateHasSubblocks();
		},
		moveAfterBlock(blockId) {
			let $target = $('[data-spec-block="'+blockId+'"]', this.$refs.list);
			let $parent = $target.parent().closest('[data-spec-block]');
			let parentId = $parent.length ? $parent.attr('data-spec-block') : null;
			let $nextBlock = $target.next('[data-spec-block]');
			let insertBeforeId = $nextBlock.length ? $nextBlock.attr('data-spec-block') : null;

			let $preceedingBlock = $target;
			this.moveBlocks(parentId, insertBeforeId, $block => {
				$block.insertAfter($preceedingBlock);
				$preceedingBlock = $block;
			});

			// this block's parent already has subblocks so no need to update
		},
		moveBlocksToBottom() {
			this.moveBlocks(null, null, $block => {
				$block.appendTo(this.$refs.list);
			});
		},
		getFinalMovingBlockIds() {
			if (!this.$store.getters.currentlyMovingBlocks) {
				return [];
			}
			if (!idsEq(this.$store.state.movingBlocksSourceSubspecId, this.subspecId)) {
				// Final block IDs were set before leaving the original context
				return this.$store.state.movingBlockIds;
			}
			let ids = [];
			// Query blocks checked for move in order of appearance in DOM
			$('[data-moving-block-id]', this.$refs.list).each((i, e) => {
				let $e = $(e);
				// Skip checked sub blocks of checked parents
				if (!$e.closest('.spec-block-list.moving').length) {
					ids.push($e.attr('data-moving-block-id'));
				}
			});
			return ids;
		},
		moveBlocks(parentId, insertBeforeId, placementFunction) {
			let movingIds = this.getFinalMovingBlockIds();
			if (!movingIds.length) {
				return;
			}
			ajaxMoveBlocks(movingIds, this.subspecId, parentId, insertBeforeId).then((blocks = null) => {
				let $firstBlock = null;
				if (blocks) {
					// moved here from another context
					for (var i = 0; i < blocks.length; i++) {
						let $block = this.insertBlock(blocks[i], false, true);
						placementFunction($block);
						if (!$firstBlock) {
							$firstBlock = $block;
						}
					}
				} else {
					// moved within current context
					let parentVcs = {};
					for (let i = 0; i < movingIds.length; i++) {
						let $moving = $('[data-spec-block="'+movingIds[i]+'"]', this.$refs.list);
						let $sourceParentBlock = $moving.parent().closest('[data-spec-block]');
						placementFunction($moving);
						if ($sourceParentBlock.length) {
							parentVcs[$sourceParentBlock.attr('data-spec-block')] = $sourceParentBlock.data('vc');
						}
						if (!$firstBlock) {
							$firstBlock = $moving;
						}
					}
					for (let id in parentVcs) {
						parentVcs[id].updateHasSubblocks();
					}
				}
				this.$store.commit('endMovingBlocks');
				this.panToBlock($firstBlock);
			});
		},
		cancelMovingBlocks(endFromBlockId) {
			this.$store.commit('endMovingBlocks');
			this.transitRelativeScroll(endFromBlockId);
		},
		openEditUrl(urlObject, updated, deleted) {
			this.$refs.editUrlModal.showEdit(urlObject, updatedUrlObject => {
				this.eventBus.$emit('url-updated', updatedUrlObject);
				if (updated) {
					updated(updatedUrlObject);
				}
			}, deletedId => {
				this.eventBus.$emit('url-deleted', deletedId);
				if (deleted) {
					deleted(deletedId);
				}
			});
		},
		playVideo(urlObject) {
			this.$emit('play-video', urlObject);
		},
		promptNavSpec() {
			this.$emit('prompt-nav-spec');
		},
		panToBlock($moving) {
			let $content = $moving.find('>.content');
			let offset = $content.offset();
			let $window = $(window);
			let windowScrollTop = $window.scrollTop();
			let windowHeight = $window.height();
			if (
				// content top appears before acceptable area
				offset.top < (windowScrollTop + CONTENT_MIN_OFFSET_TOP)
			) {
				// scroll to CONTENT_MIN_OFFSET_TOP before content top
				$window.scrollTop(offset.top - CONTENT_MIN_OFFSET_TOP);
			} else if (
				// content top appears after acceptable area
				offset.top > ((windowScrollTop + windowHeight) - CONTENT_MIN_OFFSET_BOTTOM)
			) {
				// scroll content top CONTENT_MIN_OFFSET_BOTTOM into bottom of viewport
				$window.scrollTop(offset.top - (windowHeight - CONTENT_MIN_OFFSET_BOTTOM));
			}
		},
	},
};
</script>

<style lang="scss">
@import '../_styles/_breakpoints.scss';
@import '../_styles/_spec-view.scss';
@import '../_styles/_app.scss';

.spec-view {

	// Root list
	>ul.spec-block-list {
		// add space beyond block shadow right boundary;
		// block list padding-left includes the same space
		padding-right: $spec-block-margin;
	}

	// Style for spec-block-list within spec-view or spec-block
	ul.spec-block-list {
		position: relative;
		padding-left: $spec-block-list-padding-left;
		list-style-type: none;
		counter-reset: spec-block-list-item-number 0;

		&:empty {
			display: none;
		}

		ul.spec-block-list {
			// Sub lists
			@include mobile {
				// back shift
				margin-left: #{- $spec-block-sublist-shift-sm};
			}
			border-left: thin solid lightgray;
		}

		&.dragging {
			// Make all spec-block-list elements non-zero height during drag
			ul.spec-block-list {
				display: block;
				min-height: 40px;
				&:empty {
					border: thin dashed grey;
					background-color: #eef;
				}
			}
			// Don't show empty drop zones within gu-transit
			li.spec-block.gu-transit {
				ul.spec-block-list:empty {
					display: none;
				}
			}
		}

		>li.spec-block.bullet, >li.spec-block.numbered {
			&:before {
				display: block;
				position: absolute;
				left: $spec-block-margin;
				width: $spec-block-before-width;
				text-align: right;
			}
		}

		>li.spec-block.bullet {
			&:before {
				content: '\2022'; // bullet
			}
		}

		>li.spec-block.numbered {
			counter-increment: spec-block-list-item-number;
			&:before {
				content: counter(spec-block-list-item-number) '.';
			}
		}

		// >li.spec-block.none {
		// }
	}

	>ul.mirror-list {
		>li.spec-block {
			list-style-type: none;
			&:before {
				// Don't show disc/numbering
				display: none;
			}
			>ul.spec-block-list {
				// Don't show subpoints on mirror item
				display: none;
			}
		}
	}

	>.el-button {
		margin-top: 2em;
		margin-left: $spec-block-list-padding-left;
	}
}
</style>
