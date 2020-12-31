<template>
<div v-if="spec || subspec" class="spec-view">

	<ul ref="list" class="spec-block-list" :class="{dragging}">
		<!-- managed programatically -->
	</ul>

	<template v-if="enableEditing">

		<el-button
			v-if="choosingAddPosition"
			@click="placeBlockBottom()"
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

	<play-video-modal
		ref="playVideoModal"
		/>

</div>
</template>

<script>
import $ from 'jquery';
import Vue from 'vue';
import Dragula from 'dragula';
import SpecBlock from './block.vue';
import EditBlockModal from './edit-block-modal.vue';
import EditUrlModal from './edit-url-modal.vue';
import PlayVideoModal from './play-video-modal.vue';
import {ajaxDeleteBlock, ajaxMoveBlock} from './ajax.js';
import store from '../store.js';
import router from '../router.js';
import {alertError, startAutoscroll} from '../utils.js';

const SpecBlockClass = Vue.extend(SpecBlock);

/*
transitRelativeScroll
transit - carry across a transition "in transit"
apply same relative condition from present to next state
*/

export default {
	components: {
		EditBlockModal,
		EditUrlModal,
		PlayVideoModal,
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
			return !!this.$store.state.movingBlockId;
		},
	},
	watch: {
		dragging(moving) {
			if (moving) {
				this.autoscroller = startAutoscroll();
			} else if (this.autoscroller) {
				this.autoscroller.stop();
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

		if (this.enableEditing) {
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
				ajaxMoveBlock($(el).data('vc').getBlockId(), this.subspecId, parentId, insertBeforeId);

			});
		}
	},
	beforeDestroy() {
		if (this.drake) {
			this.drake.destroy();
			this.drake = null;
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
				let offset = $(e).offset();
				if (offset.top > windowTop) {
					let diff = offset.top - windowTop;
					this.$nextTick(() => {
						// Restore relative scroll position after interface transition
						$(window).scrollTop($(e).offset().top - diff);
					});
					return false; // exit loop
				}
			});
		},
		transitRelativeScroll(blockId) {
			// Retain scroll position relative to specified block
			let windowTop = $(window).scrollTop();
			let $block = $('[data-spec-block="' + blockId + '"]', this.$refs.list);
			let offset = $block.offset();
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

			vc.$on('open-edit', this.openEdit);
			vc.$on('prompt-add-subblock', this.promptAddSubblock);
			vc.$on('prompt-delete', this.promptDeleteBlock);
			vc.$on('start-moving', this.startMovingBlock);
			vc.$on('end-moving', this.endMovingBlock);
			vc.$on('play-video', this.playVideo);
			vc.$on('prompt-nav-spec', this.promptNavSpec);

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
		promptAddBlock() {
			this.$refs.editBlockModal.showAdd(null, null, newBlock => {
				this.insertBlock(newBlock, true, true);
			});
		},
		openEdit(block, callback) {
			this.$refs.editBlockModal.showEdit(block, callback);
		},
		promptAddSubblock(parentId, insertBeforeId) {
			this.$refs.editBlockModal.showAdd(parentId, insertBeforeId, newBlock => {
				let $vc = this.insertBlock(newBlock, false, true);
				// Add to sublist
				if (insertBeforeId) {
					$vc.insertBefore('[data-spec-block="'+insertBeforeId+'"]');
				} else if (parentId) {
					let $parentBlock = $('[data-spec-block="'+parentId+'"]');
					$vc.appendTo($parentBlock.find('>ul.spec-block-list'));
					$parentBlock.data('vc').updateHasSubblocks();
				} else {
					$vc.appendTo(this.$refs.list);
				}
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
				}).fail(alertError);
			}).catch(() => { /* Cancelled */ });
		},
		startMovingBlock(blockId) {
			this.$store.commit('startMovingBlock', blockId);
			this.transitRelativeScroll(blockId);
		},
		placeBlockBottom() {
			let movingId = this.$store.state.movingBlockId;
			if (!movingId) {
				return;
			}
			ajaxMoveBlock(movingId, this.subspecId, null, null).then((block = null) => {
				if (block) {
					// moved here from another context
					this.insertBlock(block, true, false);
				} else {
					// moved within current context
					let $moving = $('[data-spec-block="'+movingId+'"]');
					let $sourceParentBlock = $moving.closest('.spec-block-list').closest('[data-spec-block]');
					$moving.appendTo(this.$refs.list);
					if ($sourceParentBlock.length) {
						$sourceParentBlock.data('vc').updateHasSubblocks();
					}
				}
				this.$store.commit('endMovingBlock');
			});
		},
		endMovingBlock(endFromBlockId) {
			this.$store.commit('endMovingBlock');
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
			this.$refs.playVideoModal.show(urlObject);
		},
		promptNavSpec() {
			this.$emit('prompt-nav-spec');
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

		>li.spec-block.none {
		}
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
