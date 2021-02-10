<template>
<ul :data-block-preview-id="blockId" class="spec-block-preview community-target-preview">
	<li>

		<div v-if="hasTitle" class="title">{{block.title}}</div>

		<template v-if="hasRefItem">
			<ref-url
				v-if="block.refType === REF_TYPE_URL"
				:item="block.refItem"
				@play="raisePlayVideo(block.refItem)"
				class="ref-item"
				/>
			<ref-subspec
				v-else-if="block.refType === REF_TYPE_SUBSPEC"
				:item="block.refItem"
				class="ref-item"
				/>
		</template>

		<el-alert v-else-if="refItemMissing"
			title="Content unavailable"
			:closable="false"
			type="warning"/>

		<template v-if="hasBody">
			<div v-if="block.contentType === CONTENT_TYPE_PLAIN"
				class="body plain" v-text="block.body"></div>
			<div v-else-if="block.contentType === CONTENT_TYPE_MARKDOWN"
				class="body markdown" v-html="block.html"></div>
		</template>

		<dynamic-stylesheet :rules="blockPreviewRules"/>

	</li>
</ul>
</template>

<script>
import $ from 'jquery';
import RefUrl from './ref-url.vue';
import RefSubspec from './ref-subspec.vue';
import DynamicStylesheet from '@macu/dynamic-stylesheet-vue/index.js';
import {
	CONTENT_TYPE_PLAIN, CONTENT_TYPE_MARKDOWN,
	REF_TYPE_URL, REF_TYPE_SUBSPEC,
} from './const.js';

export default {
	components: {
		RefUrl,
		RefSubspec,
		DynamicStylesheet,
	},
	props: {
		block: Object,
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
		blockId() {
			return parseInt(this.block.id, 10) || 0;
		},
		blockNumber() {
			return (this.block.styleType === 'numbered' && parseInt(this.block.number, 10)) || 0;
		},
		blockPreviewRules() {
			let beforeContent;
			switch (this.block.styleType) {
				case 'bullet':
					beforeContent = "'\\2022'"; // bullet
				break;
				case 'numbered':
					beforeContent = "'"+this.blockNumber+".'";
				break;
				default:
					return null;
			}
			return {
				['[data-block-preview-id="'+this.blockId+'"]']: {
					'>li:before': {
						'content': beforeContent,
					},
				},
			};
		},
		hasTitle() {
			return !!(this.block.title && this.block.title.trim());
		},
		hasRefItem() {
			return !!(this.block.refType && this.block.refItem);
		},
		hasBody() {
			return !!(
				(this.block.body && this.block.body.trim()) ||
				(this.block.html && this.block.html.trim())
			);
		},
		refItemMissing() {
			return !!this.block.refType && !this.block.refItem;
		},
	},
	methods: {
		raisePlayVideo(urlObject) {
			this.$emit('play-video', urlObject);
		},
	},
};
</script>

<style lang="scss">
@import '../_styles/_spec-view.scss';
@import '../_styles/_colours.scss';

ul.spec-block-preview {
	position: relative;
	margin: 0;
	padding-left: $spec-block-list-padding-left;
	list-style-type: none;
	border: medium solid $block-bg;
	border-radius: 8px;
	background-color: $shadow-bg;

	>li {
		padding: 20px 20px 20px 0; // ul padding-left includes space after bullet

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

		>* ~ * {
			margin-top: 10px;
		}

		&:before {
			display: block;
			position: absolute;
			left: $spec-block-margin;
			width: $spec-block-before-width;
			text-align: right;
		}
	}
}
</style>
