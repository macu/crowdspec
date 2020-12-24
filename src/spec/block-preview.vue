<template>
<div class="spec-block-preview">

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

</div>
</template>

<script>
import $ from 'jquery';
import RefUrl from './ref-url.vue';
import RefSubspec from './ref-subspec.vue';
import {REF_TYPE_URL, REF_TYPE_SUBSPEC} from './const.js';

export default {
	components: {
		RefUrl,
		RefSubspec,
	},
	props: {
		block: Object,
	},
	computed: {
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
@import '../_styles/_colours.scss';

.spec-block-preview {
	padding: 20px;
	background-color: lighten($shadow-bg, 50%);

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

}
</style>
