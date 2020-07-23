<template>
<div v-if="spec || subspec" class="spec-view">

	<ul ref="list" class="spec-block-list" :class="{dragging}">
		<!-- managed programatically -->
	</ul>

	<el-button @click="promptAddBlock()" size="small" type="primary">Add block</el-button>

	<ul ref="mirrorList" class="mirror-list">
		<!-- holds mirror element when dragging block -->
	</ul>

	<edit-block-modal
		ref="editBlockModal"
		:spec-id="specId"
		:subspec-id="subspecId"
		@open-edit-url="openEditUrl"
		/>

	<edit-url-modal
		ref="editUrlModal"
		:spec-id="specId"
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
import {ajaxDeleteBlock, ajaxMoveBlock} from './ajax.js';
import store from '../store.js';
import router from '../router.js';
import {startAutoscroll} from '../utils.js';

const SpecBlockClass = Vue.extend(SpecBlock);

export default {
	components: {
		EditBlockModal,
		EditUrlModal,
	},
	props: {
		spec: Object,
		subspec: Object,
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
			const insertSubblocks = ($parentBlock, subblocks) => {
				for (var i = 0; i < subblocks.length; i++) {
					let subblock = subblocks[i];
					let $subblock = this.insertBlock(subblock, false);
					$('>ul.spec-block-list', $parentBlock).append($subblock);
					if (subblock.subblocks) {
						insertSubblocks($subblock, subblock.subblocks);
					}
				}
			};

			for (var i = 0; i < blocks.length; i++) {
				let block = blocks[i];
				let $block = this.insertBlock(block);
				if (block.subblocks) {
					insertSubblocks($block, block.subblocks);
				}
			}
		}

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
			this.dragTransitionScrollFix();
		}).on('dragend', (el) => {
			this.$store.commit('endDragging');
			this.dragTransitionScrollFix();
		}).on('drop', (el, target, source, sibling) => {
			let $parentBlock = $(target).closest('[data-spec-block]');
			let parentId = $parentBlock.length ? $parentBlock.data('vc').getBlockId() : null;
			let insertBeforeId = sibling ? $(sibling).data('vc').getBlockId() : null;
			// TODO Revert on error (how?)
			ajaxMoveBlock($(el).data('vc').getBlockId(), null, parentId, insertBeforeId);
		});
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
		dragTransitionScrollFix() {
			// Retain scroll position relative to first visible block
			let windowTop = $(window).scrollTop();
			$('[data-spec-block]', this.$refs.list).each((i, e) => {
				let offset = $(e).offset();
				if (offset.top > windowTop) {
					let diff = offset.top - windowTop;
					this.$nextTick(() => {
						// Restore relative scroll position after empty drop zones appear or disappear
						$(window).scrollTop($(e).offset().top - diff);
					});
					return false; // exit loop
				}
			});
		},
		insertBlock(block, append = true) {
			let vc = new SpecBlockClass({
				store,
				router,
				propsData: {
					block,
					eventBus: this.eventBus,
				},
			}).$mount();

			vc.$on('open-edit', this.openEdit);
			vc.$on('prompt-add-subblock', this.promptAddSubblock);
			vc.$on('prompt-delete-block', this.promptDeleteBlock);

			let $vc = $(vc.$el).data('vc', vc);
			if (append) {
				$vc.appendTo(this.$refs.list);
			}
			return $vc;
		},
		promptAddBlock() {
			this.$refs.editBlockModal.showAdd(null, null, this.insertBlock);
		},
		openEdit(block, callback) {
			this.$refs.editBlockModal.showEdit(block, callback);
		},
		promptAddSubblock(parentId, insertBeforeId) {
			this.$refs.editBlockModal.showAdd(parentId, insertBeforeId, newBlock => {
				let $vc = this.insertBlock(newBlock, false);
				// Add to sublist
				if (insertBeforeId) {
					$vc.insertBefore('[data-spec-block="'+insertBeforeId+'"]');
				} else if (parentId) {
					$vc.appendTo('[data-spec-block="'+parentId+'"]>ul.spec-block-list');
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
				ajaxDeleteBlock(blockId).then(callback);
			}).catch(() => { /* Cancelled */ });
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
	},
};
</script>

<style lang="scss">
$spec-block-list-indent: 3em;
$spec-block-point-offset: 1em;

.spec-view {

	// Style for spec-block-list within spec-view or spec-block
	ul.spec-block-list {
		position: relative;
		padding-left: $spec-block-list-indent;
		list-style-type: none;
		counter-reset: spec-block-list-item-number 0;

		&:empty {
			display: none;
		}

		&.dragging {
			// Make all spec-block-list elements non-zero width during drag
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
				left: 0;
				width: $spec-block-list-indent - $spec-block-point-offset;
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
}
</style>
