<template>
<li :data-spec-block="block.id" class="spec-block" :class="classes">

	<div class="content">

		<div class="bg"></div>

		<div class="layover" @mouseleave="cancelChooseAddPosition()">
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
				<el-button @click="editBlock()" type="default" size="mini" icon="el-icon-edit" circle/>
				<el-button @click="promptDeleteBlock()" type="warning" size="mini" icon="el-icon-delete" circle/>
				<el-button @click="enterChooseAddPosition()" type="primary" size="mini" icon="el-icon-plus" circle/>
				<i @click="startMoving()" class="el-icon-d-caret drag-handle"></i>
			</template>
		</div>

		<div v-if="title" class="title">{{title}}</div>

		<template v-if="refType && refItem">
			<ref-url v-if="refType === 'url'" :item="refItem" class="ref-item"/>
			<ref-subspec v-else-if="refType === 'subspec'" :item="refItem" class="ref-item"/>
		</template>

		<div v-if="body" class="body">{{body}}</div>

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
import {REF_TYPE_URL} from './const.js';

export default {
	components: {
		RefUrl,
		RefSubspec,
	},
	props: {
		block: Object,
		eventBus: Object,
	},
	data() {
		return {
			// Copy dynamic values
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
		};
	},
	computed: {
		classes() {
			return [this.styleType];
		},
		movingThis() {
			return this.$store.state.moving === this.block.id;
		},
		movingAnother() {
			return this.$store.state.moving && !this.movingThis;
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
				styleType: this.styleType,
				contentType: this.contentType,
				refType: this.refType,
				refId: this.refId,
				refItem: this.refItem,
				title: this.title,
				body: this.body,
			}, updatedBlock => {
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
			this.$store.commit('startMoving', this.block.id);
		},
		cancelMoving() {
			this.$store.commit('endMoving');
		},
		moveBeforeThis() {
			let movingId = this.$store.state.moving;
			let parentId = this.getParentId();
			let insertBeforeId = this.block.id; // Add before this
			ajaxMoveBlock(movingId, null, parentId, insertBeforeId).then(() => {
				$('[data-spec-block="'+movingId+'"]').insertBefore(this.$el);
				this.$store.commit('endMoving');
			});
		},
		moveIntoThis() {
			let movingId = this.$store.state.moving;
			let parentId = this.block.id; // Add under this
			let insertBeforeId = null; // Add at end
			ajaxMoveBlock(movingId, null, parentId, insertBeforeId).then(() => {
				$('[data-spec-block="'+movingId+'"]').appendTo(this.$refs.sublist);
				this.$store.commit('endMoving');
			});
		},
		moveAfterThis() {
			let movingId = this.$store.state.moving;
			let parentId = this.getParentId();
			let insertBeforeId = this.getFollowingBlockId();
			ajaxMoveBlock(movingId, null, parentId, insertBeforeId).then(() => {
				$('[data-spec-block="'+movingId+'"]').insertAfter(this.$el);
				this.$store.commit('endMoving');
			});
		},
		urlUpdated(updatedURLObject) {
			if (this.refType === REF_TYPE_URL && updatedURLObject.id === this.refId) {
				this.refItem = updatedURLObject;
			}
		},
		urlDeleted(refId) {
			if (this.refType === REF_TYPE_URL && refId === this.refId) {
				this.refType = null;
				this.refId = null;
				this.refItem = null;
			}
		},
	},
};
</script>

<style lang="scss">
.spec-block {
	padding-top: 10px;
	padding-bottom: 10px;

	// scroll horizontally to view nested items on small screens for now
	min-width: 300px;

	&:not(:first-child) {
		border-top: thin solid #eee;
	}

	>.content {
		position: relative;
		min-height: 20px;
		>.bg {
			display: none;
			z-index: -1;
			position: absolute;
			left: -3em;
			right: -7px;
			top: -7px;
			bottom: -7px;
			background-color: #ececec;
		}
		&:hover {
			>.bg {
				display: block;
			}
		}
		>.layover {
			float: right;
			margin-left: 10px;
			// user-select: none; // Don't include in text selection
			.el-button {
				padding: 3px;
				font-size: 12px;
			}
			.el-button+.el-button {
				margin-left: 5px;
			}
			.drag-handle {
				display: inline-block;
				padding: 3px;
				font-size: 12px;
				border: 1px solid transparent;
				margin-left: 5px;
				vertical-align: middle;
				cursor: ns-resize;
			}
		}
		>.title {
			font-weight: bold;
		}
		>.body {
			white-space: pre-wrap;
		}
		>.title+.body, >.title+.ref-item, >.ref-item+.body {
			margin-top: 10px;
		}
	}

	>ul.spec-block-list {
		margin-top: 10px;
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
