<template>
<span class="username">
	<i class="el-icon-user" :style="style"></i>
	<span>{{username}}</span>
</span>
</template>

<script>
import {encodeRgb, invertHsl} from '../colour-utils.js';

export default {
	props: {
		username: {
			type: String,
			required: true,
		},
		highlight: String,
	},
	computed: {
		style() {
			if (!this.highlight) {
				return null;
			}
			let inverted = encodeRgb(invertHsl(this.highlight));
			return {
				'color': encodeRgb(this.highlight),
				'text-shadow': '-1px -1px 0 ' + inverted + ', ' +
					'1px -1px 0 ' + inverted + ', ' +
					'-1px 1px 0 ' + inverted + ', ' +
					'1px 1px 0 ' + inverted,
			};
		},
	},
};
</script>

<style lang="scss">
@import '../_styles/_app.scss';

.username {
	>i {
		display: inline-block;
		font-weight: bold;
		margin-right: $icon-spacing;
	}
	>[class*="el-icon-"]+span {
		// Override Element theme
		margin-left: 0;
	}
}
</style>
