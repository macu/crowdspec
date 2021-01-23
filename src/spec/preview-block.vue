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

		<div v-if="hasBody" class="body">{{block.body}}</div>

		<dynamic-stylesheet :rules="blockPreviewRules"/>

	</li>
</ul>
</template>

<script>
import $ from 'jquery';
import RefUrl from './ref-url.vue';
import RefSubspec from './ref-subspec.vue';
import DynamicStylesheet from '@macu/dynamic-stylesheet-vue/index.js';
import {REF_TYPE_URL, REF_TYPE_SUBSPEC} from './const.js';

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
		REF_TYPE_URL() {
			return REF_TYPE_URL;
		},
		REF_TYPE_SUBSPEC() {
			return REF_TYPE_SUBSPEC;
		},
		hasTitle() {
			return !!(this.block.title && this.block.title.trim());
		},
		hasBody() {
			return !!(this.block.body && this.block.body.trim());
		},
		hasRefItem() {
			return !!(this.block.refType && this.block.refItem);
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
			white-space: pre-wrap;
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
