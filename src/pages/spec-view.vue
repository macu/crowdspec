<template>
<div class="spec-view-page" :class="{loading}">
	<p v-if="loading">
		<loading-message message="Loading..."/>
	</p>
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
	<p v-else>
		Failed to load spec.
	</p>
</div>
</template>

<script>
import SpecView from '../spec/view.vue';
import LoadingMessage from '../widgets/loading.vue';

export default {
	components: {
		SpecView,
		LoadingMessage,
	},
	props: {
		loading: Boolean,
		spec: Object,
		enableEditing: Boolean,
	},
	emits: ['rendered', 'prop-nav-spec', 'open-community', 'play-video'],
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
		openCommunity(targetType, targetId, onAdjustUnread, onAdjustComments) {
			this.$emit('open-community', targetType, targetId, onAdjustUnread, onAdjustComments);
		},
		playVideo(urlObject) {
			this.$emit('play-video', urlObject);
		},
	},
};
</script>

<style lang="scss">
@use "sass:math";
@import '../_styles/_app.scss';
@import '../_styles/_breakpoints.scss';

.spec-view-page {
	padding: $content-area-padding;
	padding-top: math.div($content-area-padding, 2);

	@include mobile {
		padding: $content-area-padding-sm 0;

		&.loading {
			padding: $content-area-padding-sm;
		}
	}
}
</style>
