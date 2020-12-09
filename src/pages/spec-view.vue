<template>
<spec-view
	v-if="spec"
	:key="spec.renderTime"
	:spec="spec"
	:enable-editing="enableEditing"
	@prompt-nav-spec="promptNavSpec"
	class="spec-view-page"
	/>
</template>

<script>
import SpecView from '../spec/view.vue';

export default {
	props: {
		spec: Object,
		enableEditing: Boolean,
	},
	components: {
		SpecView,
	},
	beforeRouteEnter(to, from, next) {
		console.debug('beforeRouteEnter spec-view', to);
		next();
	},
	beforeRouteUpdate(to, from, next) {
		console.debug('beforeRouteUpdate spec-view', to);
		next();
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

.spec-view-page {
	padding: $content-area-padding;
	padding-top: $content-area-padding / 2;

	@include mobile {
		padding: $content-area-padding-sm;

		margin-left: #{-$content-area-padding-sm};
		margin-right: #{-$content-area-padding-sm};
	}
}
</style>
