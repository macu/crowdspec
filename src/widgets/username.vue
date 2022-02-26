<template>
<span class="username">
	<i class="material-icons" :style="style">person</i>
	<span>{{username}}</span>
</span>
</template>

<script>
import {encodeRgb, invertHsl} from '../colour-utils.js';

export default {
	name: 'username',
	props: {
		username: {
			type: String,
			required: true,
		},
		highlight: String,
	},
	computed: {
		style() {

			// Use current setting value if current user (in case updated)
			let highlight = this.username === this.$store.getters.username
				? this.$store.getters.userSettings.userProfile.highlightUsername
				: this.highlight;

			if (!highlight) {
				return null;
			}

			let color = encodeRgb(highlight);
			let inverted = encodeRgb(invertHsl(highlight));
			return {
				'color': color,
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
	white-space: nowrap;
	>i {
		display: inline-block;
		font-weight: bold;
	}
	>span {
		display: inline-block;
		margin-left: $icon-spacing;
	}
}
</style>
