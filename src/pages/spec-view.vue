<template>
<div class="spec-view-page" :class="{loading}">
	<p v-if="loading">Loading...</p>
	<spec-view
		v-else-if="spec"
		:key="spec.renderTime"
		:spec="spec"
		:enable-editing="enableEditing"
		@prompt-nav-spec="promptNavSpec"
		@open-community="openCommunity"
		@play-video="playVideo"
		@rendered="$emit('rendered')"
		/>
</div>
</template>

<script>
import SpecView from '../spec/view.vue';

export default {
	props: {
		loading: Boolean,
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
		openCommunity(targetType, targetId, onAdjustUnread) {
			this.$emit('open-community', targetType, targetId, onAdjustUnread);
		},
		playVideo(urlObject) {
			this.$emit('play-video', urlObject);
		},
	},
};
</script>

<style lang="scss">
@import '../_styles/_app.scss';
@import '../_styles/_breakpoints.scss';

.spec-view-page {
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
