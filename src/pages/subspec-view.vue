<template>
<div class="subspec-view-page" :class="{loading}">
	<p v-if="loading">Loading...</p>
	<spec-view
		v-else-if="subspec"
		:key="subspec.renderTime"
		:subspec="subspec"
		:enable-editing="enableEditing"
		@prompt-nav-spec="promptNavSpec"
		@rendered="$emit('rendered')"
		/>
</div>
</template>

<script>
import SpecView from '../spec/view.vue';

export default {
	props: {
		loading: Boolean,
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
		padding: $content-area-padding-sm 0;

		&.loading {
			padding: $content-area-padding-sm;
		}
	}
}
</style>
