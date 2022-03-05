<template>
<div class="community-context-stack-bar" :class="{[elementType]: true, restricted}">
	<el-button size="small" type="primary" circle>
		<i class="material-icons">north_west</i>
	</el-button>
	<span class="bar">
		<span class="label">
			<username v-if="showUsername"
				:username="element.username"
				:highlight="element.highlight"
				/>
			<template v-else>{{label || '...'}}</template>
		</span>
		<span class="content" v-text="content || '...'"/>
		<el-tag v-if="showPrivate" type="info" effect="dark" size="small">Private</el-tag>
	</span>
</div>
</template>

<script>
import Username from '../widgets/username.vue';
import {
	TARGET_TYPE_SPEC,
	TARGET_TYPE_SUBSPEC,
	TARGET_TYPE_BLOCK,
	TARGET_TYPE_COMMENT,
	REF_TYPE_SUBSPEC,
	REF_TYPE_URL,
} from './const.js';
import {
	ucFirst,
} from '../utils.js';

export default {
	components: {
		Username,
	},
	props: {
		restricted: Boolean,
		element: Object,
		elementType: String,
	},
	computed: {
		showUsername() {
			return this.elementType === TARGET_TYPE_COMMENT;
		},
		label() {
			let label = ucFirst(this.elementType);
			if (this.elementType === TARGET_TYPE_BLOCK) {
				if (this.element.refType) {
					// Show "(URL)" or "(Subspec)", etc.
					label += ' (' + ucFirst(this.element.refType) + ')';
				}
			}
			return label;
		},
		showPrivate() {
			switch (this.elementType) {
				case TARGET_TYPE_SPEC:
					return !this.element.public;
				case TARGET_TYPE_SUBSPEC:
					return this.element.private;
			}
			return false;
		},
		content() {
			switch (this.elementType) {
				case TARGET_TYPE_SPEC:
				case TARGET_TYPE_SUBSPEC:
					return this.element.name;
				case TARGET_TYPE_BLOCK:
					let content;
					if (content = (this.element.title || '').trim()) {
						return content;
					}
					if (content = (this.element.body || '').trim()) {
						return content;
					}
					if (this.element.refItem) {
						switch (this.element.refType) {
							case REF_TYPE_SUBSPEC:
								return this.element.refItem.name;
							case REF_TYPE_URL:
								return this.element.refItem.title || this.element.refItem.url;
						}
					}
					return '';

				case TARGET_TYPE_COMMENT:
					return this.element.body;
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
	margin: 5px;

	&.restricted {
		margin: 0;
		border: 5px solid lightgray;
		background-color: lightgray;
		&:not(:first-child) {
			border-top: none;
		}
		>.el-button {
			visibility: hidden;
		}
	}

	&:not(.restricted) {
		>.bar {
			cursor: pointer;
		}
	}

	>.el-button {
		padding: 5px;
		font-size: 12px;
		margin-right: 10px;
	}

	>.bar {
		flex: 1;
		display: flex;
		flex-direction: row;
		align-items: center;
		font-size: 12px;
		border-radius: 8px;
		overflow: hidden; // enables ellipsis
		background-color: white;

		>.label {
			align-self: stretch;
			padding: 5px 20px;
			font-weight: bold;
			background-color: $shadow-bg;
			border-radius: 8px;
			display: flex;
			flex-direction: row;
			align-items: center;
			white-space: nowrap;
		}
		>.content {
			padding: 5px 20px;
			white-space: nowrap;
			overflow: hidden;
			text-overflow: ellipsis;
		}
		>.el-tag {
			margin-left: -10px;
			margin-right: 20px;
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
