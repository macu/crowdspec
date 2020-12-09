<template>
<spec-view
	v-if="subspec"
	:key="subspec.renderTime"
	:subspec="subspec"
	:enable-editing="enableEditing"
	@prompt-nav-spec="promptNavSpec"
	class="subspec-view-page"
	/>
</template>

<script>
import SpecView from '../spec/view.vue';

export default {
	props: {
		subspec: Object,
		enableEditing: Boolean,
	},
	components: {
		SpecView,
	},
	methods: {
		promptNavSpec() {
			this.$emit('prompt-nav-spec');
		},
	},
};
</script>

<style lang="scss">
@import '../styles/_app.scss';
@import '../styles/_breakpoints.scss';

.subspec-view-page {
	padding: $content-area-padding;
	padding-top: $content-area-padding / 2;

	@include mobile {
		padding: $content-area-padding-sm;

		margin-left: #{-$content-area-padding-sm};
		margin-right: #{-$content-area-padding-sm};
	}
}
</style>
