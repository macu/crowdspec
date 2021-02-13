<template>
<div class="community-context-stack-bar" :class="targetType">
	<el-button size="mini" icon="el-icon-top-left" type="primary" circle/>
	<span class="bar">
		<span class="label">{{label || '...'}}</span>
		<span class="content" v-text="content || '...'"/>
	</span>
</div>
</template>

<script>
import {
	TARGET_TYPE_SPEC,
	TARGET_TYPE_SUBSPEC,
	TARGET_TYPE_BLOCK,
	TARGET_TYPE_COMMENT,
	REF_TYPE_SUBSPEC,
	REF_TYPE_VIDEO,
	REF_TYPE_URL,
} from './const.js';
import {
	ucFirst,
} from '../utils.js';

export default {
	props: {
		target: Object,
		targetType: String,
	},
	computed: {
		label() {
			let label = ucFirst(this.targetType);
			if (this.targetType === TARGET_TYPE_BLOCK) {
				if (this.target.refType) {
					label += ' (' + ucFirst(this.target.refType) + ')';
				}
			}
			return label;
		},
		content() {
			switch (this.targetType) {
				case TARGET_TYPE_SPEC:
				case TARGET_TYPE_SUBSPEC:
					return this.target.name;
				case TARGET_TYPE_BLOCK:
					let content;
					if (content = (this.target.title || '').trim()) {
						return content;
					}
					if (content = (this.target.body || '').trim()) {
						return content;
					}
					if (this.target.refItem) {
						switch (this.target.refType) {
							case REF_TYPE_SUBSPEC:
								return this.target.refItem.name;
							case REF_TYPE_VIDEO:
							case REF_TYPE_URL:
								return this.target.refItem.title || this.target.refItem.url;
						}
					}
					return '';

				case TARGET_TYPE_COMMENT:
					return this.target.body;
			}
		},
	},
};
</script>

<style lang="scss">
@import '../_styles/_colours.scss';

.community-context-stack-bar {
	display: flex;
	flex-direction: row;
	align-items: center; // vertical align
	margin: 5px 0;

	>.el-button {
		padding: 5px;
		font-size: 12px;
		margin-right: 10px;
	}

	>.bar {
		flex: 1;
		display: flex;
		flex-direction: row;
		font-size: 12px;
		cursor: pointer;
		border-radius: 8px;
		overflow: hidden; // enables ellipsis

		>.label {
			padding: 5px 20px;
			font-weight: bold;
			background-color: $shadow-bg;
			border-radius: 8px;
		}
		>.content {
			flex: 1;
			padding: 5px 20px;
			white-space: nowrap;
			overflow: hidden;
			text-overflow: ellipsis;
		}
	}

	&.spec>.bar {
		border: 2px solid $spec-bg;
		border-bottom: 2px solid darken($spec-bg, 10%);
		>.label {
			border-right: 2px solid $spec-bg;
		}
	}

	&.subspec>.bar {
		border: 2px solid $subspec-bg;
		border-bottom: 2px solid darken($subspec-bg, 10%);
		>.label {
			border-right: 2px solid $subspec-bg;
		}
	}

	&.block>.bar {
		border: 2px solid $block-bg;
		border-bottom: 2px solid darken($block-bg, 10%);
		>.label {
			border-right: 2px solid $block-bg;
		}
	}

	&.comment>.bar {
		border: 2px solid $comment-bg;
		border-bottom: 2px solid darken($comment-bg, 10%);
		>.label {
			border-right: 2px solid $comment-bg;
		}
	}
}
</style>
